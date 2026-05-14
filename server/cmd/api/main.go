package main

import (
	"log"
	"net/http"

	"fujiantong/internal/config"
	"fujiantong/internal/router"
	"fujiantong/internal/scheduler"
	"fujiantong/internal/service"
	"fujiantong/internal/wechat"
)

func main() {
	// 1. 加载配置
	config.Init()

	// 2. 初始化数据库
	config.InitDB()

	// 3. 初始化 Redis 缓存
	config.InitRedis()

	// 4. 初始化 Service Container
	if config.DB == nil {
		log.Fatal("数据库连接失败，无法启动。请检查 .env 中 DB_DSN 配置。")
	}
	wxComp := &wechat.ComponentClient{
		HTTPClient: &http.Client{},
	}
	svc := service.NewContainer(config.DB, wxComp)

	// 5. 启动 Cron 调度器（W3 P0：自动化数据 Pipeline）
	sched := scheduler.New(svc)
	sched.Start()
	defer sched.Stop()

	// 6. 设置路由
	r := router.SetupRouter(svc)

	// 7. 启动服务
	port := config.GetPort()
	// 默认监听 127.0.0.1 避免 Windows 防火墙弹窗；
	// 生产环境通过 SERVER_HOST=0.0.0.0 显式启用对外监听。
	host := config.GetServerHost()
	addr := host + ":" + port
	log.Printf("附件通 Server is running on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
