package service

import (
	"errors"
	"fmt"

	"fujiantong/internal/model"

	"gorm.io/gorm"
)

// TogglePromoter 铁律 7：开关推广员身份 + 写入历史表
func (c *Container) TogglePromoter(userID uint64, enable bool, adminID uint64, reason string) error {
	u, err := c.UserRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if u.IsPromoter == enable {
		return nil // 无变化
	}

	// 更新主表
	if err := c.UserRepo.UpdateFields(userID, map[string]any{"is_promoter": enable}); err != nil {
		return err
	}

	// 铁律 7：写入角色变更历史
	return c.UserRepo.InsertRoleHistory(&model.UserRoleHistory{
		UserID:         userID,
		Field:          "is_promoter",
		OldValue:       fmt.Sprintf("%v", u.IsPromoter),
		NewValue:       fmt.Sprintf("%v", enable),
		ChangedByAdmin: adminID,
		Reason:         reason,
	})
}

// ToggleAdmin 管理员身份切换 + 历史记录
func (c *Container) ToggleAdmin(userID uint64, enable bool, adminID uint64, reason string) error {
	u, err := c.UserRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if u.IsAdmin == enable {
		return nil
	}
	if err := c.UserRepo.UpdateFields(userID, map[string]any{"is_admin": enable}); err != nil {
		return err
	}
	return c.UserRepo.InsertRoleHistory(&model.UserRoleHistory{
		UserID:         userID,
		Field:          "is_admin",
		OldValue:       fmt.Sprintf("%v", u.IsAdmin),
		NewValue:       fmt.Sprintf("%v", enable),
		ChangedByAdmin: adminID,
		Reason:         reason,
	})
}

// BanUser 封禁/解封用户
func (c *Container) BanUser(userID uint64, ban bool, adminID uint64) error {
	status := int8(1)
	if ban {
		status = 0
	}
	return c.UserRepo.UpdateFields(userID, map[string]any{"status": status})
}

// SubmitAppID 作者回填 appid
func (c *Container) SubmitAppID(userID uint64, appID string) error {
	if appID == "" {
		return errors.New("appid 不能为空")
	}
	// 检查 appid 是否已被占用
	existing, err := c.UserRepo.GetByAppID(appID)
	if err == nil && existing.ID != userID {
		return errors.New("该 appid 已被其他用户绑定")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return c.UserRepo.UpdateFields(userID, map[string]any{"bound_appid": appID})
}

// GetUserProfile 获取用户资料
func (c *Container) GetUserProfile(userID uint64) (*model.User, error) {
	return c.UserRepo.GetByID(userID)
}

// RealtimeCheckPromoter 铁律 4：实时校验推广员权限（中间件调用）
func (c *Container) RealtimeCheckPromoter(userID uint64) (bool, error) {
	u, err := c.UserRepo.GetByID(userID)
	if err != nil {
		return false, err
	}
	return u.IsPromoter && u.Status == 1, nil
}

// RealtimeCheckAdmin 铁律 4：实时校验管理员权限
func (c *Container) RealtimeCheckAdmin(userID uint64) (bool, error) {
	u, err := c.UserRepo.GetByID(userID)
	if err != nil {
		return false, err
	}
	return u.IsAdmin && u.Status == 1, nil
}
