package wechat

// 模板小程序生命周期管理 — 代码部署 / 提审 / 发布 / 回退
// 文档入口：https://developers.weixin.qq.com/doc/oplatform/openApi/OpenApiDoc/miniprogram-management/code-management/commit.html

import "net/url"

// ─── wxa/add_category 设置类目 ───

type CategoryItem struct {
	First       int    `json:"first"`
	Second      int    `json:"second"`
	FirstName   string `json:"first_name,omitempty"`
	SecondName  string `json:"second_name,omitempty"`
	CertiURLs   []string `json:"certicates,omitempty"`
}

type AddCategoryResp struct {
	BaseErr
}

// AddCategory 设置小程序类目（工具→办公）
func (c *ComponentClient) AddCategory(authorizerAccessToken string, categories []CategoryItem) error {
	body := map[string]any{
		"categories": categories,
	}
	var out AddCategoryResp
	endpoint := wxAPIBase + "/cgi-bin/wxopen/addcategory?access_token=" + url.QueryEscape(authorizerAccessToken)
	if err := c.postJSON(endpoint, body, &out); err != nil {
		return err
	}
	return out.asError()
}

// ─── wxa/modify_domain 配置服务器域名 ───

type ModifyDomainReq struct {
	Action          string   `json:"action"` // set / get
	RequestDomain   []string `json:"requestdomain,omitempty"`
	WsRequestDomain []string `json:"wsrequestdomain,omitempty"`
	UploadDomain    []string `json:"uploaddomain,omitempty"`
	DownloadDomain  []string `json:"downloaddomain,omitempty"`
}

type ModifyDomainResp struct {
	BaseErr
}

// ModifyServerDomain 配置小程序服务器域名
func (c *ComponentClient) ModifyServerDomain(authorizerAccessToken string, req ModifyDomainReq) error {
	var out ModifyDomainResp
	endpoint := wxAPIBase + "/wxa/modify_domain?access_token=" + url.QueryEscape(authorizerAccessToken)
	if err := c.postJSON(endpoint, req, &out); err != nil {
		return err
	}
	return out.asError()
}

// ─── wxa/setMpPrivacySetting 设置隐私保护指引 ───

type PrivacySettingReq struct {
	PrivacyVer   int    `json:"privacy_ver"`    // 2
	OwnerSetting map[string]string `json:"owner_setting"` // contact_email, notice_method 等
	SettingList  []PrivacyItem     `json:"setting_list"`
}

type PrivacyItem struct {
	PrivacyKey   string `json:"privacy_key"`
	PrivacyText  string `json:"privacy_text"`
}

type PrivacySettingResp struct {
	BaseErr
}

// SetPrivacySetting 设置小程序隐私保护指引（黄金版本）
func (c *ComponentClient) SetPrivacySetting(authorizerAccessToken string, req PrivacySettingReq) error {
	var out PrivacySettingResp
	endpoint := wxAPIBase + "/cgi-bin/component/setprivacysetting?access_token=" + url.QueryEscape(authorizerAccessToken)
	if err := c.postJSON(endpoint, req, &out); err != nil {
		return err
	}
	return out.asError()
}

// ─── wxa/commit 上传代码 ───

type CommitReq struct {
	TemplateID  int64  `json:"template_id"`
	ExtJSON     string `json:"ext_json"`
	UserVersion string `json:"user_version"`
	UserDesc    string `json:"user_desc"`
}

type CommitResp struct {
	BaseErr
}

// Commit 上传代码到作者小程序开发版（注入 ext_json 差异化配置）
func (c *ComponentClient) Commit(authorizerAccessToken string, req CommitReq) error {
	var out CommitResp
	endpoint := wxAPIBase + "/wxa/commit?access_token=" + url.QueryEscape(authorizerAccessToken)
	if err := c.postJSON(endpoint, req, &out); err != nil {
		return err
	}
	return out.asError()
}

// ─── wxa/get_page 获取小程序页面列表 ───

type GetPageResp struct {
	BaseErr
	PageList []string `json:"page_list"`
}

// GetPage 获取小程序页面列表（提审前用）
func (c *ComponentClient) GetPage(authorizerAccessToken string) (*GetPageResp, error) {
	var out GetPageResp
	endpoint := wxAPIBase + "/wxa/get_page?access_token=" + url.QueryEscape(authorizerAccessToken)
	// GET 请求
	if err := c.getJSON(endpoint, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}

// ─── wxa/submit_audit 提交审核 ───

type AuditItemReq struct {
	Address     string `json:"address,omitempty"`
	Tag         string `json:"tag,omitempty"`
	FirstClass  string `json:"first_class"`
	SecondClass string `json:"second_class"`
	ThirdClass  string `json:"third_class,omitempty"`
	FirstID     int    `json:"first_id"`
	SecondID    int    `json:"second_id"`
	ThirdID     int    `json:"third_id,omitempty"`
	Title       string `json:"title,omitempty"`
}

type PreviewInfo struct {
	VideoIDList []string `json:"video_id_list,omitempty"`
	PicIDList   []string `json:"pic_id_list,omitempty"`
}

type UGCDeclare struct {
	Scene         []int  `json:"scene,omitempty"`
	OtherScene    string `json:"other_scene_desc,omitempty"`
	Method        []int  `json:"method,omitempty"`
	HasAuditTeam  int    `json:"has_audit_team"`
	AuditDesc     string `json:"audit_desc,omitempty"`
}

type SubmitAuditReq struct {
	ItemList        []AuditItemReq `json:"item_list"`
	VersionDesc     string         `json:"version_desc,omitempty"`
	FeedbackInfo    string         `json:"feedback_info,omitempty"`
	FeedbackStuff   string         `json:"feedback_stuff,omitempty"`
	PreviewInfo     *PreviewInfo   `json:"preview_info,omitempty"`
	UGCDeclare      *UGCDeclare    `json:"ugc_declare,omitempty"`
	PrivacyAPINotUse bool          `json:"privacy_api_not_use,omitempty"`
}

type SubmitAuditResp struct {
	BaseErr
	AuditID int64 `json:"auditid"`
}

// SubmitAudit 提交审核
func (c *ComponentClient) SubmitAudit(authorizerAccessToken string, req SubmitAuditReq) (*SubmitAuditResp, error) {
	var out SubmitAuditResp
	endpoint := wxAPIBase + "/wxa/submit_audit?access_token=" + url.QueryEscape(authorizerAccessToken)
	if err := c.postJSON(endpoint, req, &out); err != nil {
		return nil, err
	}
	return &out, out.asError()
}

// ─── wxa/release 发布上线 ───

type ReleaseResp struct {
	BaseErr
}

// Release 发布已审核通过的小程序
func (c *ComponentClient) Release(authorizerAccessToken string) error {
	var out ReleaseResp
	endpoint := wxAPIBase + "/wxa/release?access_token=" + url.QueryEscape(authorizerAccessToken)
	if err := c.postJSON(endpoint, map[string]string{}, &out); err != nil {
		return err
	}
	return out.asError()
}

// ─── wxa/undocodeaudit 撤回审核 ───

// UndoCodeAudit 撤回审核（每天 1 次限制）
func (c *ComponentClient) UndoCodeAudit(authorizerAccessToken string) error {
	var out BaseErr
	endpoint := wxAPIBase + "/wxa/undocodeaudit?access_token=" + url.QueryEscape(authorizerAccessToken)
	if err := c.getJSON(endpoint, &out); err != nil {
		return err
	}
	return out.asError()
}

// ─── wxa/revertcoderelease 版本回退 ───

// RevertCodeRelease 版本回退（紧急事件用）
func (c *ComponentClient) RevertCodeRelease(authorizerAccessToken string) error {
	var out BaseErr
	endpoint := wxAPIBase + "/wxa/revertcoderelease?access_token=" + url.QueryEscape(authorizerAccessToken)
	if err := c.getJSON(endpoint, &out); err != nil {
		return err
	}
	return out.asError()
}

// ─── wxa/get_qrcode 获取体验版二维码 ───

type GetQRCodeResp struct {
	BaseErr
	// 注意：成功时直接返回图片二进制，而非 JSON。
	// 如果 errcode != 0 才是 JSON。调用方需特殊处理。
}

// GetQRCodeURL 返回获取体验版二维码的 URL（调用方自行 GET 下载图片）
func (c *ComponentClient) GetQRCodeURL(authorizerAccessToken, path string) string {
	q := url.Values{}
	q.Set("access_token", authorizerAccessToken)
	if path != "" {
		q.Set("path", path)
	}
	return wxAPIBase + "/wxa/get_qrcode?" + q.Encode()
}
