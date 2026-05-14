package service

import (
	"errors"
	"fmt"
	"time"

	"fujiantong/internal/model"
	jwtpkg "fujiantong/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminLogin 管理员账密登录
func (c *Container) AdminLogin(username, password string) (string, *model.User, error) {
	u, err := c.UserRepo.GetByAdminUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("用户名或密码错误")
		}
		return "", nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.LoginPassphraseHash), []byte(password)); err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}
	if u.Status != 1 {
		return "", nil, errors.New("账号已被封禁")
	}
	token, err := jwtpkg.GenerateTokenWithOptions(jwtpkg.TokenOptions{
		UserID:   uint(u.ID),
		Role:     roleFlag(u),
		RoleName: roleName(u),
		TTL:      7 * 24 * time.Hour,
	})
	return token, u, err
}

// WechatScanLogin 微信扫码登录（code 换 token）
func (c *Container) WechatScanLogin(unionID, openID, nickname, avatar string) (string, *model.User, error) {
	u, err := c.UserRepo.GetByUnionID(unionID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 首次登录，自动创建作者账号
		u = &model.User{
			WxUnionID:   unionID,
			WxOpenIDWeb: openID,
			Nickname:    nickname,
			AvatarURL:   avatar,
			Status:      1,
		}
		if err := c.UserRepo.Create(u); err != nil {
			return "", nil, fmt.Errorf("创建用户失败: %w", err)
		}
	} else if err != nil {
		return "", nil, err
	} else {
		// 更新头像昵称
		if nickname != "" && nickname != u.Nickname {
			u.Nickname = nickname
		}
		if avatar != "" && avatar != u.AvatarURL {
			u.AvatarURL = avatar
		}
		if openID != "" && openID != u.WxOpenIDWeb {
			u.WxOpenIDWeb = openID
		}
		_ = c.UserRepo.Update(u)
	}
	if u.Status != 1 {
		return "", nil, errors.New("账号已被封禁")
	}
	token, err := jwtpkg.GenerateTokenWithOptions(jwtpkg.TokenOptions{
		UserID:          uint(u.ID),
		Role:            roleFlag(u),
		RoleName:        roleName(u),
		AuthorID:        u.ID,
		AuthorizerAppID: u.BoundAppID,
		TTL:             7 * 24 * time.Hour,
	})
	return token, u, err
}

// BindPromoter 绑定推广员（入驻时通过 promoter_id 绑定）
func (c *Container) BindPromoter(userID, promoterID uint64) error {
	if promoterID == 0 {
		return nil
	}
	if userID == promoterID {
		return errors.New("不能绑定自己为推广员")
	}
	// 铁律 3：防循环
	circular, err := c.UserRepo.CheckCircular(userID, promoterID)
	if err != nil {
		return err
	}
	if circular {
		return errors.New("检测到循环绑定关系，操作被拒绝")
	}
	// 校验推广员存在且 is_promoter=true
	promoter, err := c.UserRepo.GetByID(promoterID)
	if err != nil {
		return errors.New("推广员不存在")
	}
	if !promoter.IsPromoter {
		return errors.New("该用户不是推广员")
	}
	return c.UserRepo.UpdateFields(userID, map[string]any{
		"parent_promoter_user_id": promoterID,
	})
}

func roleFlag(u *model.User) int8 {
	if u.IsAdmin {
		return 9
	}
	if u.IsPromoter {
		return 2
	}
	return 1
}

func roleName(u *model.User) string {
	if u.IsAdmin {
		return "admin"
	}
	if u.IsPromoter {
		return "promoter"
	}
	return "author"
}
