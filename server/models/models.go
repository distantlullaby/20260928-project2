package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// 心愿状态
const (
	WishStatusOpen     = "open"     // 求代吃中
	WishStatusSettled  = "settled"  // 已采纳，悬赏已结算
	WishStatusCanceled = "canceled" // 已取消，悬赏已退回
)

// 尝鲜币流水类型
const (
	CoinTypeRegister     = "register"      // 注册赠送
	CoinTypeFreeze       = "freeze"        // 发布/追加悬赏冻结（余额 -> 冻结）
	CoinTypeUnfreeze     = "unfreeze"      // 取消心愿，冻结退回余额
	CoinTypeSettleSpend  = "settle_spend"  // 采纳后悬赏支出（冻结扣除）
	CoinTypeSettleIncome = "settle_income" // 代吃悬赏收入
)

// User 用户表
type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Username      string    `gorm:"size:32;uniqueIndex;not null" json:"username"`
	PasswordHash  string    `gorm:"size:128;not null" json:"-"`
	Nickname      string    `gorm:"size:32" json:"nickname"`
	Avatar        string    `gorm:"size:512;default:''" json:"avatar"`
	CoinBalance   int       `gorm:"not null;default:0;comment:可用尝鲜币" json:"coin_balance"`
	FrozenBalance int       `gorm:"not null;default:0;comment:冻结中的悬赏币" json:"frozen_balance"`
	CreatedAt     time.Time `json:"created_at"`
}

// Wish 心愿表（求代吃需求）
type Wish struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	PosterID           uint           `gorm:"index;not null" json:"poster_id"`
	Restaurant         string         `gorm:"size:128;not null;comment:心仪餐厅/摊位" json:"restaurant"`
	Dish               string         `gorm:"size:128;not null;comment:心愿菜单" json:"dish"`
	Reason             string         `gorm:"size:1024;not null;comment:种草理由" json:"reason"`
	Address            string         `gorm:"size:256;default:''" json:"address"`
	Bounty             int            `gorm:"not null;default:0;comment:当前悬赏尝鲜币(冻结中)" json:"bounty"`
	Status             string         `gorm:"size:16;not null;default:'open';index" json:"status"`
	AcceptedResponseID *uint          `gorm:"comment:被采纳的品尝回应" json:"accepted_response_id,omitempty"`
	CreatedAt          time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	// 聚合字段，不入库
	Poster   *User             `gorm:"-" json:"poster,omitempty"`
	Response *TastingResponse  `gorm:"-" json:"accepted_response,omitempty"`
	RespCount int              `gorm:"-" json:"response_count"`
	Mine      bool             `gorm:"-" json:"mine"`
}

// TastingResponse 品尝回应表（双面明信片的背面）
type TastingResponse struct {
	ID        uint     `gorm:"primaryKey" json:"id"`
	WishID    uint     `gorm:"index;not null" json:"wish_id"`
	TasterID  uint     `gorm:"index;not null" json:"taster_id"`
	Photos    Strings  `gorm:"type:json;comment:现场实拍URL列表" json:"photos"`
	Review    string   `gorm:"size:2048;not null;comment:口味点评" json:"review"`
	Rating    int      `gorm:"not null;default:5;comment:口味打分1-5" json:"rating"`
	CreatedAt time.Time `json:"created_at"`
	// 聚合字段
	Taster *User `gorm:"-" json:"taster,omitempty"`
	Wish   *Wish `gorm:"-" json:"wish,omitempty"`
	Mine   bool  `gorm:"-" json:"mine"`
}

// CoinTransaction 尝鲜币流水表
type CoinTransaction struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	Amount       int       `gorm:"not null;comment:变动金额(正为收入,负为支出)" json:"amount"`
	BalanceAfter int       `gorm:"not null;comment:变动后可用余额" json:"balance_after"`
	FrozenAfter  int       `gorm:"not null;comment:变动后冻结余额" json:"frozen_after"`
	Type         string    `gorm:"size:32;not null;index" json:"type"`
	RefType      string    `gorm:"size:32;default:''" json:"ref_type"`
	RefID        uint      `gorm:"default:0" json:"ref_id"`
	Remark       string    `gorm:"size:256;default:''" json:"remark"`
	CreatedAt    time.Time `json:"index" json:"created_at"`
}

// Strings 是存入 JSON 列的字符串切片
type Strings []string

func (s Strings) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

func (s *Strings) Scan(src interface{}) error {
	if src == nil {
		*s = Strings{}
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("unsupported type for Strings")
	}
	if len(data) == 0 {
		*s = Strings{}
		return nil
	}
	return json.Unmarshal(data, s)
}
