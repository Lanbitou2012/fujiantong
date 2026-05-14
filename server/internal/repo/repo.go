package repo

import "gorm.io/gorm"

// Repo 统一持有 *gorm.DB，所有子 repo 组合此字段。
type Repo struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *Repo {
	return &Repo{DB: db}
}
