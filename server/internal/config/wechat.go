package config

import (
	"fujiantong/internal/model"
	"os"
)

// SettingsGetter 由 service 层注入：从 DB（system_settings）读取动态配置。
// 优先级：DB 配置 > 环境变量。
// 在 main 启动时调用 SetSettingsGetter() 注册。
type SettingsGetter func(key string) string

var settingsGetter SettingsGetter

func SetSettingsGetter(g SettingsGetter) { settingsGetter = g }

// fromDBOrEnv 先读 DB，空则回退环境变量
func fromDBOrEnv(dbKey, envKey string) string {
	if settingsGetter != nil {
		if v := settingsGetter(dbKey); v != "" {
			return v
		}
	}
	return os.Getenv(envKey)
}

// ComponentConfig 微信开放平台第三方平台配置
type ComponentConfig struct {
	AppID          string
	AppSecret      string
	Token          string
	EncodingAESKey string
}

func GetComponentConfig() ComponentConfig {
	return ComponentConfig{
		AppID:          fromDBOrEnv(model.KeyWxComponentAppID, "WX_COMPONENT_APPID"),
		AppSecret:      fromDBOrEnv(model.KeyWxComponentAppSecret, "WX_COMPONENT_APPSECRET"),
		Token:          fromDBOrEnv(model.KeyWxComponentToken, "WX_COMPONENT_TOKEN"),
		EncodingAESKey: fromDBOrEnv(model.KeyWxComponentAESKey, "WX_COMPONENT_ENCODING_AES_KEY"),
	}
}

// OpenPlatformConfig 开放平台「网站应用」(扫码登录) 配置
type OpenPlatformConfig struct {
	AppID       string
	AppSecret   string
	RedirectURI string
}

func GetOpenPlatformConfig() OpenPlatformConfig {
	return OpenPlatformConfig{
		AppID:       fromDBOrEnv(model.KeyWechatOpenAppID, "WECHAT_OPEN_APPID"),
		AppSecret:   fromDBOrEnv(model.KeyWechatOpenSecret, "WECHAT_OPEN_APPSECRET"),
		RedirectURI: fromDBOrEnv(model.KeyWechatOpenRedirect, "WECHAT_OPEN_REDIRECT_URI"),
	}
}

// COSConfig 腾讯云 COS 对象存储配置
type COSConfig struct {
	Driver     string // local / cos
	BucketURL  string
	SecretID   string
	SecretKey  string
	CDNBaseURL string
}

func GetCOSConfig() COSConfig {
	driver := fromDBOrEnv(model.KeyStorageDriver, "STORAGE_DRIVER")
	if driver == "" {
		driver = "local"
	}
	return COSConfig{
		Driver:     driver,
		BucketURL:  fromDBOrEnv(model.KeyCOSBucketURL, "COS_BUCKET_URL"),
		SecretID:   fromDBOrEnv(model.KeyCOSSecretID, "COS_SECRET_ID"),
		SecretKey:  fromDBOrEnv(model.KeyCOSSecretKey, "COS_SECRET_KEY"),
		CDNBaseURL: fromDBOrEnv(model.KeyCOSCDNBaseURL, "CDN_BASE_URL"),
	}
}

// PlatformConfig 附件通平台自身配置
type PlatformConfig struct {
	Domain     string
	APIBaseURL string
}

func GetPlatformConfig() PlatformConfig {
	domain := os.Getenv("PLATFORM_DOMAIN")
	if domain == "" {
		domain = "fujian.5g6g.top"
	}
	apiBase := os.Getenv("API_BASE_URL")
	if apiBase == "" {
		apiBase = "https://" + domain
	}
	return PlatformConfig{
		Domain:     domain,
		APIBaseURL: apiBase,
	}
}
