package model

import "time"

// SystemSetting 平台级动态配置 key-value 表
// 支持运行时热更新（微信开放平台、COS 等敏感凭据）
// 优先级：DB 设置 > 环境变量
type SystemSetting struct {
	Key       string    `gorm:"primaryKey;type:varchar(64)" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	Category  string    `gorm:"index;type:varchar(32);not null;default:''" json:"category"` // wechat / cos / platform
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy uint64    `gorm:"not null;default:0" json:"updated_by"`
}

func (SystemSetting) TableName() string { return "system_settings" }

// 设置项 key 常量
const (
	// 微信开放平台 - 第三方平台
	KeyWxComponentAppID      = "wx_component_appid"
	KeyWxComponentAppSecret  = "wx_component_appsecret"
	KeyWxComponentToken      = "wx_component_token"
	KeyWxComponentAESKey     = "wx_component_encoding_aes_key"

	// 微信开放平台 - 网站应用（扫码登录）
	KeyWechatOpenAppID    = "wechat_open_appid"
	KeyWechatOpenSecret   = "wechat_open_appsecret"
	KeyWechatOpenRedirect = "wechat_open_redirect_uri"

	// 腾讯云 COS
	KeyStorageDriver  = "storage_driver" // local / cos
	KeyCOSBucketURL   = "cos_bucket_url"
	KeyCOSSecretID    = "cos_secret_id"
	KeyCOSSecretKey   = "cos_secret_key"
	KeyCOSCDNBaseURL  = "cdn_base_url"
)
