package model

import "time"

// WxToken component_access_token 与 authorizer_access_token 中央存储。
// component_access_token：key = "component"，只有一条。
// authorizer_access_token：key = 作者小程序 appid，每个作者一条。
type WxToken struct {
	Key          string    `gorm:"primaryKey;type:varchar(64)" json:"key"`
	AccessToken  string    `gorm:"type:varchar(512);not null;default:''" json:"-"`
	RefreshToken string    `gorm:"type:varchar(512);not null;default:''" json:"-"`
	ExpiresAt    time.Time `json:"expires_at"`
	Extra        string    `gorm:"type:text" json:"extra"` // 存 verify_ticket 等
	UpdatedAt    time.Time `json:"updated_at"`
}

func (WxToken) TableName() string { return "wx_tokens" }
