package api

import (
	"strings"

	"fujiantong/internal/middleware"
	"fujiantong/internal/model"
	"fujiantong/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type SettingsHandler struct {
	Svc *service.Container
}

func NewSettingsHandler(svc *service.Container) *SettingsHandler {
	return &SettingsHandler{Svc: svc}
}

// ChangePassword POST /api/v1/admin/settings/password
func (h *SettingsHandler) ChangePassword(c *gin.Context) {
	adminID := middleware.CurrentUserID(c)
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "请输入旧密码与新密码（至少 8 位）")
		return
	}
	user, err := h.Svc.UserRepo.GetByID(uint64(adminID))
	if err != nil || user == nil || user.ID == 0 {
		Fail(c, 404, "用户不存在")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.LoginPassphraseHash), []byte(req.OldPassword)); err != nil {
		Fail(c, 401, "原密码不正确")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		FailErr(c, err)
		return
	}
	if err := h.Svc.UserRepo.UpdateFields(user.ID, map[string]any{"login_passphrase_hash": string(hash)}); err != nil {
		FailErr(c, err)
		return
	}
	h.Svc.AdminRepo.CreateLog(&model.AdminActionLog{
		AdminID: uint64(adminID),
		Action:  "change_password",
		Target:  "self",
		IP:      c.ClientIP(),
	})
	OKMsg(c, "密码修改成功")
}

// GetPlatformSettings GET /api/v1/admin/settings/platform
// 返回当前所有可配置项；敏感字段脱敏（仅显示是否已配置 + 末尾 4 位）
func (h *SettingsHandler) GetPlatformSettings(c *gin.Context) {
	keys := []string{
		model.KeyWxComponentAppID,
		model.KeyWxComponentAppSecret,
		model.KeyWxComponentToken,
		model.KeyWxComponentAESKey,
		model.KeyWechatOpenAppID,
		model.KeyWechatOpenSecret,
		model.KeyWechatOpenRedirect,
		model.KeyStorageDriver,
		model.KeyCOSBucketURL,
		model.KeyCOSSecretID,
		model.KeyCOSSecretKey,
		model.KeyCOSCDNBaseURL,
	}
	raw := h.Svc.SettingRepo.GetAll(keys)
	masked := make(map[string]any, len(raw))
	for k, v := range raw {
		if isSecretKey(k) {
			masked[k] = mask(v)
		} else {
			masked[k] = v
		}
	}
	masked["_secret_set"] = secretSetMap(raw)
	OK(c, masked)
}

// UpdatePlatformSettings PUT /api/v1/admin/settings/platform
// 接收前端提交的部分字段；
//   - 敏感字段（secret/key）：空字符串视为不修改（保留原值）
//   - 普通字段：直接覆盖（含空字符串，用于清空）
//   - 特殊值 "__CLEAR__"：强制清空（用于清掉敏感字段）
func (h *SettingsHandler) UpdatePlatformSettings(c *gin.Context) {
	adminID := middleware.CurrentUserID(c)
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}

	allowedKeys := map[string]string{
		model.KeyWxComponentAppID:     "wechat",
		model.KeyWxComponentAppSecret: "wechat",
		model.KeyWxComponentToken:     "wechat",
		model.KeyWxComponentAESKey:    "wechat",
		model.KeyWechatOpenAppID:      "wechat",
		model.KeyWechatOpenSecret:     "wechat",
		model.KeyWechatOpenRedirect:   "wechat",
		model.KeyStorageDriver:        "cos",
		model.KeyCOSBucketURL:         "cos",
		model.KeyCOSSecretID:          "cos",
		model.KeyCOSSecretKey:         "cos",
		model.KeyCOSCDNBaseURL:        "cos",
	}

	groups := map[string]map[string]string{
		"wechat": {},
		"cos":    {},
	}
	for k, v := range req {
		cat, ok := allowedKeys[k]
		if !ok {
			continue
		}
		// 特殊值 __CLEAR__：强制清空（无论是否敏感字段）
		if v == "__CLEAR__" {
			groups[cat][k] = ""
			continue
		}
		// 敏感字段空字符串 = 保留旧值，不更新
		if isSecretKey(k) && strings.TrimSpace(v) == "" {
			continue
		}
		groups[cat][k] = v
	}

	for cat, items := range groups {
		if len(items) == 0 {
			continue
		}
		if err := h.Svc.SettingRepo.SetMany(items, cat, uint64(adminID)); err != nil {
			FailErr(c, err)
			return
		}
	}

	h.Svc.AdminRepo.CreateLog(&model.AdminActionLog{
		AdminID: uint64(adminID),
		Action:  "update_settings",
		Target:  "platform",
		IP:      c.ClientIP(),
	})
	OKMsg(c, "保存成功，配置已实时生效")
}

// ─── helpers ───

func isSecretKey(k string) bool {
	switch k {
	case model.KeyWxComponentAppSecret,
		model.KeyWxComponentToken,
		model.KeyWxComponentAESKey,
		model.KeyWechatOpenSecret,
		model.KeyCOSSecretID,
		model.KeyCOSSecretKey:
		return true
	}
	return false
}

func mask(v string) string {
	if v == "" {
		return ""
	}
	if len(v) <= 4 {
		return "****"
	}
	return "****" + v[len(v)-4:]
}

func secretSetMap(raw map[string]string) map[string]bool {
	out := make(map[string]bool)
	for k, v := range raw {
		if isSecretKey(k) {
			out[k] = v != ""
		}
	}
	return out
}
