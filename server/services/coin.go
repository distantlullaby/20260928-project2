package services

import (
	"errors"
	"fmt"

	"taste-server/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrInsufficient 尝鲜币余额不足
var ErrInsufficient = errors.New("尝鲜币余额不足")

// bizError 业务错误，handler 层据此返回 4xx
type bizError struct{ msg string }

func (e *bizError) Error() string { return e.msg }

// BizError 构造业务错误
func BizError(format string, args ...interface{}) error {
	return &bizError{msg: fmt.Sprintf(format, args...)}
}

// IsBiz 判断是否业务错误
func IsBiz(err error) (*bizError, bool) {
	var b *bizError
	if errors.As(err, &b) {
		return b, true
	}
	return nil, false
}

// WriteCoinTx 在事务内锁定用户行、调整余额并写入一条尝鲜币流水。
// deltaAvail / deltaFrozen 分别表示可用余额、冻结余额的增量（可正可负）。
func WriteCoinTx(tx *gorm.DB, userID uint, deltaAvail, deltaFrozen int, kind, refType string, refID uint, remark string) (*models.User, error) {
	var user models.User
	// 行级锁，防止并发发帖/结算导致超支
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
		return nil, err
	}
	newAvail := user.CoinBalance + deltaAvail
	newFrozen := user.FrozenBalance + deltaFrozen
	if newAvail < 0 || newFrozen < 0 {
		return nil, ErrInsufficient
	}
	if err := tx.Model(&user).Updates(map[string]interface{}{
		"coin_balance":   newAvail,
		"frozen_balance": newFrozen,
	}).Error; err != nil {
		return nil, err
	}
	rec := models.CoinTransaction{
		UserID:       userID,
		Amount:       deltaAvail,
		BalanceAfter: newAvail,
		FrozenAfter:  newFrozen,
		Type:         kind,
		RefType:      refType,
		RefID:        refID,
		Remark:       remark,
	}
	if err := tx.Create(&rec).Error; err != nil {
		return nil, err
	}
	user.CoinBalance = newAvail
	user.FrozenBalance = newFrozen
	return &user, nil
}
