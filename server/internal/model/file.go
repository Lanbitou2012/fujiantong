package model

import "time"

// File 附件元信息表。
// 所有业务表预留 authorizer_appid 作为租户字段，为分库分表做准备。
type File struct {
	ID           uint64 `gorm:"primaryKey" json:"id"`
	AuthorUserID uint64 `gorm:"index;not null" json:"author_user_id"`
	AppID        string `gorm:"index;type:varchar(64);not null;default:''" json:"appid"`

	// 文件基本信息
	Name         string `gorm:"type:varchar(255);not null" json:"name"`
	OriginalName string `gorm:"type:varchar(255);not null" json:"original_name"`
	Ext          string `gorm:"type:varchar(20);not null" json:"ext"`
	Size         int64  `gorm:"not null;default:0" json:"size"`
	MimeType     string `gorm:"type:varchar(120);not null;default:''" json:"mime_type"`

	// 存储
	StorageType string `gorm:"type:varchar(32);not null;default:'local'" json:"storage_type"` // local / cos
	StoragePath string `gorm:"type:varchar(512);not null" json:"storage_path"`

	// 下载码（短链接码）
	DownloadCode string `gorm:"uniqueIndex;type:varchar(32);not null" json:"download_code"`

	// 统计
	ViewCount     uint64 `gorm:"not null;default:0" json:"view_count"`
	DownloadCount uint64 `gorm:"not null;default:0" json:"download_count"`

	// 内容审核
	MediaCheckStatus  string `gorm:"type:varchar(32);not null;default:'pending'" json:"media_check_status"` // pending / pass / reject / reviewing
	MediaCheckTraceID string `gorm:"type:varchar(128);not null;default:''" json:"media_check_trace_id"`

	// 状态：active / offline / banned
	Status string `gorm:"index;type:varchar(32);not null;default:'active'" json:"status"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at"`
}

func (File) TableName() string { return "files" }

// FileShareCard 生成的复制内容快照（文本超链接 / 小程序卡片 / H5 备份三种 HTML 形式）。
type FileShareCard struct {
	ID       uint64 `gorm:"primaryKey" json:"id"`
	FileID   uint64 `gorm:"index;not null" json:"file_id"`
	CardType string `gorm:"type:varchar(32);not null" json:"card_type"` // weapp_text_link / mp_miniprogram / h5_backup / miniprogram_path
	HTML     string `gorm:"type:longtext;not null" json:"html"`
	AppID    string `gorm:"index;type:varchar(64);not null;default:''" json:"appid"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (FileShareCard) TableName() string { return "file_share_cards" }
