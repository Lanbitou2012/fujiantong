package repo

import (
	"os"
	"sync"

	"fujiantong/internal/model"

	"gorm.io/gorm"
)

// SettingRepo 提供 system_settings 的 CRUD 及内存缓存。
// 读取优先级：内存缓存 → DB → 环境变量。
type SettingRepo struct {
	*Repo
	mu    sync.RWMutex
	cache map[string]string
}

func NewSettingRepo(db *gorm.DB) *SettingRepo {
	r := &SettingRepo{Repo: New(db), cache: make(map[string]string)}
	r.reload()
	return r
}

// reload 从 DB 全量加载到内存缓存
func (r *SettingRepo) reload() {
	if r.DB == nil {
		return
	}
	var list []model.SystemSetting
	if err := r.DB.Find(&list).Error; err != nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string]string, len(list))
	for _, s := range list {
		r.cache[s.Key] = s.Value
	}
}

// Get 优先返回 DB 配置，缺省回退到环境变量
func (r *SettingRepo) Get(key, envKey string) string {
	r.mu.RLock()
	v, ok := r.cache[key]
	r.mu.RUnlock()
	if ok && v != "" {
		return v
	}
	if envKey != "" {
		return os.Getenv(envKey)
	}
	return ""
}

// GetRaw 仅返回 DB 中的值，不回退环境变量
func (r *SettingRepo) GetRaw(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cache[key]
}

// GetAll 按分类返回，默认空字符串
func (r *SettingRepo) GetAll(keys []string) map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		out[k] = r.cache[k]
	}
	return out
}

// Set 写入并刷新缓存
func (r *SettingRepo) Set(key, value, category string, adminID uint64) error {
	if r.DB == nil {
		return nil
	}
	row := model.SystemSetting{Key: key, Value: value, Category: category, UpdatedBy: adminID}
	err := r.DB.Save(&row).Error
	if err == nil {
		r.mu.Lock()
		r.cache[key] = value
		r.mu.Unlock()
	}
	return err
}

// SetMany 批量写入
func (r *SettingRepo) SetMany(items map[string]string, category string, adminID uint64) error {
	if r.DB == nil {
		return nil
	}
	tx := r.DB.Begin()
	for k, v := range items {
		if err := tx.Save(&model.SystemSetting{Key: k, Value: v, Category: category, UpdatedBy: adminID}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	r.mu.Lock()
	for k, v := range items {
		r.cache[k] = v
	}
	r.mu.Unlock()
	return nil
}
