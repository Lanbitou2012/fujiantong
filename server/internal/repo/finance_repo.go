package repo

import (
	"fujiantong/internal/model"

	"gorm.io/gorm"
)

type FinanceRepo struct{ *Repo }

func NewFinanceRepo(db *gorm.DB) *FinanceRepo { return &FinanceRepo{New(db)} }

// ─── SettlementRecord ───

func (r *FinanceRepo) CreateSettlement(s *model.SettlementRecord) error {
	return r.DB.Create(s).Error
}

func (r *FinanceRepo) BatchCreateSettlements(records []model.SettlementRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.DB.CreateInBatches(records, 50).Error
}

func (r *FinanceRepo) GetSettlements(appID string, page, size int) ([]model.SettlementRecord, int64, error) {
	var list []model.SettlementRecord
	var total int64
	q := r.DB.Model(&model.SettlementRecord{}).Where("appid = ?", appID)
	q.Count(&total)
	err := q.Offset((page - 1) * size).Limit(size).Order("settle_period DESC").Find(&list).Error
	return list, total, err
}

// ─── CommissionSettlement ───

func (r *FinanceRepo) CreateCommission(c *model.CommissionSettlement) error {
	return r.DB.Create(c).Error
}

func (r *FinanceRepo) GetCommissionByID(id uint64) (*model.CommissionSettlement, error) {
	var c model.CommissionSettlement
	err := r.DB.First(&c, id).Error
	return &c, err
}

func (r *FinanceRepo) UpdateCommission(c *model.CommissionSettlement) error {
	return r.DB.Save(c).Error
}

func (r *FinanceRepo) ListCommissions(promoterID uint64, page, size int) ([]model.CommissionSettlement, int64, error) {
	var list []model.CommissionSettlement
	var total int64
	q := r.DB.Model(&model.CommissionSettlement{})
	if promoterID > 0 {
		q = q.Where("promoter_id = ?", promoterID)
	}
	q.Count(&total)
	err := q.Offset((page - 1) * size).Limit(size).Order("settle_period DESC").Find(&list).Error
	return list, total, err
}

func (r *FinanceRepo) ListPendingCommissions() ([]model.CommissionSettlement, error) {
	var list []model.CommissionSettlement
	err := r.DB.Where("status IN ('pending','approved')").Order("id ASC").Find(&list).Error
	return list, err
}

// SumPlatformFeeByPromoterSubs 统计某推广员名下所有作者在某期的平台分成总额
func (r *FinanceRepo) SumPlatformFeeByPromoterSubs(promoterID uint64, period string, db *gorm.DB) (int64, error) {
	var total int64
	subQ := db.Model(&model.User{}).Select("bound_appid").Where("parent_promoter_user_id = ? AND status = 1 AND bound_appid != ''", promoterID)
	err := db.Model(&model.SettlementRecord{}).
		Where("settle_period = ? AND appid IN (?)", period, subQ).
		Select("COALESCE(SUM(platform_share), 0)").Scan(&total).Error
	return total, err
}
