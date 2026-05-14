package api

import (
	"fujiantong/internal/middleware"
	"fujiantong/internal/model"
	"fujiantong/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	Svc *service.Container
}

func NewAdminHandler(svc *service.Container) *AdminHandler {
	return &AdminHandler{Svc: svc}
}

// Dashboard GET /api/v1/admin/dashboard
func (h *AdminHandler) Dashboard(c *gin.Context) {
	stats, err := h.Svc.GetDashboardStats()
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, stats)
}

// TodoList GET /api/v1/admin/todo
func (h *AdminHandler) TodoList(c *gin.Context) {
	pending, _ := h.Svc.FinanceRepo.ListPendingCommissions()
	// 分为待审核和待打款
	var toApprove, toPay int
	for _, cs := range pending {
		switch cs.Status {
		case "pending":
			toApprove++
		case "approved":
			toPay++
		}
	}
	// 待复审文件
	var bannedCount int64
	h.Svc.DB.Model(&model.File{}).Where("media_check_status = 'reject'").Count(&bannedCount)

	// 部署异常
	deployStats, _ := h.Svc.DeploymentRepo.CountByStatus()

	OK(c, gin.H{
		"settlements_to_approve": toApprove,
		"settlements_to_pay":     toPay,
		"files_to_review":        bannedCount,
		"deploy_status":          deployStats,
	})
}

// ListUsers GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, size := pageParams(c)
	keyword := c.Query("keyword")
	list, total, err := h.Svc.UserRepo.ListAll(page, size, keyword)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, gin.H{"list": list, "total": total, "page": page, "size": size})
}

// TogglePromoter PUT /api/v1/admin/users/:id/promoter
func (h *AdminHandler) TogglePromoter(c *gin.Context) {
	adminID := middleware.CurrentUserID(c)
	userID := atoui64(c.Param("id"))
	var req struct {
		Enable bool   `json:"enable"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	if err := h.Svc.TogglePromoter(userID, req.Enable, uint64(adminID), req.Reason); err != nil {
		FailErr(c, err)
		return
	}
	// 审计日志
	action := "disable_promoter"
	if req.Enable {
		action = "enable_promoter"
	}
	h.Svc.AdminRepo.CreateLog(&model.AdminActionLog{
		AdminID: uint64(adminID),
		Action:  action,
		Target:  c.Param("id"),
		Detail:  req.Reason,
		IP:      c.ClientIP(),
	})
	OKMsg(c, "操作成功")
}

// BanUser PUT /api/v1/admin/users/:id/status
func (h *AdminHandler) BanUser(c *gin.Context) {
	adminID := middleware.CurrentUserID(c)
	userID := atoui64(c.Param("id"))
	var req struct {
		Ban bool `json:"ban"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	if err := h.Svc.BanUser(userID, req.Ban, uint64(adminID)); err != nil {
		FailErr(c, err)
		return
	}
	h.Svc.AdminRepo.CreateLog(&model.AdminActionLog{
		AdminID: uint64(adminID),
		Action:  "ban_user",
		Target:  c.Param("id"),
		IP:      c.ClientIP(),
	})
	OKMsg(c, "操作成功")
}

// ListSettlements GET /api/v1/admin/settlements
func (h *AdminHandler) ListSettlements(c *gin.Context) {
	page, size := pageParams(c)
	list, total, err := h.Svc.FinanceRepo.ListCommissions(0, page, size)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, gin.H{"list": list, "total": total})
}

// ApproveSettlement POST /api/v1/admin/settlements/:id/approve
func (h *AdminHandler) ApproveSettlement(c *gin.Context) {
	adminID := middleware.CurrentUserID(c)
	id := atoui64(c.Param("id"))
	if err := h.Svc.ApproveCommission(id, uint64(adminID)); err != nil {
		FailErr(c, err)
		return
	}
	OKMsg(c, "审核通过")
}

// MarkSettlementPaid POST /api/v1/admin/settlements/:id/mark-paid
func (h *AdminHandler) MarkSettlementPaid(c *gin.Context) {
	id := atoui64(c.Param("id"))
	var req struct {
		TransferProof string `json:"transfer_proof"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.Svc.MarkCommissionPaid(id, req.TransferProof); err != nil {
		FailErr(c, err)
		return
	}
	OKMsg(c, "已标记打款")
}

// ListDeployments GET /api/v1/admin/deployments
func (h *AdminHandler) ListDeployments(c *gin.Context) {
	page, size := pageParams(c)
	status := c.Query("status")
	list, total, err := h.Svc.DeploymentRepo.ListDeployments(page, size, status)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, gin.H{"list": list, "total": total})
}
