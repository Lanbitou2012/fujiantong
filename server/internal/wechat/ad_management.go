package wechat

// 流量主代运营 — 广告位管理 + 数据拉取 + 结算
// 文档：https://developers.weixin.qq.com/doc/oplatform/openApi/OpenApiDoc/ams/

import (
	"net/url"
)

// ─── AgencyCreateAdunit 创建广告位 ───

type CreateAdunitReq struct {
	AdSlot string `json:"ad_slot"` // slot_id_interstitial / slot_id_reward_video / slot_id_splash 等
	Name   string `json:"name"`    // 广告位名称
}

type CreateAdunitResp struct {
	BaseErr
	AdUnitID string `json:"ad_unit_id"`
}

// AgencyCreateAdunit 为作者小程序创建广告位（authorizer_access_token）
func (c *ComponentClient) AgencyCreateAdunit(authorizerAccessToken string, req CreateAdunitReq) (*CreateAdunitResp, error) {
	var out CreateAdunitResp
	endpoint := wxAPIBase + "/wxa/operationams?action=agency_create_adunit&access_token=" + url.QueryEscape(authorizerAccessToken)
	if err := c.postJSON(endpoint, req, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}

// ─── GetAdposDetail 拉取广告数据详情 ───

type GetAdposDetailReq struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	StartDate string `json:"start_date"` // YYYY-MM-DD
	EndDate   string `json:"end_date"`
	AdSlot    string `json:"ad_slot,omitempty"`
}

type AdposItem struct {
	Date          string  `json:"date"`
	SlotID        string  `json:"slot_id"`
	AdSlot        string  `json:"ad_slot"`
	Income        int64   `json:"income"` // 单位：分
	ExposureCount int64   `json:"exposure_count"`
	ClickCount    int64   `json:"click_count"`
	ECPM          float64 `json:"ecpm"`
}

type GetAdposDetailResp struct {
	BaseErr
	List     []AdposItem `json:"list"`
	TotalNum int         `json:"total_num"`
}

// GetAdposDetail 拉取广告位详细数据（api_getadposdetail）
func (c *ComponentClient) GetAdposDetail(componentAccessToken string, authorizerAppID string, req GetAdposDetailReq) (*GetAdposDetailResp, error) {
	body := map[string]any{
		"page":       req.Page,
		"page_size":  req.PageSize,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
	}
	if req.AdSlot != "" {
		body["ad_slot"] = req.AdSlot
	}
	var out GetAdposDetailResp
	endpoint := wxAPIBase + "/wxa/operationams?action=agency_get_adpos_detail&component_access_token=" + url.QueryEscape(componentAccessToken) + "&authorizer_appid=" + url.QueryEscape(authorizerAppID)
	if err := c.postJSON(endpoint, body, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}

// ─── GetSettlement 拉取结算数据 ───

type GetSettlementReq struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Month    string `json:"month,omitempty"` // YYYY-MM
}

type SettlementItem struct {
	SettlementDate string `json:"settlement_date"`
	SettledRevenue int64  `json:"settled_revenue"`  // 分
	SettAmount     int64  `json:"sett_amount"`      // 分
	SettNoTaxAmt   int64  `json:"sett_no_tax_amt"`
	SettStatus     int    `json:"sett_status"`
}

type GetSettlementResp struct {
	BaseErr
	List     []SettlementItem `json:"list"`
	TotalNum int              `json:"total_num"`
}

// GetSettlement 拉取作者小程序结算数据（api_getsettlement）
func (c *ComponentClient) GetSettlement(componentAccessToken string, authorizerAppID string, req GetSettlementReq) (*GetSettlementResp, error) {
	body := map[string]any{
		"page":      req.Page,
		"page_size": req.PageSize,
	}
	if req.Month != "" {
		body["month"] = req.Month
	}
	var out GetSettlementResp
	endpoint := wxAPIBase + "/wxa/operationams?action=agency_get_settlement&component_access_token=" + url.QueryEscape(componentAccessToken) + "&authorizer_appid=" + url.QueryEscape(authorizerAppID)
	if err := c.postJSON(endpoint, body, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}

// ─── GetAgencySettlement 拉取服务商汇总结算数据（对账用） ───

type GetAgencySettlementResp struct {
	BaseErr
	List     []SettlementItem `json:"list"`
	TotalNum int              `json:"total_num"`
}

// GetAgencySettlement 拉取服务商汇总结算
func (c *ComponentClient) GetAgencySettlement(componentAccessToken string, req GetSettlementReq) (*GetAgencySettlementResp, error) {
	body := map[string]any{
		"page":      req.Page,
		"page_size": req.PageSize,
	}
	if req.Month != "" {
		body["month"] = req.Month
	}
	var out GetAgencySettlementResp
	endpoint := wxAPIBase + "/wxa/operationams?action=agency_get_agency_settlement&component_access_token=" + url.QueryEscape(componentAccessToken)
	if err := c.postJSON(endpoint, body, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}
