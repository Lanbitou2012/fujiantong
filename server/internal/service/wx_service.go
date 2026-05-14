package service

import (
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
// 返回授权对应的 user，便于前端回调时签发 token。
func (c *Container) HandleAuthorizationEvent(componentAppID, componentSecret, authCode string) (*model.User, error) {
	componentToken, err := c.EnsureComponentToken(componentAppID, componentSecret)
	if err != nil {
		return nil, err
	}
	queryResp, err := c.WxComponent.QueryAuth(componentToken, componentAppID, authCode)
	if err != nil {
		return nil, err
	}
	info := queryResp.AuthorizationInfo
	appID := info.AuthorizerAppID
	now := time.Now()

	// 拉小程序资料（昵称 / 头像 / 主体）
	var nickname, avatar, principal string
	if detail, derr := c.WxComponent.GetAuthorizerInfo(componentToken, componentAppID, appID); derr == nil {
		nickname = detail.AuthorizerInfo.NickName
		avatar = detail.AuthorizerInfo.HeadImg
		principal = detail.AuthorizerInfo.PrincipalName
	} else {
		log.Printf("[WxAuth] GetAuthorizerInfo 失败（继续）: %v", derr)
	}

	// 查找用户：先按 bound_appid
	u, err := c.UserRepo.GetByAppID(appID)
	if err != nil {
		// 不存在 → 创建新作者
		// wx_unionid 是 NOT NULL UNIQUE，但 component 授权不返回 unionid，
		// 这里用 "app:<appid>" 作为合成占位值保证唯一性，将来作者真扫码登录时再覆盖。
		newUser := &model.User{
			WxUnionID:              "app:" + appID,
			BoundAppID:             appID,
			Nickname:               nickname,
			AvatarURL:              avatar,
			AuthorizerRefreshToken: info.AuthorizerRefreshToken,
			Status:                 1,
		}
		if nickname == "" {
			newUser.Nickname = "作者_" + appID[len(appID)-6:]
		}
		if err := c.UserRepo.Create(newUser); err != nil {
			return nil, fmt.Errorf("创建作者账号失败: %w", err)
		}
		u = newUser
		log.Printf("[WxAuth] 新作者注册: id=%d appid=%s nickname=%s", u.ID, appID, u.Nickname)
	} else {
		// 已存在 → 更新 refresh_token 和资料
		updates := map[string]any{
			"authorizer_refresh_token": info.AuthorizerRefreshToken,
		}
		if nickname != "" && u.Nickname == "" {
			updates["nickname"] = nickname
		}
		if avatar != "" && u.AvatarURL == "" {
			updates["avatar_url"] = avatar
		}
		_ = c.UserRepo.UpdateFields(u.ID, updates)
	}
	_ = principal // 主体名后续如需保存可加字段

	// 写入/更新授权表
	auth := &model.Authorization{
		AppID:                  appID,
		UserID:                 u.ID,
		AuthorizerRefreshToken: info.AuthorizerRefreshToken,
		Status:                 "authorized",
		AuthorizedAt:           &now,
		GrantedPermissionIDs:   formatFuncInfo(info.FuncInfo),
	}
	if err := c.AuthRepo.Upsert(auth); err != nil {
		return u, err
	}

	// 缓存 authorizer_access_token
	key := "auth_" + appID
	_ = c.WechatRepo.UpsertToken(&model.WxToken{
		Key:          key,
		AccessToken:  info.AuthorizerAccessToken,
		RefreshToken: info.AuthorizerRefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(info.ExpiresIn-300) * time.Second),
	})

	// V3.0：模板小程序自动部署流水线（commit/submit_audit/release）挪到 V1.5。
	// V1.0 的链路是：作者授权完成 → 落库 user → 跳前端 /auth-success 落 token，
	// 作者直接进工作台上传文件、复制超链接到公众号文章。

	return u, nil
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

// V3.0：以下流水线相关函数（triggerDeployPipeline / HandleAuditSuccess / HandleAuditFail）
// 全部已删除，对应能力挪到 V1.5 P1 重新实现。
// 真要重写时可参考 e:\项目\小项目\附件通\CORE_ARCHITECTURE.v2.8.history.md §15.2.8。

func formatFuncInfo(funcInfo []wechat.FuncScopeItem) string {
	ids := make([]string, 0, len(funcInfo))
	for _, f := range funcInfo {
		ids = append(ids, fmt.Sprintf("%d", f.FuncScopeCategory.ID))
	}
	return strings.Join(ids, ",")
}
