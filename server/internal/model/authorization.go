package model

import "time"

// Authorization 作者小程序授权记录。
// 一个作者绑定一个小程序 appid，授权后平台可代运营其流量主 + 代码管理。
type Authorization struct {
	ID     uint64 `gorm:"primaryKey" json:"id"`
	UserID uint64 `gorm:"index;not null" json:"user_id"`
	AppID  string `gorm:"uniqueIndex;type:varchar(64);not null" json:"appid"`

	// 授权信息
	AuthorizerRefreshToken string     `gorm:"type:varchar(512);not null;default:''" json:"-"`
	ShareRatio             int        `gorm:"not null;default:28" json:"share_ratio"` // 平台占比（默认 28）
	ShareRatioConfirmedAt  *time.Time `json:"share_ratio_confirmed_at"`               // 作者确认协议时间

	// 小程序详情（从 getAuthorizerInfo 同步）
	MpNickname    string `gorm:"type:varchar(100);not null;default:''" json:"mp_nickname"`
	MpHeadImg     string `gorm:"type:varchar(512);not null;default:''" json:"mp_head_img"`
	PrincipalName string `gorm:"type:varchar(100);not null;default:''" json:"principal_name"`

	// 流量主状态
	PublisherStatus string `gorm:"type:varchar(32);not null;default:'inactive'" json:"publisher_status"` // inactive / active

	// 授权权限集（逗号分隔的权限集 ID，如 "17,135"）
	GrantedPermissionIDs string `gorm:"type:varchar(255);not null;default:''" json:"granted_permission_ids"`

	// 状态：authorized / deauthorized
	Status string `gorm:"type:varchar(32);not null;default:'authorized'" json:"status"`

	AuthorizedAt   *time.Time `json:"authorized_at"`
	DeauthorizedAt *time.Time `json:"deauthorized_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (Authorization) TableName() string { return "authorizations" }
