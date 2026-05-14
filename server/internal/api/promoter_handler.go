package api

import (
	"fujiantong/internal/middleware"
	"fujiantong/internal/service"

	"github.com/gin-gonic/gin"
)

type PromoterHandler struct {
	Svc *service.Container
}

func NewPromoterHandler(svc *service.Container) *PromoterHandler {
	return &PromoterHandler{Svc: svc}
}

// ListAuthors GET /api/v1/promoter/authors
func (h *PromoterHandler) ListAuthors(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	page, size := pageParams(c)
	list, total, err := h.Svc.UserRepo.ListByPromoter(uint64(uid), page, size)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, gin.H{"list": list, "total": total})
}

// GetCommission GET /api/v1/promoter/commission
func (h *PromoterHandler) GetCommission(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	page, size := pageParams(c)
	list, total, err := h.Svc.FinanceRepo.ListCommissions(uint64(uid), page, size)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, gin.H{"list": list, "total": total})
}

// UpdatePaymentInfo PUT /api/v1/promoter/payment-info
func (h *PromoterHandler) UpdatePaymentInfo(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var req struct {
		PaymentInfo string `json:"payment_info" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	if err := h.Svc.UserRepo.UpdateFields(uint64(uid), map[string]any{
		"payment_info": req.PaymentInfo,
	}); err != nil {
		FailErr(c, err)
		return
	}
	OKMsg(c, "收款信息已更新")
}
