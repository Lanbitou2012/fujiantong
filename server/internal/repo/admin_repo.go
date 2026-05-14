package repo

import (
	"fujiantong/internal/model"

	"gorm.io/gorm"
)

type AdminRepo struct{ *Repo }

func NewAdminRepo(db *gorm.DB) *AdminRepo { return &AdminRepo{New(db)} }

func (r *AdminRepo) CreateLog(log *model.AdminActionLog) error {
	return r.DB.Create(log).Error
}

func (r *AdminRepo) ListLogs(page, size int, action string) ([]model.AdminActionLog, int64, error) {
	var list []model.AdminActionLog
	var total int64
	q := r.DB.Model(&model.AdminActionLog{})
	if action != "" {
		q = q.Where("action = ?", action)
	}
	q.Count(&total)
	err := q.Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}
