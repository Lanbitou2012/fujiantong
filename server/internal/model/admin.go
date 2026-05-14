package model

import "time"

// AdminActionLog 主理人关键操作审计日志。
type AdminActionLog struct {
	ID       uint64 `gorm:"primaryKey" json:"id"`
	AdminID  uint64 `gorm:"index;not null" json:"admin_id"`
	Action   string `gorm:"index;type:varchar(64);not null" json:"action"`
	Target   string `gorm:"type:varchar(128);not null;default:''" json:"target"`
	Detail   string `gorm:"type:longtext" json:"detail"`
	IP       string `gorm:"type:varchar(64);not null;default:''" json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

func (AdminActionLog) TableName() string { return "admin_action_log" }
