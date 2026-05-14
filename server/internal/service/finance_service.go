package service

import (
	"fmt"
	"time"

	"fujiantong/internal/model"
)

// GenerateCommissionSettlement 生成某推广员某期佣金结算单
// 铁律 2：单层提成。铁律 5：主理人自营打标。
func (c *Container) GenerateCommissionSettlement(promoterID uint64, period string) (*model.CommissionSettlement, error) {
	promoter, err := c.UserRepo.GetByID(promoterID)
	if err != nil {
		return nil, fmt.Errorf("推广员不存在: %w", err)
	}
	if !promoter.IsPromoter {
		return nil, fmt.Errorf("用户 %d 不是推广员", promoterID)
	}

	// 统计名下作者在该期的平台分成总额
	platformFee, err := c.FinanceRepo.SumPlatformFeeByPromoterSubs(promoterID, period, c.DB)
	if err != nil {
		return nil, err
	}

	// 佣金 = 平台分成 × 30%
	commission := platformFee * 30 / 100

	cs := &model.CommissionSettlement{
		PromoterID:         promoterID,
		SettlePeriod:       period,
		TotalAuthorRevenue: 0,
		PlatformFee:        platformFee,
		Commission:         commission,
		IsSelfPromoter:     promoter.IsAdmin, // 铁律 5
		Status:             "pending",
	}
	if err := c.FinanceRepo.CreateCommission(cs); err != nil {
		return nil, err
	}
	return cs, nil
}

// ApproveCommission 审核通过结算单
func (c *Container) ApproveCommission(id, adminID uint64) error {
	cs, err := c.FinanceRepo.GetCommissionByID(id)
	if err != nil {
		return err
	}
	if cs.Status != "pending" {
		return fmt.Errorf("结算单状态非 pending，当前: %s", cs.Status)
	}
	cs.Status = "approved"
	cs.ApprovedBy = adminID
	return c.FinanceRepo.UpdateCommission(cs)
}

// MarkCommissionPaid 标记结算单已打款
func (c *Container) MarkCommissionPaid(id uint64, transferProof string) error {
	cs, err := c.FinanceRepo.GetCommissionByID(id)
	if err != nil {
		return err
	}
	if cs.Status != "approved" {
		return fmt.Errorf("结算单状态非 approved，当前: %s", cs.Status)
	}
	now := time.Now()
	cs.Status = "paid"
	cs.PaidAt = &now
	cs.TransferProof = transferProof
	return c.FinanceRepo.UpdateCommission(cs)
}

// GetDashboardStats 仪表盘数据
type DashboardStats struct {
	TotalAuthors        int64 `json:"total_authors"`
	TotalPromoters      int64 `json:"total_promoters"`
	MonthPlatformIncome int64 `json:"month_platform_income"`
	PendingCommissions  int   `json:"pending_commissions"`
}

func (c *Container) GetDashboardStats() (*DashboardStats, error) {
	var authorCount, promoterCount int64
	c.DB.Model(&model.User{}).Where("status = 1 AND bound_appid != ''").Count(&authorCount)
	c.DB.Model(&model.User{}).Where("status = 1 AND is_promoter = true").Count(&promoterCount)

	// 本月平台收入
	now := time.Now()
	monthStart := fmt.Sprintf("%d-%02d-01", now.Year(), now.Month())
	monthEnd := fmt.Sprintf("%d-%02d-31", now.Year(), now.Month())
	var monthIncome int64
	c.DB.Model(&model.SettlementRecord{}).
		Where("settle_period >= ? AND settle_period <= ?", monthStart, monthEnd).
		Select("COALESCE(SUM(platform_share), 0)").Scan(&monthIncome)

	pending, _ := c.FinanceRepo.ListPendingCommissions()

	return &DashboardStats{
		TotalAuthors:        authorCount,
		TotalPromoters:      promoterCount,
		MonthPlatformIncome: monthIncome,
		PendingCommissions:  len(pending),
	}, nil
}
