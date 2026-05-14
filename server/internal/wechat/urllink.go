package wechat

// 小程序跳转能力 — URL Scheme / URL Link / 短链 / wxacode
// 文档入口：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/url-scheme/urlscheme.generate.html
//
// 使用场景：
//   - URL Scheme（weixin://...）：在 **非微信浏览器** 中点击可拉起小程序，仅 5 分钟（默认）或长期有效
//   - URL Link  （https://...）：HTTPS 链接，同样仅微信外有效，安卓/iOS 系统浏览器可解析

import "net/url"

// ─── wxa/generatescheme：生成 URL Scheme ───

type JumpWxaInfo struct {
	Path       string `json:"path,omitempty"`        // 通过 scheme 跳转到的页面
	Query      string `json:"query,omitempty"`       // 跳转参数
	EnvVersion string `json:"env_version,omitempty"` // release / trial / develop
}

type GenerateSchemeReq struct {
	JumpWxa        JumpWxaInfo `json:"jump_wxa"`
	IsExpire       bool        `json:"is_expire,omitempty"`
	ExpireType     int         `json:"expire_type,omitempty"`     // 0=失效时间 1=失效间隔天数
	ExpireTime     int64       `json:"expire_time,omitempty"`     // ExpireType=0 时使用，秒
	ExpireInterval int         `json:"expire_interval,omitempty"` // ExpireType=1 时使用，1-365 天
}

type GenerateSchemeResp struct {
	BaseErr
	OpenLink string `json:"openlink"` // weixin://dl/business/?t=xxxx
}

// GenerateScheme 调用 authorizer 的 wxa/generatescheme，需要 authorizer_access_token
// （即作者小程序自己的 access_token，因为 scheme 必须绑定具体小程序的 appid）
func (c *ComponentClient) GenerateScheme(authorizerAccessToken string, req GenerateSchemeReq) (*GenerateSchemeResp, error) {
	var out GenerateSchemeResp
	endpoint := wxAPIBase + "/wxa/generatescheme?access_token=" + url.QueryEscape(authorizerAccessToken)
	if err := c.postJSON(endpoint, req, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}

// ─── wxa/generate_urllink：生成 URL Link ───

type GenerateUrlLinkReq struct {
	Path           string `json:"path,omitempty"`
	Query          string `json:"query,omitempty"`
	IsExpire       bool   `json:"is_expire,omitempty"`
	ExpireType     int    `json:"expire_type,omitempty"`
	ExpireTime     int64  `json:"expire_time,omitempty"`
	ExpireInterval int    `json:"expire_interval,omitempty"`
	EnvVersion     string `json:"env_version,omitempty"`
}

type GenerateUrlLinkResp struct {
	BaseErr
	UrlLink string `json:"url_link"` // https://wxaurl.cn/xxxx
}

// GenerateUrlLink 生成 HTTPS URL Link（5 分钟 - 长期可配）
func (c *ComponentClient) GenerateUrlLink(authorizerAccessToken string, req GenerateUrlLinkReq) (*GenerateUrlLinkResp, error) {
	var out GenerateUrlLinkResp
	endpoint := wxAPIBase + "/wxa/generate_urllink?access_token=" + url.QueryEscape(authorizerAccessToken)
	if err := c.postJSON(endpoint, req, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}
