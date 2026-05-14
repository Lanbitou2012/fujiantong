package model

import "time"

// DailyAdStat 每个授权小程序每日广告数据（从 api_getadposdetail 拉取）。
type DailyAdStat struct {
	ID    uint64 `gorm:"primaryKey" json:"id"`
	AppID string `gorm:"index;type:varchar(64);not null" json:"appid"`
	Date  string `gorm:"index;type:varchar(10);not null" json:"date"` // YYYY-MM-DD

	// 各广告位数据
	SlotID        string  `gorm:"type:varchar(64);not null;default:''" json:"slot_id"`
	AdSlot        string  `gorm:"type:varchar(64);not null;default:''" json:"ad_slot"` // rewarded_video / interstitial / splash
	ExposureCount int64   `gorm:"not null;default:0" json:"exposure_count"`
	ClickCount    int64   `gorm:"not null;default:0" json:"click_count"`
	Income        int64   `gorm:"not null;default:0" json:"income"` // 单位：分
	ECPM          float64 `gorm:"not null;default:0" json:"ecpm"`

	CreatedAt time.Time `json:"created_at"`
}

func (DailyAdStat) TableName() string { return "daily_ad_stats" }

// SettlementRecord 每个授权小程序每期结算数据（从 api_getsettlement 拉取）。
// 每半月一次（次月 1 号、15 号前发结算单）。
type SettlementRecord struct {
	ID    uint64 `gorm:"primaryKey" json:"id"`
	AppID string `gorm:"index;type:varchar(64);not null" json:"appid"`

	// 结算期标识，如 "2026-05-01" 或 "2026-05-15"
	SettlePeriod string `gorm:"index;type:varchar(20);not null" json:"settle_period"`

	// 金额（单位：分）
	TotalRevenue  int64 `gorm:"not null;default:0" json:"total_revenue"`   // 广告变现总收入
	PlatformShare int64 `gorm:"not null;default:0" json:"platform_share"`  // 平台（28%）
	AuthorShare   int64 `gorm:"not null;default:0" json:"author_share"`    // 作者（72%）

	// 关联作者
	UserID uint64 `gorm:"index;not null;default:0" json:"user_id"`

	// 原始数据 JSON（方便对账）
	RawJSON string `gorm:"type:longtext" json:"raw_json"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (SettlementRecord) TableName() string { return "settlement_records" }

// CommissionSettlement 推广员佣金结算单。
// 佣金 = 名下作者流量主毛收益 × 28% × 30%。
// 铁律 5：主理人自营打标 is_self_promoter=true。
type CommissionSettlement struct {
	ID         uint64 `gorm:"primaryKey" json:"id"`
	PromoterID uint64 `gorm:"index;not null" json:"promoter_id"` // 推广员 user_id

	// 结算期
	SettlePeriod string `gorm:"index;type:varchar(20);not null" json:"settle_period"`

	// 金额（单位：分）
	TotalAuthorRevenue int64 `gorm:"not null;default:0" json:"total_author_revenue"` // 名下作者毛收入总和
	PlatformFee        int64 `gorm:"not null;default:0" json:"platform_fee"`         // 平台 28% 总和
	Commission         int64 `gorm:"not null;default:0" json:"commission"`            // 推广员应得（platform_fee × 30%）

	// 铁律 5：主理人自营标记
	IsSelfPromoter bool `gorm:"not null;default:false" json:"is_self_promoter"`

	// 状态机：pending → approved → paid / rejected
	Status string `gorm:"type:varchar(32);not null;default:'pending'" json:"status"`

	// 打款信息
	TransferProof string     `gorm:"type:varchar(512);not null;default:''" json:"transfer_proof"` // 凭证图 URL
	PaidAt        *time.Time `json:"paid_at"`
	ApprovedBy    uint64     `gorm:"not null;default:0" json:"approved_by"` // 审核管理员 user_id

	// 明细 JSON（各下级作者的贡献详情）
	DetailJSON string `gorm:"type:longtext" json:"detail_json"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (CommissionSettlement) TableName() string { return "commission_settlements" }
