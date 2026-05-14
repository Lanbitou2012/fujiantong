package model

import "time"

// MpTemplateVersion 模板小程序代码版本管理。
// 模板代码 1 套、配置 N 套。模板 bug 修复一次推送全体。
type MpTemplateVersion struct {
	ID          uint64 `gorm:"primaryKey" json:"id"`
	TemplateID  int64  `gorm:"not null" json:"template_id"` // 微信开放平台的 template_id
	UserVersion string `gorm:"type:varchar(64);not null" json:"user_version"`
	UserDesc    string `gorm:"type:varchar(512);not null;default:''" json:"user_desc"`

	// 发布状态：draft / active / deprecated
	Status string `gorm:"type:varchar(32);not null;default:'draft'" json:"status"`

	// 变更日志
	ChangeLog string `gorm:"type:text" json:"change_log"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (MpTemplateVersion) TableName() string { return "mp_template_versions" }

// MpDeployment 作者小程序部署记录。
// 部署状态机见 CORE_ARCHITECTURE.md 15.2.8。
type MpDeployment struct {
	ID     uint64 `gorm:"primaryKey" json:"id"`
	UserID uint64 `gorm:"index;not null" json:"user_id"`
	AppID  string `gorm:"index;type:varchar(64);not null" json:"appid"`

	// 使用的模板版本
	TemplateVersionID uint64 `gorm:"not null;default:0" json:"template_version_id"`
	TemplateID        int64  `gorm:"not null;default:0" json:"template_id"`
	UserVersion       string `gorm:"type:varchar(64);not null;default:''" json:"user_version"`

	// ext_json 快照
	ExtJSON string `gorm:"type:longtext" json:"ext_json"`

	// 部署状态机
	// pending → configuring → committing → submitted_for_audit → ready_to_release → live
	// submitted_for_audit → audit_failed → waiting_template_fix / waiting_author_supplement / manual_review
	// live → committing（升级流水）
	Status string `gorm:"type:varchar(64);not null;default:'pending';index" json:"status"`

	// 阶段 1 配置结果
	ConfigResult string `gorm:"type:text" json:"config_result"`

	// 提审相关
	AuditID    int64  `gorm:"not null;default:0" json:"audit_id"` // 微信返回的 auditid
	FailReason string `gorm:"type:text" json:"fail_reason"`

	// 事件流日志（JSON 数组，每个事件一个对象）
	EventLog string `gorm:"type:longtext" json:"event_log"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (MpDeployment) TableName() string { return "mp_deployments" }

// MpAudit 提审记录。
// 每次 submit_audit 生成一条，审核结果由事件推送更新。
type MpAudit struct {
	ID     uint64 `gorm:"primaryKey" json:"id"`
	AppID  string `gorm:"index;type:varchar(64);not null" json:"appid"`
	UserID uint64 `gorm:"index;not null" json:"user_id"`

	DeploymentID uint64 `gorm:"index;not null" json:"deployment_id"`

	// 微信返回的 auditid
	WxAuditID int64 `gorm:"not null;default:0" json:"wx_audit_id"`

	// 提审参数快照
	VersionDesc  string `gorm:"type:varchar(512);not null;default:''" json:"version_desc"`
	PreviewInfo  string `gorm:"type:text" json:"preview_info"`
	UGCDeclare   string `gorm:"type:text" json:"ugc_declare"`

	// 审核状态：submitted / success / fail / delay / undone
	Status string `gorm:"type:varchar(32);not null;default:'submitted'" json:"status"`

	// 驳回详情
	FailReason    string `gorm:"type:text" json:"fail_reason"`
	FailCategory  string `gorm:"type:varchar(64);not null;default:''" json:"fail_category"` // code / qualification / unknown
	RetryCount    int    `gorm:"not null;default:0" json:"retry_count"`

	SubmittedAt time.Time  `json:"submitted_at"`
	ResultAt    *time.Time `json:"result_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (MpAudit) TableName() string { return "mp_audits" }
