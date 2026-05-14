package service

import (
	"errors"
	"fmt"
	"net/url"

	"fujiantong/internal/config"
)

// AuthURLInfo 作者授权链接
type AuthURLInfo struct {
	AuthURL     string `json:"auth_url"`     // 跳转到微信扫码授权页
	PreAuthCode string `json:"pre_auth_code"` // 调试用
	ExpiresIn   int    `json:"expires_in"`
}

// GenerateAuthURL 生成"作者授权小程序代运营"的微信扫码链接
// V2.8 §15.2.8：必须同时勾选「流量主代运营(135)」+「代码管理权限集」两组权限
func (c *Container) GenerateAuthURL(promoterID uint64, bizAppID string) (*AuthURLInfo, error) {
	wxCfg := config.GetComponentConfig()
	if wxCfg.AppID == "" {
		return nil, errors.New("微信开放平台未配置")
	}

	componentToken, err := c.EnsureComponentToken(wxCfg.AppID, wxCfg.AppSecret)
	if err != nil {
		return nil, fmt.Errorf("获取 component_token 失败: %w", err)
	}

	resp, err := c.WxComponent.CreatePreAuthCode(componentToken, wxCfg.AppID)
	if err != nil {
		return nil, fmt.Errorf("生成 pre_auth_code 失败: %w", err)
	}

	platform := config.GetPlatformConfig()
	redirectURI := fmt.Sprintf("%s/api/v1/wx/component/callback", platform.APIBaseURL)
	if promoterID > 0 {
		redirectURI += fmt.Sprintf("?promoter_id=%d", promoterID)
	}

	// V2.8 §15.2.8 关键前提：必须同时授权流量主代运营 + 代码管理两组权限集
	// 流量主代运营 = 135；代码管理权限集 = 17（典型，需后台确认）
	// 拼接：category_id_list=135,17
	q := url.Values{}
	q.Set("component_appid", wxCfg.AppID)
	q.Set("pre_auth_code", resp.PreAuthCode)
	q.Set("redirect_uri", redirectURI)
	q.Set("auth_type", "2") // 2 = 仅小程序
	q.Set("category_id_list", "135,17")
	if bizAppID != "" {
		q.Set("biz_appid", bizAppID)
	}

	authURL := "https://mp.weixin.qq.com/cgi-bin/componentloginpage?" + q.Encode()
	return &AuthURLInfo{
		AuthURL:     authURL,
		PreAuthCode: resp.PreAuthCode,
		ExpiresIn:   resp.ExpiresIn,
	}, nil
}
