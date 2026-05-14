package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"fujiantong/internal/model"
	"fujiantong/internal/wechat"
)

// HandleVerifyTicket 接收并存储 component_verify_ticket
func (c *Container) HandleVerifyTicket(ticket string) error {
	return c.WechatRepo.SaveVerifyTicket(ticket)
}

// EnsureComponentToken 确保 component_access_token 有效（提前 5 分钟刷新）
func (c *Container) EnsureComponentToken(componentAppID, componentSecret string) (string, error) {
	expired, err := c.WechatRepo.IsTokenExpired("component")
	if err != nil {
		return "", err
	}
	if !expired {
		t, _ := c.WechatRepo.GetToken("component")
		return t.AccessToken, nil
	}
	// 需要刷新
	ticket, err := c.WechatRepo.GetVerifyTicket()
	if err != nil {
		return "", fmt.Errorf("无法获取 verify_ticket: %w", err)
	}
	resp, err := c.WxComponent.GetComponentAccessToken(componentAppID, componentSecret, ticket)
	if err != nil {
		return "", err
	}
	err = c.WechatRepo.UpsertToken(&model.WxToken{
		Key:         "component",
		AccessToken: resp.ComponentAccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(resp.ExpiresIn-300) * time.Second),
	})
	return resp.ComponentAccessToken, err
}

// EnsureAuthorizerToken 确保某作者小程序的 access_token 有效
func (c *Container) EnsureAuthorizerToken(componentAppID, componentSecret, appID string) (string, error) {
	key := "auth_" + appID
	expired, err := c.WechatRepo.IsTokenExpired(key)
	if err != nil {
		return "", err
	}
	if !expired {
		t, _ := c.WechatRepo.GetToken(key)
		return t.AccessToken, nil
	}
	// 刷新
	componentToken, err := c.EnsureComponentToken(componentAppID, componentSecret)
	if err != nil {
		return "", err
	}
	auth, err := c.AuthRepo.GetByAppID(appID)
	if err != nil {
		return "", fmt.Errorf("未找到 appid=%s 的授权记录: %w", appID, err)
	}
	resp, err := c.WxComponent.RefreshAuthorizerToken(componentToken, componentAppID, appID, auth.AuthorizerRefreshToken)
	if err != nil {
		return "", err
	}
	// 更新 refresh_token（微信可能会更新）
	if resp.AuthorizerRefreshToken != "" && resp.AuthorizerRefreshToken != auth.AuthorizerRefreshToken {
		auth.AuthorizerRefreshToken = resp.AuthorizerRefreshToken
		_ = c.AuthRepo.Update(auth)
	}
	err = c.WechatRepo.UpsertToken(&model.WxToken{
		Key:          key,
		AccessToken:  resp.AuthorizerAccessToken,
		RefreshToken: resp.AuthorizerRefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.ExpiresIn-300) * time.Second),
	})
	return resp.AuthorizerAccessToken, err
}

// HandleAuthorizationEvent 处理授权事件（新增 / 更新授权）
func (c *Container) HandleAuthorizationEvent(componentAppID, componentSecret, authCode string) error {
	componentToken, err := c.EnsureComponentToken(componentAppID, componentSecret)
	if err != nil {
		return err
	}
	queryResp, err := c.WxComponent.QueryAuth(componentToken, componentAppID, authCode)
	if err != nil {
		return err
	}
	info := queryResp.AuthorizationInfo
	appID := info.AuthorizerAppID
	now := time.Now()

	// 写入/更新授权表
	auth := &model.Authorization{
		AppID:                  appID,
		AuthorizerRefreshToken: info.AuthorizerRefreshToken,
		Status:                 "authorized",
		AuthorizedAt:           &now,
		GrantedPermissionIDs:   formatFuncInfo(info.FuncInfo),
	}

	// 关联用户（通过 bound_appid 查找）
	u, err := c.UserRepo.GetByAppID(appID)
	if err == nil {
		auth.UserID = u.ID
		// 同步 refresh_token 到用户表
		_ = c.UserRepo.UpdateFields(u.ID, map[string]any{
			"authorizer_refresh_token": info.AuthorizerRefreshToken,
		})
	}

	if err := c.AuthRepo.Upsert(auth); err != nil {
		return err
	}

	// 缓存 authorizer_access_token
	key := "auth_" + appID
	_ = c.WechatRepo.UpsertToken(&model.WxToken{
		Key:          key,
		AccessToken:  info.AuthorizerAccessToken,
		RefreshToken: info.AuthorizerRefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(info.ExpiresIn-300) * time.Second),
	})

	// 触发自动部署流水线（异步）
	go c.triggerDeployPipeline(componentAppID, componentSecret, appID)

	return nil
}

// HandleDeauthorizationEvent 取消授权
func (c *Container) HandleDeauthorizationEvent(appID string) error {
	auth, err := c.AuthRepo.GetByAppID(appID)
	if err != nil {
		return err
	}
	now := time.Now()
	auth.Status = "deauthorized"
	auth.DeauthorizedAt = &now
	return c.AuthRepo.Update(auth)
}

// triggerDeployPipeline 授权后自动部署流水线（§15.2.8）
func (c *Container) triggerDeployPipeline(componentAppID, componentSecret, appID string) {
	log.Printf("[DeployPipeline] 开始为 %s 执行自动部署流水线", appID)

	// 创建 deployment 记录
	deploy := &model.MpDeployment{
		AppID:  appID,
		Status: "pending",
	}
	// 查用户
	if auth, err := c.AuthRepo.GetByAppID(appID); err == nil {
		deploy.UserID = auth.UserID
	}
	if err := c.DeploymentRepo.CreateDeployment(deploy); err != nil {
		log.Printf("[DeployPipeline] 创建部署记录失败: %v", err)
		return
	}

	// 获取 authorizer_access_token
	authToken, err := c.EnsureAuthorizerToken(componentAppID, componentSecret, appID)
	if err != nil {
		log.Printf("[DeployPipeline] 获取 authorizer token 失败: %v", err)
		c.DeploymentRepo.UpdateDeploymentStatus(deploy.ID, "configuring_failed")
		return
	}

	// ─── 阶段 1：基础配置 ───
	c.DeploymentRepo.UpdateDeploymentStatus(deploy.ID, "configuring")

	componentToken, _ := c.EnsureComponentToken(componentAppID, componentSecret)

	// 1.1 SetShareRatio(28)
	if err := c.WxComponent.SetShareRatio(componentToken, wechat.SetShareRatioReq{
		ShareRatio:   28,
		AuthorizerID: appID,
	}); err != nil {
		log.Printf("[DeployPipeline] SetShareRatio 失败: %v", err)
	}

	// 1.2 AgencyCreateAdunit × 3
	adSlots := []struct {
		slot string
		name string
	}{
		{"slot_id_reward_video", "激励视频"},
		{"slot_id_interstitial", "插屏广告"},
		{"slot_id_splash", "开屏广告"},
	}
	for _, s := range adSlots {
		resp, err := c.WxComponent.AgencyCreateAdunit(authToken, wechat.CreateAdunitReq{
			AdSlot: s.slot,
			Name:   s.name,
		})
		if err != nil {
			log.Printf("[DeployPipeline] 创建广告位 %s 失败: %v", s.slot, err)
		} else {
			log.Printf("[DeployPipeline] 广告位 %s 创建成功: %s", s.slot, resp.AdUnitID)
		}
	}

	// 1.3 wxa/add_category（工具→办公）
	if err := c.WxComponent.AddCategory(authToken, []wechat.CategoryItem{
		{First: 287, Second: 296, FirstName: "工具", SecondName: "办公"},
	}); err != nil {
		log.Printf("[DeployPipeline] AddCategory 失败: %v", err)
	}

	// 1.4 wxa/modifyserverdomain
	if err := c.WxComponent.ModifyServerDomain(authToken, wechat.ModifyDomainReq{
		Action:         "set",
		RequestDomain:  []string{"https://fujian.5g6g.top"},
		UploadDomain:   []string{"https://fujian.5g6g.top"},
		DownloadDomain: []string{"https://fujian.5g6g.top"},
	}); err != nil {
		log.Printf("[DeployPipeline] ModifyServerDomain 失败: %v", err)
	}

	// 1.5 wxa/setMpPrivacySetting（黄金版本）
	if err := c.WxComponent.SetPrivacySetting(authToken, wechat.PrivacySettingReq{
		PrivacyVer: 2,
		OwnerSetting: map[string]string{
			"contact_email": "552002521@qq.com",
			"notice_method": "弹窗提示",
		},
		SettingList: []wechat.PrivacyItem{
			{PrivacyKey: "UserInfo", PrivacyText: "用于展示您的头像和昵称"},
			{PrivacyKey: "Location", PrivacyText: "不收集位置信息"},
		},
	}); err != nil {
		log.Printf("[DeployPipeline] SetPrivacySetting 失败: %v", err)
	}

	// ─── 阶段 2：代码部署 ───
	c.DeploymentRepo.UpdateDeploymentStatus(deploy.ID, "committing")

	tmpl, err := c.DeploymentRepo.GetActiveTemplate()
	if err != nil {
		log.Printf("[DeployPipeline] 未找到激活模板版本: %v", err)
		c.DeploymentRepo.UpdateDeploymentStatus(deploy.ID, "waiting_template_fix")
		return
	}

	extJSON := buildExtJSON(appID, deploy.UserID)
	deploy.TemplateVersionID = tmpl.ID
	deploy.TemplateID = tmpl.TemplateID
	deploy.UserVersion = tmpl.UserVersion
	deploy.ExtJSON = extJSON

	if err := c.WxComponent.Commit(authToken, wechat.CommitReq{
		TemplateID:  tmpl.TemplateID,
		ExtJSON:     extJSON,
		UserVersion: tmpl.UserVersion + fmt.Sprintf("-uid%d", deploy.UserID),
		UserDesc:    tmpl.UserDesc,
	}); err != nil {
		log.Printf("[DeployPipeline] Commit 失败: %v", err)
		c.DeploymentRepo.UpdateDeploymentStatus(deploy.ID, "committing_failed")
		return
	}

	// ─── 阶段 3：提审 ───
	c.DeploymentRepo.UpdateDeploymentStatus(deploy.ID, "submitted_for_audit")

	auditResp, err := c.WxComponent.SubmitAudit(authToken, wechat.SubmitAuditReq{
		ItemList: []wechat.AuditItemReq{
			{FirstClass: "工具", SecondClass: "办公", FirstID: 287, SecondID: 296, Title: "附件通"},
		},
		VersionDesc:  "附件通 " + tmpl.UserVersion,
		FeedbackInfo: "552002521@qq.com",
		UGCDeclare: &wechat.UGCDeclare{
			Scene:        []int{1, 2},
			Method:       []int{1},
			HasAuditTeam: 1,
			AuditDesc:    "平台对所有用户上传的附件进行机审（msgSecCheck + mediaCheckAsync），违规内容自动拦截",
		},
		PrivacyAPINotUse: true,
	})
	if err != nil {
		log.Printf("[DeployPipeline] SubmitAudit 失败: %v", err)
		c.DeploymentRepo.UpdateDeploymentStatus(deploy.ID, "audit_submit_failed")
		return
	}

	deploy.AuditID = auditResp.AuditID
	_ = c.DeploymentRepo.UpdateDeployment(deploy)

	// 创建审核记录
	_ = c.DeploymentRepo.CreateAudit(&model.MpAudit{
		AppID:        appID,
		UserID:       deploy.UserID,
		DeploymentID: deploy.ID,
		WxAuditID:    auditResp.AuditID,
		VersionDesc:  tmpl.UserVersion,
		Status:       "submitted",
		SubmittedAt:  time.Now(),
	})

	log.Printf("[DeployPipeline] %s 提审成功, auditid=%d", appID, auditResp.AuditID)
}

// HandleAuditSuccess 审核通过 → 自动发布
func (c *Container) HandleAuditSuccess(appID string, componentAppID, componentSecret string) error {
	authToken, err := c.EnsureAuthorizerToken(componentAppID, componentSecret, appID)
	if err != nil {
		return err
	}
	if err := c.WxComponent.Release(authToken); err != nil {
		return fmt.Errorf("release 失败: %w", err)
	}
	// 更新部署状态
	deploy, err := c.DeploymentRepo.GetLatestDeployment(appID)
	if err == nil {
		deploy.Status = "live"
		_ = c.DeploymentRepo.UpdateDeployment(deploy)
	}
	return nil
}

// HandleAuditFail 审核驳回 → 分流
func (c *Container) HandleAuditFail(appID, reason string) error {
	deploy, err := c.DeploymentRepo.GetLatestDeployment(appID)
	if err != nil {
		return err
	}
	deploy.FailReason = reason

	// 关键词分流
	category := "unknown"
	reasonLower := reason
	if contains(reasonLower, "代码", "code", "bug") {
		category = "code"
		deploy.Status = "waiting_template_fix"
	} else if contains(reasonLower, "资质", "类目", "主体", "证件") {
		category = "qualification"
		deploy.Status = "waiting_author_supplement"
	} else {
		deploy.Status = "manual_review"
	}

	_ = c.DeploymentRepo.UpdateDeployment(deploy)

	// 更新审核记录
	audit, err := c.DeploymentRepo.GetAuditByDeployment(deploy.ID)
	if err == nil {
		now := time.Now()
		audit.Status = "fail"
		audit.FailReason = reason
		audit.FailCategory = category
		audit.ResultAt = &now
		_ = c.DeploymentRepo.UpdateAudit(audit)
	}
	return nil
}

func buildExtJSON(appID string, userID uint64) string {
	ext := map[string]any{
		"extEnable": true,
		"extAppid":  appID,
		"ext": map[string]any{
			"platformUserId": fmt.Sprintf("uid_%d", userID),
			"apiBaseUrl":     "https://fujian.5g6g.top",
			"privacyVersion": "v1.0",
		},
	}
	b, _ := json.Marshal(ext)
	return string(b)
}

func contains(s string, keywords ...string) bool {
	for _, k := range keywords {
		if len(s) >= len(k) {
			for i := 0; i <= len(s)-len(k); i++ {
				if s[i:i+len(k)] == k {
					return true
				}
			}
		}
	}
	return false
}

func formatFuncInfo(funcInfo []wechat.FuncScopeItem) string {
	ids := make([]string, 0, len(funcInfo))
	for _, f := range funcInfo {
		ids = append(ids, fmt.Sprintf("%d", f.FuncScopeCategory.ID))
	}
	return strings.Join(ids, ",")
}
