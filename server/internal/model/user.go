package model

import "time"

// User B 端用户统一表。
// 角色模型：基础角色（作者） + 叠加权限（is_promoter / is_admin）。
// C 端粉丝不在此表，仅通过小程序广告统计 / UV 数据体现。
type User struct {
	ID                      uint64     `gorm:"primaryKey" json:"id"`
	WxUnionID               string     `gorm:"uniqueIndex;type:varchar(64);not null;default:''" json:"wx_unionid"`
	WxOpenIDWeb             string     `gorm:"index;type:varchar(64);not null;default:''" json:"wx_openid_web"`
	Nickname                string     `gorm:"type:varchar(100);not null;default:''" json:"nickname"`
	AvatarURL               string     `gorm:"type:varchar(512);not null;default:''" json:"avatar_url"`
	Phone                   string     `gorm:"type:varchar(20);not null;default:''" json:"phone"`

	// 角色叠加位
	IsPromoter bool `gorm:"not null;default:false;index" json:"is_promoter"`
	IsAdmin    bool `gorm:"not null;default:false;index" json:"is_admin"`

	// 推广员绑定（铁律 2：单层，绝对禁止多层）
	// CHECK (id <> parent_promoter_user_id) 在 DDL 层约束
	ParentPromoterUserID uint64 `gorm:"index;not null;default:0" json:"parent_promoter_user_id"`

	// 作者绑定的小程序
	BoundAppID              string `gorm:"uniqueIndex;type:varchar(64);not null;default:''" json:"bound_appid"`
	AuthorizerRefreshToken  string `gorm:"type:varchar(512);not null;default:''" json:"-"`

	// 管理员登录（仅 is_admin=true 时使用）
	AdminUsername       string `gorm:"index;type:varchar(64);not null;default:''" json:"admin_username"`
	LoginPassphraseHash string `gorm:"type:varchar(255);not null;default:''" json:"-"`

	// 收款信息（推广员用）
	PaymentInfo string `gorm:"type:text" json:"payment_info"`

	// 状态：1=正常 0=封禁
	Status int8 `gorm:"not null;default:1;index" json:"status"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at"`
}

func (User) TableName() string { return "users" }

// UserRoleHistory 角色位变更历史（铁律 7）。
// 支撑铁律 1：按广告曝光日逐日归属时，需查历史表判定「某日该用户是否是推广员」。
type UserRoleHistory struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	UserID         uint64    `gorm:"index;not null" json:"user_id"`
	Field          string    `gorm:"type:varchar(32);not null" json:"field"`           // is_promoter / is_admin
	OldValue       string    `gorm:"type:varchar(16);not null;default:''" json:"old_value"` // "true" / "false"
	NewValue       string    `gorm:"type:varchar(16);not null;default:''" json:"new_value"`
	ChangedByAdmin uint64    `gorm:"not null;default:0" json:"changed_by_admin"`
	Reason         string    `gorm:"type:varchar(512);not null;default:''" json:"reason"`
	CreatedAt      time.Time `json:"created_at"`
}

func (UserRoleHistory) TableName() string { return "user_role_history" }
