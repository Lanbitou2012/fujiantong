package repo

import (
	"fujiantong/internal/model"

	"gorm.io/gorm"
)

type AuthorizationRepo struct{ *Repo }

func NewAuthorizationRepo(db *gorm.DB) *AuthorizationRepo { return &AuthorizationRepo{New(db)} }

func (r *AuthorizationRepo) Upsert(a *model.Authorization) error {
	var existing model.Authorization
	err := r.DB.Where("app_id = ?", a.AppID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.DB.Create(a).Error
	}
	if err != nil {
		return err
	}
	a.ID = existing.ID
	return r.DB.Save(a).Error
}

func (r *AuthorizationRepo) GetByAppID(appID string) (*model.Authorization, error) {
	var a model.Authorization
	err := r.DB.Where("app_id = ?", appID).First(&a).Error
	return &a, err
}

func (r *AuthorizationRepo) GetByUserID(userID uint64) (*model.Authorization, error) {
	var a model.Authorization
	err := r.DB.Where("user_id = ?", userID).First(&a).Error
	return &a, err
}

func (r *AuthorizationRepo) Update(a *model.Authorization) error {
	return r.DB.Save(a).Error
}

func (r *AuthorizationRepo) ListAll(page, size int) ([]model.Authorization, int64, error) {
	var list []model.Authorization
	var total int64
	q := r.DB.Model(&model.Authorization{})
	q.Count(&total)
	err := q.Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

func (r *AuthorizationRepo) ListAuthorized() ([]model.Authorization, error) {
	var list []model.Authorization
	err := r.DB.Where("status = 'authorized'").Find(&list).Error
	return list, err
}
