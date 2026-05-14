package repo

import (
	"fujiantong/internal/model"

	"gorm.io/gorm"
)

type UserRepo struct{ *Repo }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{New(db)} }

func (r *UserRepo) Create(u *model.User) error {
	return r.DB.Create(u).Error
}

func (r *UserRepo) GetByID(id uint64) (*model.User, error) {
	var u model.User
	err := r.DB.First(&u, id).Error
	return &u, err
}

func (r *UserRepo) GetByUnionID(unionID string) (*model.User, error) {
	var u model.User
	err := r.DB.Where("wx_union_id = ?", unionID).First(&u).Error
	return &u, err
}

func (r *UserRepo) GetByAppID(appID string) (*model.User, error) {
	var u model.User
	err := r.DB.Where("bound_appid = ?", appID).First(&u).Error
	return &u, err
}

func (r *UserRepo) GetByAdminUsername(username string) (*model.User, error) {
	var u model.User
	err := r.DB.Where("admin_username = ? AND is_admin = ?", username, true).First(&u).Error
	return &u, err
}

func (r *UserRepo) Update(u *model.User) error {
	return r.DB.Save(u).Error
}

func (r *UserRepo) UpdateFields(id uint64, fields map[string]any) error {
	return r.DB.Model(&model.User{}).Where("id = ?", id).Updates(fields).Error
}

// ListByPromoter 查询某推广员名下所有作者
func (r *UserRepo) ListByPromoter(promoterID uint64, page, size int) ([]model.User, int64, error) {
	var list []model.User
	var total int64
	q := r.DB.Model(&model.User{}).Where("parent_promoter_user_id = ? AND status = 1", promoterID)
	q.Count(&total)
	err := q.Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

// ListAll 分页查询全部用户（管理员用）
func (r *UserRepo) ListAll(page, size int, keyword string) ([]model.User, int64, error) {
	var list []model.User
	var total int64
	q := r.DB.Model(&model.User{}).Where("status >= 0")
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("nickname LIKE ? OR bound_appid LIKE ? OR wx_union_id LIKE ?", like, like, like)
	}
	q.Count(&total)
	err := q.Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

// CheckCircular 铁律 3：递归向上 walk 检测是否会形成循环
func (r *UserRepo) CheckCircular(userID, newParentID uint64) (bool, error) {
	if newParentID == 0 || userID == newParentID {
		return userID == newParentID && newParentID != 0, nil
	}
	visited := map[uint64]bool{userID: true}
	current := newParentID
	for current != 0 {
		if visited[current] {
			return true, nil
		}
		visited[current] = true
		var u model.User
		if err := r.DB.Select("parent_promoter_user_id").First(&u, current).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return false, nil
			}
			return false, err
		}
		current = u.ParentPromoterUserID
	}
	return false, nil
}

// InsertRoleHistory 铁律 7：角色变更写入历史表
func (r *UserRepo) InsertRoleHistory(h *model.UserRoleHistory) error {
	return r.DB.Create(h).Error
}

// IsPromoterAt 铁律 1 支撑：查某用户在某日是否是推广员
func (r *UserRepo) IsPromoterAt(userID uint64, date string) (bool, error) {
	// 找最近一条 created_at <= date 的 is_promoter 变更记录
	var h model.UserRoleHistory
	err := r.DB.Where("user_id = ? AND field = 'is_promoter' AND created_at <= ?", userID, date+" 23:59:59").
		Order("created_at DESC").First(&h).Error
	if err == gorm.ErrRecordNotFound {
		// 没有变更记录，查主表当前值（说明从未变更过）
		var u model.User
		if e2 := r.DB.Select("is_promoter").First(&u, userID).Error; e2 != nil {
			return false, e2
		}
		return u.IsPromoter, nil
	}
	if err != nil {
		return false, err
	}
	return h.NewValue == "true", nil
}
