package repo

import (
	"fujiantong/internal/model"

	"gorm.io/gorm"
)

type FileRepo struct{ *Repo }

func NewFileRepo(db *gorm.DB) *FileRepo { return &FileRepo{New(db)} }

func (r *FileRepo) Create(f *model.File) error {
	return r.DB.Create(f).Error
}

func (r *FileRepo) GetByID(id uint64) (*model.File, error) {
	var f model.File
	err := r.DB.First(&f, id).Error
	return &f, err
}

func (r *FileRepo) GetByDownloadCode(code string) (*model.File, error) {
	var f model.File
	err := r.DB.Where("download_code = ? AND status = 'active'", code).First(&f).Error
	return &f, err
}

func (r *FileRepo) Update(f *model.File) error {
	return r.DB.Save(f).Error
}

func (r *FileRepo) UpdateFields(id uint64, fields map[string]any) error {
	return r.DB.Model(&model.File{}).Where("id = ?", id).Updates(fields).Error
}

func (r *FileRepo) IncrViewCount(id uint64) error {
	return r.DB.Model(&model.File{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *FileRepo) IncrDownloadCount(id uint64) error {
	return r.DB.Model(&model.File{}).Where("id = ?", id).
		UpdateColumn("download_count", gorm.Expr("download_count + 1")).Error
}

// ListByAuthor 作者附件列表（分页）
func (r *FileRepo) ListByAuthor(authorUserID uint64, page, size int) ([]model.File, int64, error) {
	var list []model.File
	var total int64
	q := r.DB.Model(&model.File{}).Where("author_user_id = ? AND deleted_at IS NULL", authorUserID)
	q.Count(&total)
	err := q.Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

// ListAll 管理员查全部文件
func (r *FileRepo) ListAll(page, size int, status string) ([]model.File, int64, error) {
	var list []model.File
	var total int64
	q := r.DB.Model(&model.File{}).Where("deleted_at IS NULL")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)
	err := q.Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

// ─── FileShareCard ───

func (r *FileRepo) CreateShareCard(card *model.FileShareCard) error {
	return r.DB.Create(card).Error
}

func (r *FileRepo) GetShareCards(fileID uint64) ([]model.FileShareCard, error) {
	var list []model.FileShareCard
	err := r.DB.Where("file_id = ?", fileID).Find(&list).Error
	return list, err
}

func (r *FileRepo) UpsertShareCard(card *model.FileShareCard) error {
	var existing model.FileShareCard
	err := r.DB.Where("file_id = ? AND card_type = ?", card.FileID, card.CardType).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.DB.Create(card).Error
	}
	if err != nil {
		return err
	}
	existing.HTML = card.HTML
	existing.AppID = card.AppID
	return r.DB.Save(&existing).Error
}
