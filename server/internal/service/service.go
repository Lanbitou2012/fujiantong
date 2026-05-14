package service

import (
	"fujiantong/internal/repo"
	"fujiantong/internal/wechat"

	"gorm.io/gorm"
)

// Container 统一持有所有 repo 和微信客户端，供各 service 使用。
type Container struct {
	DB             *gorm.DB
	UserRepo       *repo.UserRepo
	FileRepo       *repo.FileRepo
	AuthRepo       *repo.AuthorizationRepo
	FinanceRepo    *repo.FinanceRepo
	WechatRepo     *repo.WechatRepo
	DeploymentRepo *repo.DeploymentRepo
	AdminRepo      *repo.AdminRepo
	SettingRepo    *repo.SettingRepo
	WxComponent    *wechat.ComponentClient
}

func NewContainer(db *gorm.DB, wxComp *wechat.ComponentClient) *Container {
	return &Container{
		DB:             db,
		UserRepo:       repo.NewUserRepo(db),
		FileRepo:       repo.NewFileRepo(db),
		AuthRepo:       repo.NewAuthorizationRepo(db),
		FinanceRepo:    repo.NewFinanceRepo(db),
		WechatRepo:     repo.NewWechatRepo(db),
		DeploymentRepo: repo.NewDeploymentRepo(db),
		AdminRepo:      repo.NewAdminRepo(db),
		SettingRepo:    repo.NewSettingRepo(db),
		WxComponent:    wxComp,
	}
}
