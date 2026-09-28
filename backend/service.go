package main

import (
	"errors"

	"gorm.io/gorm"
)

// 业务哨兵错误
var (
	errNotFound      = errors.New("记录不存在")
	errInsufficient  = errors.New("尝鲜币余额不足")
	errInvalidAmount = errors.New("金额必须大于 0")
	errConflict      = errors.New("当前状态不允许该操作")
)

// register 注册并在同一事务内赠送初始尝鲜币，保证"注册即赠送"原子完成。
func register(username, passwordHash, nickname, emoji string) (*User, error) {
	user := &User{
		Username:    username,
		Password:    passwordHash,
		Nickname:    nickname,
		AvatarEmoji: emoji,
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		user.Coins = initialCoins
		if err := tx.Model(user).Update("coins", initialCoins).Error; err != nil {
			return err
		}
		ledger := CoinTransaction{
			UserID:       user.ID,
			Amount:       initialCoins,
			BalanceAfter: initialCoins,
			Type:         TxnRegisterGift,
			Remark:       "注册赠送初始尝鲜币",
		}
		return tx.Create(&ledger).Error
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

// createWish 发布求代吃：事务内锁定用户行，校验余额，可用币 -> 冻结币，写流水。
func createWish(userID uint, restaurant, address, dishes, reason string, reward int) (*Wish, error) {
	if reward <= 0 {
		return nil, errInvalidAmount
	}
	wish := &Wish{
		UserID:         userID,
		RestaurantName: restaurant,
		Address:        address,
		DishList:       dishes,
		Reason:         reason,
		RewardCoins:    reward,
		FrozenCoins:    reward,
		Status:         WishStatusOpen,
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		// SELECT ... FOR UPDATE：串行化同一用户的资金操作，避免并发超扣
		if err := tx.Clauses(clauseForUpdate()).First(&user, userID).Error; err != nil {
			return err
		}
		if user.Coins < reward {
			return errInsufficient
		}
		user.Coins -= reward
		user.FrozenCoins += reward
		if err := tx.Model(&user).Updates(map[string]interface{}{
			"coins":        user.Coins,
			"frozen_coins": user.FrozenCoins,
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(wish).Error; err != nil {
			return err
		}
		ledger := CoinTransaction{
			UserID:       userID,
			Amount:       -reward,
			BalanceAfter: user.Coins,
			Type:         TxnWishFreeze,
			Remark:       "发布求代吃，冻结悬赏：" + restaurant,
			WishID:       &wish.ID,
		}
		return tx.Create(&ledger).Error
	})
	if err != nil {
		return nil, err
	}
	return wish, nil
}

// appendReward 追加悬赏：同 createWish 的冻结逻辑。
func appendReward(wishID, userID uint, amount int) (*Wish, error) {
	if amount <= 0 {
		return nil, errInvalidAmount
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		var wish Wish
		if err := tx.Clauses(clauseForUpdate()).First(&wish, wishID).Error; err != nil {
			return err
		}
		if wish.UserID != userID {
			return errNotFound
		}
		if wish.Status != WishStatusOpen {
			return errConflict
		}
		var user User
		if err := tx.Clauses(clauseForUpdate()).First(&user, userID).Error; err != nil {
			return err
		}
		if user.Coins < amount {
			return errInsufficient
		}
		user.Coins -= amount
		user.FrozenCoins += amount
		if err := tx.Model(&user).Updates(map[string]interface{}{
			"coins":        user.Coins,
			"frozen_coins": user.FrozenCoins,
		}).Error; err != nil {
			return err
		}
		wish.RewardCoins += amount
		wish.FrozenCoins += amount
		if err := tx.Model(&wish).Updates(map[string]interface{}{
			"reward_coins": wish.RewardCoins,
			"frozen_coins": wish.FrozenCoins,
		}).Error; err != nil {
			return err
		}
		ledger := CoinTransaction{
			UserID:       userID,
			Amount:       -amount,
			BalanceAfter: user.Coins,
			Type:         TxnAppendFreeze,
			Remark:       "追加悬赏：" + wish.RestaurantName,
			WishID:       &wishID,
		}
		return tx.Create(&ledger).Error
	})
	if err != nil {
		return nil, err
	}
	var updated Wish
	DB.First(&updated, wishID)
	return &updated, err
}

// submitResponse 代吃者提交"现场实拍 + 口味点评"。
func submitResponse(wishID, tasterID uint, photos []string, comment string, rating int) (*TasteResponse, error) {
	if rating < 1 || rating > 5 {
		rating = 5
	}
	var wish Wish
	if err := DB.First(&wish, wishID).Error; err != nil {
		return nil, errNotFound
	}
	if wish.Status != WishStatusOpen {
		return nil, errConflict
	}
	if wish.UserID == tasterID {
		return nil, errors.New("不能替自己的心愿代吃哦")
	}
	var count int64
	if err := DB.Model(&TasteResponse{}).
		Where("wish_id = ? AND user_id = ?", wishID, tasterID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("你已经替 TA 吃过这家啦，换个心愿试试")
	}
	resp := &TasteResponse{
		WishID:  wishID,
		UserID:  tasterID,
		Photos:  photos,
		Comment: comment,
		Rating:  rating,
	}
	if err := DB.Create(resp).Error; err != nil {
		return nil, err
	}
	return resp, nil
}

// settleWish 发起人采纳某条代吃回应并结算。
// 事务内锁定心愿与两名用户行，将悬赏从冻结池转入代吃者可用余额，
// 其余回应标记未采纳，双方各记一条流水。
func settleWish(wishID, ownerID, responseID uint) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var wish Wish
		if err := tx.Clauses(clauseForUpdate()).First(&wish, wishID).Error; err != nil {
			return errNotFound
		}
		if wish.UserID != ownerID {
			return errNotFound
		}
		if wish.Status != WishStatusOpen || wish.FrozenCoins <= 0 {
			return errConflict
		}
		var resp TasteResponse
		if err := tx.Clauses(clauseForUpdate()).First(&resp, responseID).Error; err != nil {
			return errNotFound
		}
		if resp.WishID != wishID || resp.IsAdopted {
			return errConflict
		}

		reward := wish.FrozenCoins

		// 发起人：冻结币扣减（可用币在冻结时已扣，此处不再变动）
		var owner User
		if err := tx.Clauses(clauseForUpdate()).First(&owner, ownerID).Error; err != nil {
			return err
		}
		if owner.FrozenCoins < reward {
			return errors.New("冻结账目异常")
		}
		owner.FrozenCoins -= reward
		if err := tx.Model(&owner).Update("frozen_coins", owner.FrozenCoins).Error; err != nil {
			return err
		}

		// 代吃者：悬赏入账
		var taster User
		if err := tx.Clauses(clauseForUpdate()).First(&taster, resp.UserID).Error; err != nil {
			return err
		}
		taster.Coins += reward
		if err := tx.Model(&taster).Update("coins", taster.Coins).Error; err != nil {
			return err
		}

		// 心愿状态与冻结清零，中选回应标记采纳
		if err := tx.Model(&wish).Updates(map[string]interface{}{
			"status":       WishStatusSettled,
			"frozen_coins": 0,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&resp).Update("is_adopted", true).Error; err != nil {
			return err
		}

		// 双方流水（同一条事务内提交）。
		// 发起人的可用币在冻结时已扣减，结算只动冻结池，故此处记金额 0 的
		// "结算完成"信息流水，保证"流水金额之和 == 可用余额"恒等式不被破坏。
		ownerLedger := CoinTransaction{
			UserID:       ownerID,
			Amount:       0,
			BalanceAfter: owner.Coins,
			Type:         TxnSettleExpense,
			Remark:       "采纳代吃，悬赏已结算给代吃者：" + wish.RestaurantName,
			WishID:       &wishID,
		}
		tasterLedger := CoinTransaction{
			UserID:       taster.ID,
			Amount:       reward,
			BalanceAfter: taster.Coins,
			Type:         TxnSettleIncome,
			Remark:       "代吃被采纳，悬赏入账：" + wish.RestaurantName,
			WishID:       &wishID,
		}
		return tx.Create([]CoinTransaction{ownerLedger, tasterLedger}).Error
	})
}

// cancelWish 发起人取消心愿，仍冻结的悬赏退回可用余额。
func cancelWish(wishID, ownerID uint) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var wish Wish
		if err := tx.Clauses(clauseForUpdate()).First(&wish, wishID).Error; err != nil {
			return errNotFound
		}
		if wish.UserID != ownerID {
			return errNotFound
		}
		if wish.Status != WishStatusOpen {
			return errConflict
		}
		refund := wish.FrozenCoins
		if refund <= 0 {
			return errConflict
		}

		var owner User
		if err := tx.Clauses(clauseForUpdate()).First(&owner, ownerID).Error; err != nil {
			return err
		}
		owner.FrozenCoins -= refund
		owner.Coins += refund
		if err := tx.Model(&owner).Updates(map[string]interface{}{
			"coins":        owner.Coins,
			"frozen_coins": owner.FrozenCoins,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&wish).Updates(map[string]interface{}{
			"status":       WishStatusCanceled,
			"frozen_coins": 0,
		}).Error; err != nil {
			return err
		}
		ledger := CoinTransaction{
			UserID:       ownerID,
			Amount:       refund,
			BalanceAfter: owner.Coins,
			Type:         TxnCancelRefund,
			Remark:       "取消求代吃，悬赏退回：" + wish.RestaurantName,
			WishID:       &wishID,
		}
		return tx.Create(&ledger).Error
	})
}
