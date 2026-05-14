package wechat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Code2SessionResult struct {
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type MiniProgramClient struct {
	AppID      string
	AppSecret  string
	HTTPClient *http.Client
}

func NewMiniProgramClient(appID string, appSecret string) *MiniProgramClient {
	return &MiniProgramClient{
		AppID:     appID,
		AppSecret: appSecret,
		HTTPClient: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

func (c *MiniProgramClient) Code2Session(code string) (*Code2SessionResult, error) {
	if c.AppID == "" || c.AppSecret == "" {
		return nil, errors.New("微信小程序 AppID 或 AppSecret 未配置")
	}
	if code == "" {
		return nil, errors.New("微信登录 code 不能为空")
	}

	endpoint := "https://api.weixin.qq.com/sns/jscode2session"
	query := url.Values{}
	query.Set("appid", c.AppID)
	query.Set("secret", c.AppSecret)
	query.Set("js_code", code)
	query.Set("grant_type", "authorization_code")

	resp, err := c.HTTPClient.Get(endpoint + "?" + query.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Code2SessionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信 code2session 失败: %d %s", result.ErrCode, result.ErrMsg)
	}
	if result.OpenID == "" {
		return nil, errors.New("微信 code2session 未返回 openid")
	}
	return &result, nil
}
