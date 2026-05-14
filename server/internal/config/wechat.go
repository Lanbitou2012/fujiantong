package config

import "os"

// ComponentConfig 微信开放平台第三方平台配置
type ComponentConfig struct {
	AppID          string
	AppSecret      string
	Token          string
	EncodingAESKey string
}

func GetComponentConfig() ComponentConfig {
	return ComponentConfig{
		AppID:          os.Getenv("WX_COMPONENT_APPID"),
		AppSecret:      os.Getenv("WX_COMPONENT_APPSECRET"),
		Token:          os.Getenv("WX_COMPONENT_TOKEN"),
		EncodingAESKey: os.Getenv("WX_COMPONENT_ENCODING_AES_KEY"),
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
		AppID:       os.Getenv("WECHAT_OPEN_APPID"),
		AppSecret:   os.Getenv("WECHAT_OPEN_APPSECRET"),
		RedirectURI: os.Getenv("WECHAT_OPEN_REDIRECT_URI"),
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
