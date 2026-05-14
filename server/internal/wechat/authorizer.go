package wechat

// 授权方管理 — getAuthorizerList / getAuthorizerInfo
// 文档：https://developers.weixin.qq.com/doc/oplatform/openApi/OpenApiDoc/component-management/getAuthorizerList.html

import "net/url"

// ─── GetAuthorizerList 拉取已授权账号列表 ───

type AuthorizerListItem struct {
	AuthorizerAppID string `json:"authorizer_appid"`
	RefreshToken    string `json:"refresh_token"`
	AuthTime        int64  `json:"auth_time"`
}

type GetAuthorizerListResp struct {
	BaseErr
	List     []AuthorizerListItem `json:"list"`
	TotalNum int                  `json:"total_count"`
}

// GetAuthorizerList 拉取已授权账号列表（每日同步用）
func (c *ComponentClient) GetAuthorizerList(componentAccessToken, componentAppID string, offset, count int) (*GetAuthorizerListResp, error) {
	body := map[string]any{
		"component_appid": componentAppID,
		"offset":          offset,
		"count":           count,
	}
	var out GetAuthorizerListResp
	endpoint := wxAPIBase + "/cgi-bin/component/api_get_authorizer_list?component_access_token=" + url.QueryEscape(componentAccessToken)
	if err := c.postJSON(endpoint, body, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}
