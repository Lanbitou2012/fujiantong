package wechat

// 微信开放平台第三方平台 OpenAPI 封装
// 文档入口：https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/Third_party_platform_appid.html

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const wxAPIBase = "https://api.weixin.qq.com"

type ComponentClient struct {
	HTTPClient *http.Client
}

func NewComponentClient() *ComponentClient {
	return &ComponentClient{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type BaseErr struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (e BaseErr) asError() error {
	if e.ErrCode == 0 {
		return nil
	}
	return fmt.Errorf("wechat api error: %d %s", e.ErrCode, e.ErrMsg)
}

type ComponentAccessTokenResp struct {
	BaseErr
	ComponentAccessToken string `json:"component_access_token"`
	ExpiresIn            int    `json:"expires_in"`
}

// GetComponentAccessToken 通过 verify_ticket 换取 component_access_token（2 小时有效）
func (c *ComponentClient) GetComponentAccessToken(appID, appSecret, verifyTicket string) (*ComponentAccessTokenResp, error) {
	body := map[string]string{
		"component_appid":         appID,
		"component_appsecret":     appSecret,
		"component_verify_ticket": verifyTicket,
	}
	var out ComponentAccessTokenResp
	if err := c.postJSON(wxAPIBase+"/cgi-bin/component/api_component_token", body, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}

type PreAuthCodeResp struct {
	BaseErr
	PreAuthCode string `json:"pre_auth_code"`
	ExpiresIn   int    `json:"expires_in"`
}

// CreatePreAuthCode 生成 pre_auth_code，用于扫码授权页面
func (c *ComponentClient) CreatePreAuthCode(componentAccessToken, appID string) (*PreAuthCodeResp, error) {
	body := map[string]string{"component_appid": appID}
	var out PreAuthCodeResp
	endpoint := wxAPIBase + "/cgi-bin/component/api_create_preauthcode?component_access_token=" + url.QueryEscape(componentAccessToken)
	if err := c.postJSON(endpoint, body, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}

type FuncScopeItem struct {
	FuncScopeCategory struct {
		ID int `json:"id"`
	} `json:"funcscope_category"`
}

type AuthorizationInfo struct {
	AuthorizerAppID        string          `json:"authorizer_appid"`
	AuthorizerAccessToken  string          `json:"authorizer_access_token"`
	ExpiresIn              int             `json:"expires_in"`
	AuthorizerRefreshToken string          `json:"authorizer_refresh_token"`
	FuncInfo               []FuncScopeItem `json:"func_info"`
}

type QueryAuthResp struct {
	BaseErr
	AuthorizationInfo AuthorizationInfo `json:"authorization_info"`
}

// QueryAuth 扫码授权成功后，用 authorization_code 换取 authorizer_refresh_token
func (c *ComponentClient) QueryAuth(componentAccessToken, componentAppID, authorizationCode string) (*QueryAuthResp, error) {
	body := map[string]string{
		"component_appid":    componentAppID,
		"authorization_code": authorizationCode,
	}
	var out QueryAuthResp
	endpoint := wxAPIBase + "/cgi-bin/component/api_query_auth?component_access_token=" + url.QueryEscape(componentAccessToken)
	if err := c.postJSON(endpoint, body, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}

type AuthorizerTokenResp struct {
	BaseErr
	AuthorizerAccessToken  string `json:"authorizer_access_token"`
	ExpiresIn              int    `json:"expires_in"`
	AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
}

// RefreshAuthorizerToken 用 authorizer_refresh_token 刷新 access_token
func (c *ComponentClient) RefreshAuthorizerToken(componentAccessToken, componentAppID, authorizerAppID, refreshToken string) (*AuthorizerTokenResp, error) {
	body := map[string]string{
		"component_appid":          componentAppID,
		"authorizer_appid":         authorizerAppID,
		"authorizer_refresh_token": refreshToken,
	}
	var out AuthorizerTokenResp
	endpoint := wxAPIBase + "/cgi-bin/component/api_authorizer_token?component_access_token=" + url.QueryEscape(componentAccessToken)
	if err := c.postJSON(endpoint, body, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}

type AuthorizerInfoResp struct {
	BaseErr
	AuthorizerInfo struct {
		NickName      string `json:"nick_name"`
		HeadImg       string `json:"head_img"`
		PrincipalName string `json:"principal_name"`
		UserName      string `json:"user_name"`
		Alias         string `json:"alias"`
		Signature     string `json:"signature"`
		QrcodeUrl     string `json:"qrcode_url"`
	} `json:"authorizer_info"`
	AuthorizationInfo AuthorizationInfo `json:"authorization_info"`
}

// GetAuthorizerInfo 获取授权方详细资料（昵称、主体等）
func (c *ComponentClient) GetAuthorizerInfo(componentAccessToken, componentAppID, authorizerAppID string) (*AuthorizerInfoResp, error) {
	body := map[string]string{
		"component_appid":  componentAppID,
		"authorizer_appid": authorizerAppID,
	}
	var out AuthorizerInfoResp
	endpoint := wxAPIBase + "/cgi-bin/component/api_get_authorizer_info?component_access_token=" + url.QueryEscape(componentAccessToken)
	if err := c.postJSON(endpoint, body, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}

// MsgSecCheck 文本内容合规检测 (v2)
// 文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/sec-check/security.msgSecCheck.html
type MsgSecCheckResp struct {
	BaseErr
	Result struct {
		Suggest string `json:"suggest"`
		Label   int    `json:"label"`
	} `json:"result"`
	Detail  []map[string]any `json:"detail"`
	TraceID string           `json:"trace_id"`
}

func (c *ComponentClient) MsgSecCheck(authorizerAccessToken, openID, content, scene string) (*MsgSecCheckResp, error) {
	if scene == "" {
		scene = "2"
	}
	body := map[string]any{
		"version": 2,
		"openid":  openID,
		"scene":   scene,
		"content": content,
	}
	var out MsgSecCheckResp
	endpoint := wxAPIBase + "/wxa/msg_sec_check?access_token=" + url.QueryEscape(authorizerAccessToken)
	if err := c.postJSON(endpoint, body, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}

type MediaCheckAsyncResp struct {
	BaseErr
	TraceID string `json:"trace_id"`
}

// MediaCheckAsync 异步媒体机审
// mediaType: 1 音频 / 2 图片；其他媒体先转图片或抽帧后再调
func (c *ComponentClient) MediaCheckAsync(authorizerAccessToken, mediaURL string, mediaType int, openID string) (*MediaCheckAsyncResp, error) {
	body := map[string]any{
		"media_url":  mediaURL,
		"media_type": mediaType,
		"version":    2,
		"openid":     openID,
		"scene":      1,
	}
	var out MediaCheckAsyncResp
	endpoint := wxAPIBase + "/wxa/media_check_async?access_token=" + url.QueryEscape(authorizerAccessToken)
	if err := c.postJSON(endpoint, body, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}

// PublisherStat 流量主广告数据对账
// https://developers.weixin.qq.com/miniprogram/dev/framework/adManagement/advertisement-data-interface.html
type PublisherStatResp struct {
	BaseErr
	List []struct {
		Date          string  `json:"date"`
		SlotID        int64   `json:"slot_id"`
		AdSlot        string  `json:"ad_slot"`
		Income        float64 `json:"income"` // 单位：分
		ExposureCount int64   `json:"exposure_count"`
		ClickCount    int64   `json:"click_count"`
		ECPM          float64 `json:"ecpm"`
	} `json:"list"`
	Summary struct {
		Income float64 `json:"income"`
	} `json:"summary"`
	TotalNum int `json:"total_num"`
}

// GetPublisherStat 拉取流量主广告收益统计
func (c *ComponentClient) GetPublisherStat(authorizerAccessToken, action string, startDate, endDate string) (*PublisherStatResp, error) {
	if action == "" {
		action = "publisher_adpos_general"
	}
	q := url.Values{}
	q.Set("access_token", authorizerAccessToken)
	q.Set("action", action)
	q.Set("start_date", startDate)
	q.Set("end_date", endDate)
	q.Set("page", "1")
	q.Set("page_size", "50")
	endpoint := wxAPIBase + "/publisher/stat?" + q.Encode()

	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out PublisherStatResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("publisher/stat decode: %w, raw=%s", err, string(raw))
	}
	return &out, out.asError()
}

func (c *ComponentClient) postJSON(endpoint string, body any, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if len(respBody) == 0 {
		return errors.New("empty response")
	}
	return json.Unmarshal(respBody, out)
}

func (c *ComponentClient) getJSON(endpoint string, out any) error {
	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if len(respBody) == 0 {
		return errors.New("empty response")
	}
	return json.Unmarshal(respBody, out)
}
