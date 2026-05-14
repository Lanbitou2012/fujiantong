package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fujiantong/internal/api"
	"fujiantong/internal/middleware"
	"fujiantong/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置基础路由（V2.8 全端点已接入）
func SetupRouter(svc *service.Container) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), middleware.RecoverWithLog(), middleware.CORS())

	// 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "pong - 附件通 is alive!"})
	})

	// 本地存储模式
	r.Static("/uploads", "./uploads")

	// handler 实例化
	authH := api.NewAuthHandler(svc)
	authorH := api.NewAuthorHandler(svc)
	cH := api.NewCHandler(svc)
	promoterH := api.NewPromoterHandler(svc)
	adminH := api.NewAdminHandler(svc)
	wxH := api.NewWxHandler(svc)
	settingsH := api.NewSettingsHandler(svc)

	v1 := r.Group("/api/v1")
	{
		// ── 认证（公开） ──
		auth := v1.Group("/auth")
		{
			auth.POST("/admin/login", authH.AdminLogin)
			auth.POST("/wechat/login", authH.WechatLogin)
			auth.GET("/scan/url", authH.ScanLoginURL)
			auth.GET("/wx-component/url", authH.WxComponentAuthURL)
		}

		// ── 微信开放平台回调（公开） ──
		wx := v1.Group("/wx")
		{
			comp := wx.Group("/component")
			{
				comp.POST("/ticket", wxH.ComponentTicket)
				comp.POST("/notify/:appid", wxH.ComponentNotify)
				comp.GET("/callback", wxH.AuthCallback)
			}
			wx.POST("/media-check/:appid", wxH.MediaCheckCallback)
		}

		// ── C 端（公开） ──
		cGroup := v1.Group("/c")
		{
			cGroup.GET("/file/:code", cH.GetFileByCode)
			cGroup.GET("/download/:code", cH.DownloadFile)
			cGroup.GET("/launch/:code", cH.GetLaunchInfo)
			cGroup.POST("/uv", cH.ReportUV)
		}

		// ── 作者端（需 JWT） ──
		authorGroup := v1.Group("/author")
		authorGroup.Use(middleware.RequireJWT())
		{
			authorGroup.GET("/profile", authorH.GetProfile)
			authorGroup.GET("/files", authorH.ListFiles)
			authorGroup.POST("/files/upload", authorH.UploadFile)
			authorGroup.PUT("/files/:id", authorH.ReplaceFile)
			authorGroup.POST("/files/:id/share-card", authorH.GenerateShareCard)
			authorGroup.GET("/revenue", authorH.GetRevenue)
			authorGroup.POST("/appid", authorH.SubmitAppID)
		}

		// ── 推广员端（需 JWT + 实时校验 is_promoter，铁律 4） ──
		promoterGroup := v1.Group("/promoter")
		promoterGroup.Use(middleware.RequireJWT())
		{
			promoterGroup.GET("/authors", promoterH.ListAuthors)
			promoterGroup.GET("/commission", promoterH.GetCommission)
			promoterGroup.PUT("/payment-info", promoterH.UpdatePaymentInfo)
		}

		// ── 管理员端（需 JWT + 实时校验 is_admin） ──
		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.RequireJWT(), middleware.RequireRole(9))
		{
			adminGroup.GET("/dashboard", adminH.Dashboard)
			adminGroup.GET("/todo", adminH.TodoList)
			adminGroup.GET("/users", adminH.ListUsers)
			adminGroup.PUT("/users/:id/promoter", adminH.TogglePromoter)
			adminGroup.PUT("/users/:id/status", adminH.BanUser)
			adminGroup.GET("/settlements", adminH.ListSettlements)
			adminGroup.POST("/settlements/:id/approve", adminH.ApproveSettlement)
			adminGroup.POST("/settlements/:id/mark-paid", adminH.MarkSettlementPaid)

			// 设置：修改密码 + 平台凭据（微信开放平台 / COS）
			adminGroup.POST("/settings/password", settingsH.ChangePassword)
			adminGroup.GET("/settings/platform", settingsH.GetPlatformSettings)
			adminGroup.PUT("/settings/platform", settingsH.UpdatePlatformSettings)
		}
	}

	// 前端 SPA 静态文件托管（生产部署模式：admin/dist 打包到 ./public）
	// 与公众号附件助手项目同模式：Nginx 仅反代 8081，前端由 Go 后端 serve
	const publicDir = "./public"
	if _, err := os.Stat(publicDir); err == nil {
		// 静态资源（带 hash 的 assets/ 等）
		r.Static("/assets", filepath.Join(publicDir, "assets"))
		r.StaticFile("/favicon.ico", filepath.Join(publicDir, "favicon.ico"))

		// SPA history 模式：所有非 /api、非 /uploads、非 /ping、非 /assets 都回退到 index.html
		r.NoRoute(func(c *gin.Context) {
			p := c.Request.URL.Path
			if strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/uploads/") {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "not found"})
				return
			}
			c.File(filepath.Join(publicDir, "index.html"))
		})
	}

	return r
}
