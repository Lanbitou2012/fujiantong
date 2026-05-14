package service

import (
	"errors"
	"fmt"

	"fujiantong/internal/config"
	"fujiantong/internal/wechat"
)

// LaunchInfo C 端 H5 拉起小程序所需的全套信息
type LaunchInfo struct {
	AppID   string `json:"appid"`        // 作者小程序 AppID
	Path    string `json:"path"`         // pages/file-detail/file-detail
	Query   string `json:"query"`        // id=xxx
	Scheme  string `json:"url_scheme"`   // weixin://dl/business/?t=xxx，可在微信外打开
	UrlLink string `json:"url_link"`     // https://wxaurl.cn/xxx，也是微信外用
}

// GenerateLaunchInfo 为某文件生成"H5 拉起小程序"所需的链接
// 调用方：C 端 H5 下载页（V2.8 §4.5 兜底路径主动改善：先尝试拉起小程序）
func (c *Container) GenerateLaunchInfo(downloadCode string) (*LaunchInfo, error) {
	f, err := c.GetFileByCode(downloadCode)
	if err != nil {
		return nil, fmt.Errorf("文件不存在: %w", err)
	}
	if f.AppID == "" {
		// 作者还没绑定小程序，无法拉起
		return nil, errors.New("该作者尚未绑定小程序")
	}

	cfg := config.GetComponentConfig()
	if cfg.AppID == "" || cfg.AppSecret == "" {
		return nil, errors.New("微信开放平台未配置")
	}

	// 调对应小程序的 authorizer_access_token
	authToken, err := c.EnsureAuthorizerToken(cfg.AppID, cfg.AppSecret, f.AppID)
	if err != nil {
		return nil, fmt.Errorf("获取作者 token 失败: %w", err)
	}

	info := &LaunchInfo{
		AppID: f.AppID,
		Path:  "pages/file-detail/file-detail",
		Query: fmt.Sprintf("id=%d", f.ID),
	}

	// 生成 URL Scheme（微信外 - 系统浏览器拉起）
	schemeReq := wechat.GenerateSchemeReq{
		JumpWxa: wechat.JumpWxaInfo{
			Path:       info.Path,
			Query:      info.Query,
			EnvVersion: "release",
		},
		IsExpire:       true,
		ExpireType:     1,
		ExpireInterval: 30, // 30 天有效
	}
	if resp, err := c.WxComponent.GenerateScheme(authToken, schemeReq); err == nil {
		info.Scheme = resp.OpenLink
	}

	// 生成 URL Link（HTTPS 形式，与 Scheme 同等用途）
	linkReq := wechat.GenerateUrlLinkReq{
		Path:           info.Path,
		Query:          info.Query,
		IsExpire:       true,
		ExpireType:     1,
		ExpireInterval: 30,
		EnvVersion:     "release",
	}
	if resp, err := c.WxComponent.GenerateUrlLink(authToken, linkReq); err == nil {
		info.UrlLink = resp.UrlLink
	}

	return info, nil
}
