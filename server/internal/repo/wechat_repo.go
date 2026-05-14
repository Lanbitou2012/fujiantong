package repo

import (
	"fujiantong/internal/model"
	"time"

	"gorm.io/gorm"
)

type WechatRepo struct{ *Repo }

func NewWechatRepo(db *gorm.DB) *WechatRepo { return &WechatRepo{New(db)} }

// ─── WxToken 中央 Token 存储 ───

func (r *WechatRepo) GetToken(key string) (*model.WxToken, error) {
	var t model.WxToken
	err := r.DB.Where("`key` = ?", key).First(&t).Error
	return &t, err
}

func (r *WechatRepo) UpsertToken(t *model.WxToken) error {
	var existing model.WxToken
	err := r.DB.Where("`key` = ?", t.Key).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.DB.Create(t).Error
	}
	if err != nil {
		return err
	}
	return r.DB.Model(&existing).Updates(map[string]any{
		"access_token":  t.AccessToken,
		"refresh_token": t.RefreshToken,
		"expires_at":    t.ExpiresAt,
		"extra":         t.Extra,
	}).Error
}

// IsTokenExpired 检查 token 是否已过期或即将过期（提前 5 分钟）
func (r *WechatRepo) IsTokenExpired(key string) (bool, error) {
	t, err := r.GetToken(key)
	if err == gorm.ErrRecordNotFound {
		return true, nil
	}
	if err != nil {
		return true, err
	}
	return t.ExpiresAt.Before(time.Now().Add(5 * time.Minute)), nil
}

// SaveVerifyTicket 保存 component_verify_ticket
func (r *WechatRepo) SaveVerifyTicket(ticket string) error {
	return r.UpsertToken(&model.WxToken{
		Key:         "component_verify_ticket",
		AccessToken: ticket,
		ExpiresAt:   time.Now().Add(12 * time.Hour),
	})
}

func (r *WechatRepo) GetVerifyTicket() (string, error) {
	t, err := r.GetToken("component_verify_ticket")
	if err != nil {
		return "", err
	}
	return t.AccessToken, nil
}
