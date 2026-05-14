package wechat

// 微信开放平台「网站应用扫码登录」OAuth 封装
// 文档：https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const openOAuthBase = "https://api.weixin.qq.com"

type OpenOAuthClient struct {
	AppID      string
	AppSecret  string
	HTTPClient *http.Client
}

func NewOpenOAuthClient(appID, secret string) *OpenOAuthClient {
	return &OpenOAuthClient{
		AppID:      appID,
		AppSecret:  secret,
		HTTPClient: &http.Client{Timeout: 8 * time.Second},
	}
}

// BuildLoginURL 返回扫码登录二维码页 URL
func (c *OpenOAuthClient) BuildLoginURL(redirectURI, state string) string {
	q := url.Values{}
	q.Set("appid", c.AppID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "snsapi_login")
	q.Set("state", state)
	return "https://open.weixin.qq.com/connect/qrconnect?" + q.Encode() + "#wechat_redirect"
}

type OpenAccessToken struct {
	BaseErr
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	UnionID      string `json:"unionid"`
	Scope        string `json:"scope"`
}

// Code2Token 用授权 code 换 access_token + openid + unionid
func (c *OpenOAuthClient) Code2Token(code string) (*OpenAccessToken, error) {
	if c.AppID == "" || c.AppSecret == "" {
		return nil, errors.New("未配置开放平台 AppID/Secret")
	}
	q := url.Values{}
	q.Set("appid", c.AppID)
	q.Set("secret", c.AppSecret)
	q.Set("code", code)
	q.Set("grant_type", "authorization_code")
	resp, err := c.HTTPClient.Get(openOAuthBase + "/sns/oauth2/access_token?" + q.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out OpenAccessToken
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.ErrCode != 0 {
		return nil, fmt.Errorf("oauth2 access_token error: %d %s", out.ErrCode, out.ErrMsg)
	}
	return &out, nil
}

type OpenUserInfo struct {
	BaseErr
	OpenID     string   `json:"openid"`
	UnionID    string   `json:"unionid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
}

// UserInfo 拉取扫码用户资料
func (c *OpenOAuthClient) UserInfo(accessToken, openID string) (*OpenUserInfo, error) {
	q := url.Values{}
	q.Set("access_token", accessToken)
	q.Set("openid", openID)
	resp, err := c.HTTPClient.Get(openOAuthBase + "/sns/userinfo?" + q.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out OpenUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.ErrCode != 0 {
		return nil, fmt.Errorf("userinfo error: %d %s", out.ErrCode, out.ErrMsg)
	}
	return &out, nil
}
