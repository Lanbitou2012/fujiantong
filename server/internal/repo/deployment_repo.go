package repo

import (
	"fujiantong/internal/model"

	"gorm.io/gorm"
)

type DeploymentRepo struct{ *Repo }

func NewDeploymentRepo(db *gorm.DB) *DeploymentRepo { return &DeploymentRepo{New(db)} }

// ─── MpTemplateVersion ───

func (r *DeploymentRepo) CreateTemplateVersion(v *model.MpTemplateVersion) error {
	return r.DB.Create(v).Error
}

func (r *DeploymentRepo) GetActiveTemplate() (*model.MpTemplateVersion, error) {
	var v model.MpTemplateVersion
	err := r.DB.Where("status = 'active'").Order("id DESC").First(&v).Error
	return &v, err
}

func (r *DeploymentRepo) ListTemplateVersions() ([]model.MpTemplateVersion, error) {
	var list []model.MpTemplateVersion
	err := r.DB.Order("id DESC").Find(&list).Error
	return list, err
}

func (r *DeploymentRepo) UpdateTemplateVersion(v *model.MpTemplateVersion) error {
	return r.DB.Save(v).Error
}

// DeprecateAllTemplates 弃用所有模板（发新版前调用）
func (r *DeploymentRepo) DeprecateAllTemplates() error {
	return r.DB.Model(&model.MpTemplateVersion{}).Where("status = 'active'").
		Update("status", "deprecated").Error
}

// ─── MpDeployment ───

func (r *DeploymentRepo) CreateDeployment(d *model.MpDeployment) error {
	return r.DB.Create(d).Error
}

func (r *DeploymentRepo) GetDeploymentByID(id uint64) (*model.MpDeployment, error) {
	var d model.MpDeployment
	err := r.DB.First(&d, id).Error
	return &d, err
}

func (r *DeploymentRepo) GetLatestDeployment(appID string) (*model.MpDeployment, error) {
	var d model.MpDeployment
	err := r.DB.Where("appid = ?", appID).Order("id DESC").First(&d).Error
	return &d, err
}

func (r *DeploymentRepo) UpdateDeployment(d *model.MpDeployment) error {
	return r.DB.Save(d).Error
}

func (r *DeploymentRepo) UpdateDeploymentStatus(id uint64, status string) error {
	return r.DB.Model(&model.MpDeployment{}).Where("id = ?", id).Update("status", status).Error
}

// ListDeployments 管理员查看部署列表
func (r *DeploymentRepo) ListDeployments(page, size int, status string) ([]model.MpDeployment, int64, error) {
	var list []model.MpDeployment
	var total int64
	q := r.DB.Model(&model.MpDeployment{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)
	err := q.Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

// CountByStatus 部署状态总览
func (r *DeploymentRepo) CountByStatus() (map[string]int64, error) {
	type Row struct {
		Status string
		Cnt    int64
	}
	var rows []Row
	err := r.DB.Model(&model.MpDeployment{}).
		Select("status, COUNT(*) as cnt").Group("status").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]int64)
	for _, row := range rows {
		m[row.Status] = row.Cnt
	}
	return m, nil
}

// ─── MpAudit ───

func (r *DeploymentRepo) CreateAudit(a *model.MpAudit) error {
	return r.DB.Create(a).Error
}

func (r *DeploymentRepo) GetAuditByWxID(wxAuditID int64) (*model.MpAudit, error) {
	var a model.MpAudit
	err := r.DB.Where("wx_audit_id = ?", wxAuditID).First(&a).Error
	return &a, err
}

func (r *DeploymentRepo) GetAuditByDeployment(deploymentID uint64) (*model.MpAudit, error) {
	var a model.MpAudit
	err := r.DB.Where("deployment_id = ?", deploymentID).Order("id DESC").First(&a).Error
	return &a, err
}

func (r *DeploymentRepo) UpdateAudit(a *model.MpAudit) error {
	return r.DB.Save(a).Error
}
