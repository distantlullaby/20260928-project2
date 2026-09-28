package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ---------- 心愿状态 ----------
const (
	WishStatusOpen     = "open"     // 等待代吃
	WishStatusSettled  = "settled"  // 已采纳结算
	WishStatusCanceled = "canceled" // 已取消（悬赏退回）
)

// ---------- 尝鲜币流水类型 ----------
const (
	TxnRegisterGift  = "register_gift"  // 注册赠送
	TxnWishFreeze    = "wish_freeze"    // 发布心愿冻结悬赏
	TxnAppendFreeze  = "append_freeze"  // 追加悬赏冻结
	TxnSettleIncome  = "settle_income"  // 代吃被采纳，悬赏入账
	TxnSettleExpense = "settle_expense" // 采纳结算（冻结悬赏转出）
	TxnCancelRefund  = "cancel_refund"  // 取消心愿，悬赏退回
)

// JSONStringSlice 用于在 MySQL 中以 JSON 文本存储字符串数组（如多张实拍图）。
type JSONStringSlice []string

func (s JSONStringSlice) Value() (driver.Value, error) {
	if s == nil {
		s = JSONStringSlice{}
	}
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *JSONStringSlice) Scan(v interface{}) error {
	if v == nil {
		*s = JSONStringSlice{}
		return nil
	}
	var data []byte
	switch t := v.(type) {
	case []byte:
		data = t
	case string:
		data = []byte(t)
	default:
		return errors.New("unsupported type for JSONStringSlice")
	}
	if len(data) == 0 {
		*s = JSONStringSlice{}
		return nil
	}
	return json.Unmarshal(data, s)
}

// User 用户表。coins 为可用余额，frozen_coins 为被心愿悬赏冻结的币。
type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"size:32;uniqueIndex;not null" json:"username"`
	Password    string    `gorm:"size:128;not null" json:"-"`
	Nickname    string    `gorm:"size:32;not null" json:"nickname"`
	AvatarEmoji string    `gorm:"size:16;not null;default:'🍜'" json:"avatar_emoji"`
	Coins       int       `gorm:"not null;default:0" json:"coins"`
	FrozenCoins int       `gorm:"not null;default:0" json:"frozen_coins"`
	CreatedAt   time.Time `json:"created_at"`
}

// Wish 心愿表（求代吃需求）。
type Wish struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"index;not null" json:"user_id"`
	User           User      `json:"-"`
	RestaurantName string    `gorm:"size:64;not null" json:"restaurant_name"`
	Address        string    `gorm:"size:128;not null" json:"address"`
	DishList       string    `gorm:"size:255;not null" json:"dish_list"` // 心愿菜单
	Reason         string    `gorm:"size:512;not null" json:"reason"`    // 种草理由
	RewardCoins    int       `gorm:"not null" json:"reward_coins"`       // 当前悬赏总额（含追加）
	FrozenCoins    int       `gorm:"not null" json:"frozen_coins"`       // 该心愿当前仍冻结的币
	Status         string    `gorm:"size:16;index;not null;default:'open'" json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TasteResponse 品尝回应表（替 TA 去吃的现场记录）。
type TasteResponse struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	WishID    uint            `gorm:"index;not null" json:"wish_id"`
	UserID    uint            `gorm:"index;not null" json:"user_id"` // 代吃者
	User      User            `json:"-"`
	Photos    JSONStringSlice `gorm:"type:text" json:"photos"`           // 现场实拍
	Comment   string          `gorm:"size:1024;not null" json:"comment"` // 口味点评
	Rating    int             `gorm:"not null;default:5" json:"rating"`  // 1~5 星
	IsAdopted bool            `gorm:"not null;default:false" json:"is_adopted"`
	CreatedAt time.Time       `json:"created_at"`
}

// CoinTransaction 尝鲜币流水表。amount 带符号：正为入账，负为出账。
// balance_after 记录该笔变动后的可用余额（冻结变动不体现在可用余额上）。
type CoinTransaction struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	Amount       int       `gorm:"not null" json:"amount"`
	BalanceAfter int       `gorm:"not null" json:"balance_after"`
	Type         string    `gorm:"size:24;index;not null" json:"type"`
	Remark       string    `gorm:"size:255;not null;default:''" json:"remark"`
	WishID       *uint     `gorm:"index" json:"wish_id"`
	CreatedAt    time.Time `json:"created_at"`
}
