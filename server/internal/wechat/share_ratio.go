package wechat

// 流量主代运营 — 分账比例设置 / 查询
// 文档：https://developers.weixin.qq.com/doc/oplatform/openApi/OpenApiDoc/ams/open/SetShareRatio.html

import "net/url"

// ─── SetShareRatio 设置默认/单作者分账比例 ───

type SetShareRatioReq struct {
	ShareRatio   int    `json:"share_ratio"`             // 服务商占比 0-100（附件通默认 28）
	AuthorizerID string `json:"authorizer_appid,omitempty"` // 为空时设全局默认，填值时设单作者
}

type SetShareRatioResp struct {
	BaseErr
}

// SetShareRatio 设置分账比例（默认 or 差异化）
// action=set_share_ratio
func (c *ComponentClient) SetShareRatio(componentAccessToken string, req SetShareRatioReq) error {
	var out SetShareRatioResp
	endpoint := wxAPIBase + "/wxa/setdefaultamsinfo?action=set_share_ratio&component_access_token=" + url.QueryEscape(componentAccessToken)
	if err := c.postJSON(endpoint, req, &out); err != nil {
		return err
	}
	return out.asError()
}

// ─── GetShareRatio 查询分账比例 ───

type GetShareRatioResp struct {
	BaseErr
	ShareRatio int `json:"share_ratio"`
}

// GetShareRatio 查询默认/单作者分账比例
// action=get_share_ratio
func (c *ComponentClient) GetShareRatio(componentAccessToken string, authorizerAppID string) (*GetShareRatioResp, error) {
	body := map[string]string{}
	if authorizerAppID != "" {
		body["authorizer_appid"] = authorizerAppID
	}
	var out GetShareRatioResp
	endpoint := wxAPIBase + "/wxa/getdefaultamsinfo?action=get_share_ratio&component_access_token=" + url.QueryEscape(componentAccessToken)
	if err := c.postJSON(endpoint, body, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}
