package api

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"fujiantong/internal/config"
	"fujiantong/internal/service"
	"fujiantong/internal/wechat"
	jwtpkg "fujiantong/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type WxHandler struct {
	Svc *service.Container
}

func NewWxHandler(svc *service.Container) *WxHandler {
	return &WxHandler{Svc: svc}
}

// xmlMsg 微信推送的通用 XML 消息体
type xmlMsg struct {
	XMLName    xml.Name `xml:"xml"`
	AppID      string   `xml:"AppId"`
	InfoType   string   `xml:"InfoType"`
	AuthCode   string   `xml:"AuthorizationCode"`
	AuthAppID  string   `xml:"AuthorizerAppid"`
	Ticket     string   `xml:"ComponentVerifyTicket"`
	ToUserName string   `xml:"ToUserName"`
}

// ComponentTicket POST /api/v1/wx/component/ticket
// 接收 component_verify_ticket 推送
func (h *WxHandler) ComponentTicket(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	cfg := config.GetComponentConfig()

	// 解密
	msgSign := c.Query("msg_signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")

	if !wechat.SignatureValid(cfg.Token, timestamp, nonce, string(body), msgSign) {
		log.Printf("[WxCallback] ticket 签名校验失败")
	}

	aesKey, _ := wechat.DecodeAESKey(cfg.EncodingAESKey)
	// 从 XML 中提取 Encrypt 字段
	type encXML struct {
		Encrypt string `xml:"Encrypt"`
	}
	var enc encXML
	xml.Unmarshal(body, &enc)
	decrypted, _, err := wechat.Decrypt(enc.Encrypt, aesKey)
	if err != nil {
		log.Printf("[WxCallback] ticket 解密失败: %v", err)
		c.String(http.StatusOK, "success")
		return
	}

	var msg xmlMsg
	if err := xml.Unmarshal(decrypted, &msg); err != nil {
		log.Printf("[WxCallback] ticket XML 解析失败: %v", err)
		c.String(http.StatusOK, "success")
		return
	}

	switch msg.InfoType {
	case "component_verify_ticket":
		if err := h.Svc.HandleVerifyTicket(msg.Ticket); err != nil {
			log.Printf("[WxCallback] 保存 verify_ticket 失败: %v", err)
		} else {
			log.Printf("[WxCallback] verify_ticket 已更新")
		}
	case "authorized", "updateauthorized":
		componentCfg := config.GetComponentConfig()
		if _, err := h.Svc.HandleAuthorizationEvent(componentCfg.AppID, componentCfg.AppSecret, msg.AuthCode); err != nil {
			log.Printf("[WxCallback] 处理授权事件失败: %v", err)
		}
	case "unauthorized":
		if err := h.Svc.HandleDeauthorizationEvent(msg.AuthAppID); err != nil {
			log.Printf("[WxCallback] 处理取消授权失败: %v", err)
		}
	default:
		log.Printf("[WxCallback] 未知 InfoType: %s", msg.InfoType)
	}

	c.String(http.StatusOK, "success")
}

// ComponentNotify POST /api/v1/wx/component/notify/:appid
// 接收审核事件推送（weapp_audit_success / weapp_audit_fail / weapp_audit_delay）
func (h *WxHandler) ComponentNotify(c *gin.Context) {
	appID := c.Param("appid")
	body, _ := io.ReadAll(c.Request.Body)
	cfg := config.GetComponentConfig()

	aesKey, _ := wechat.DecodeAESKey(cfg.EncodingAESKey)
	type encXML2 struct {
		Encrypt string `xml:"Encrypt"`
	}
	var enc2 encXML2
	xml.Unmarshal(body, &enc2)
	decrypted2, _, err := wechat.Decrypt(enc2.Encrypt, aesKey)
	if err != nil {
		log.Printf("[WxNotify] 解密失败 appid=%s: %v", appID, err)
		c.String(http.StatusOK, "success")
		return
	}

	// 解析审核事件
	type auditEvent struct {
		XMLName  xml.Name `xml:"xml"`
		MsgType  string   `xml:"MsgType"`
		Event    string   `xml:"Event"`
		SuccTime int64    `xml:"SuccTime"`
		FailTime int64    `xml:"FailTime"`
		Reason   string   `xml:"Reason"`
	}
	var evt auditEvent
	if err := xml.Unmarshal(decrypted2, &evt); err != nil {
		log.Printf("[WxNotify] XML 解析失败: %v", err)
		c.String(http.StatusOK, "success")
		return
	}

	// V3.0：模板小程序自动部署 / 提审 / 发布流水线已挪到 V1.5 P1。
	// 这里仅记录审核事件日志，不再触发任何动作。
	switch evt.Event {
	case "weapp_audit_success":
		log.Printf("[WxNotify] %s 审核通过（V3.0 不自动 release）", appID)
	case "weapp_audit_fail":
		log.Printf("[WxNotify] %s 审核驳回: %s", appID, evt.Reason)
	case "weapp_audit_delay":
		log.Printf("[WxNotify] %s 审核延期", appID)
	default:
		log.Printf("[WxNotify] appid=%s 未知事件: %s", appID, evt.Event)
	}

	c.String(http.StatusOK, "success")
}

// AuthCallback GET /api/v1/wx/component/callback
// 作者在微信完成授权后，微信浏览器跳回此页：
//
//	?auth_code=xxx&expires_in=xxx&promoter_id=N (透传 redirect_uri 里的 query)
//
// 流程：换取 refresh_token → 创建/更新作者账号 → 绑推广员 → 签发 JWT → 重定向到前端 /auth-success
func (h *WxHandler) AuthCallback(c *gin.Context) {
	authCode := c.Query("auth_code")
	if authCode == "" {
		c.Redirect(http.StatusFound, "/auth-success?error=missing_auth_code")
		return
	}
	componentCfg := config.GetComponentConfig()
	user, err := h.Svc.HandleAuthorizationEvent(componentCfg.AppID, componentCfg.AppSecret, authCode)
	if err != nil {
		log.Printf("[WxCallback] AuthCallback 处理失败: %v", err)
		c.Redirect(http.StatusFound, "/auth-success?error="+err.Error())
		return
	}

	// 绑定推广员（从 query 取，redirect_uri 拼接时塞进来）
	if pid := c.Query("promoter_id"); pid != "" {
		if promoterID, _ := strconv.ParseUint(pid, 10, 64); promoterID > 0 {
			if berr := h.Svc.BindPromoter(user.ID, promoterID); berr != nil {
				log.Printf("[WxCallback] 绑定推广员 %d 失败: %v", promoterID, berr)
			}
		}
	}

	// 签发 JWT
	token, err := jwtpkg.GenerateTokenWithOptions(jwtpkg.TokenOptions{
		UserID:          uint(user.ID),
		Role:            int8(1), // author
		RoleName:        "author",
		AuthorID:        user.ID,
		AuthorizerAppID: user.BoundAppID,
		TTL:             7 * 24 * time.Hour,
	})
	if err != nil {
		log.Printf("[WxCallback] 签发 token 失败: %v", err)
		c.Redirect(http.StatusFound, "/auth-success?error=token_failed")
		return
	}

	// 重定向到前端授权成功页，token 放在 hash 里（避免被服务器日志记录）
	c.Redirect(http.StatusFound, fmt.Sprintf(
		"/auth-success?appid=%s&user_id=%d#token=%s",
		user.BoundAppID, user.ID, token,
	))
}

// MediaCheckCallback POST /api/v1/wx/media-check/:appid
func (h *WxHandler) MediaCheckCallback(c *gin.Context) {
	var req struct {
		TraceID string `json:"trace_id"`
		Result  struct {
			Suggest string `json:"suggest"` // risky / pass
		} `json:"result"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusOK, "success")
		return
	}
	status := "pass"
	if req.Result.Suggest == "risky" {
		status = "reject"
	}
	if err := h.Svc.UpdateMediaCheckStatus(req.TraceID, status); err != nil {
		log.Printf("[MediaCheck] 更新失败 trace=%s: %v", req.TraceID, err)
	}
	c.String(http.StatusOK, "success")
}
