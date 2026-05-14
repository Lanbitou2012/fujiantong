package api

import (
	"fujiantong/internal/config"
	"fujiantong/internal/service"
	"fujiantong/internal/wechat"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Svc *service.Container
}

func NewAuthHandler(svc *service.Container) *AuthHandler {
	return &AuthHandler{Svc: svc}
}

// openClient 每次请求从最新配置（DB 优先）构造客户端，支持热更新。
// 返回 (client, redirectURI, ok)；ok=false 表示网站应用未配置。
func (h *AuthHandler) openClient() (*wechat.OpenOAuthClient, string, bool) {
	cfg := config.GetOpenPlatformConfig()
	if cfg.AppID == "" || cfg.AppSecret == "" {
		return nil, "", false
	}
	return wechat.NewOpenOAuthClient(cfg.AppID, cfg.AppSecret), cfg.RedirectURI, true
}

// AdminLogin POST /api/v1/auth/admin/login
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	token, user, err := h.Svc.AdminLogin(req.Username, req.Password)
	if err != nil {
		Fail(c, 401, err.Error())
		return
	}
	OK(c, gin.H{"token": token, "user": user})
}

// ScanLoginURL GET /api/v1/auth/scan/url
func (h *AuthHandler) ScanLoginURL(c *gin.Context) {
	state := c.DefaultQuery("state", "login")
	client, redirectURI, ok := h.openClient()
	if !ok {
		Fail(c, 503, "微信扫码登录未配置")
		return
	}
	url := client.BuildLoginURL(redirectURI, state)
	OK(c, gin.H{"url": url})
}

// WxComponentAuthURL GET /api/v1/auth/wx-component/url?promoter_id=xxx&biz_appid=xxx
// 生成"作者授权小程序代运营"的微信扫码链接
func (h *AuthHandler) WxComponentAuthURL(c *gin.Context) {
	promoterID := atoui64(c.Query("promoter_id"))
	bizAppID := c.Query("biz_appid")
	info, err := h.Svc.GenerateAuthURL(promoterID, bizAppID)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, info)
}

// WechatLogin POST /api/v1/auth/wechat/login
func (h *AuthHandler) WechatLogin(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	client, _, ok := h.openClient()
	if !ok {
		Fail(c, 503, "微信扫码登录未配置")
		return
	}
	// code → access_token + openid + unionid
	tokenResp, err := client.Code2Token(req.Code)
	if err != nil {
		FailErr(c, err)
		return
	}
	// 获取用户信息
	userInfo, err := client.UserInfo(tokenResp.AccessToken, tokenResp.OpenID)
	if err != nil {
		FailErr(c, err)
		return
	}
	// 登录/注册
	token, user, err := h.Svc.WechatScanLogin(
		tokenResp.UnionID,
		tokenResp.OpenID,
		userInfo.Nickname,
		userInfo.HeadImgURL,
	)
	if err != nil {
		FailErr(c, err)
		return
	}
	// 处理推广员绑定（入驻时 state=promoter_<id>）
	if pid := c.Query("promoter_id"); pid != "" {
		_ = h.Svc.BindPromoter(user.ID, atoui64(pid))
	}
	OK(c, gin.H{"token": token, "user": user})
}
