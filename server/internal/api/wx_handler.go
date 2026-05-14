package api

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"

	"fujiantong/internal/config"
	"fujiantong/internal/service"
	"fujiantong/internal/wechat"

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
		if err := h.Svc.HandleAuthorizationEvent(componentCfg.AppID, componentCfg.AppSecret, msg.AuthCode); err != nil {
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

	componentCfg := config.GetComponentConfig()
	switch evt.Event {
	case "weapp_audit_success":
		log.Printf("[WxNotify] %s 审核通过", appID)
		if err := h.Svc.HandleAuditSuccess(appID, componentCfg.AppID, componentCfg.AppSecret); err != nil {
			log.Printf("[WxNotify] 发布失败: %v", err)
		}
	case "weapp_audit_fail":
		log.Printf("[WxNotify] %s 审核驳回: %s", appID, evt.Reason)
		_ = h.Svc.HandleAuditFail(appID, evt.Reason)
	case "weapp_audit_delay":
		log.Printf("[WxNotify] %s 审核延期", appID)
	default:
		log.Printf("[WxNotify] appid=%s 未知事件: %s", appID, evt.Event)
	}

	c.String(http.StatusOK, "success")
}

// AuthCallback GET /api/v1/wx/component/callback
// 授权成功回调（用户在微信确认授权后跳回此页）
func (h *WxHandler) AuthCallback(c *gin.Context) {
	authCode := c.Query("auth_code")
	if authCode == "" {
		c.String(http.StatusBadRequest, "missing auth_code")
		return
	}
	componentCfg := config.GetComponentConfig()
	if err := h.Svc.HandleAuthorizationEvent(componentCfg.AppID, componentCfg.AppSecret, authCode); err != nil {
		log.Printf("[WxCallback] AuthCallback 处理失败: %v", err)
		c.String(http.StatusInternalServerError, "授权处理失败")
		return
	}
	// 跳转到前端成功页
	c.Redirect(http.StatusFound, "/auth-success")
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
