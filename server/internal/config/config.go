package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Init 加载环境变量
func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found or error loading it. Using system environment variables.")
	}
}

// GetPort 获取服务运行端口
func GetPort() string {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		return "8080"
	}
	return port
}

// GetServerHost 获取监听地址。默认 127.0.0.1（不触发 Windows 防火墙弹窗），
// 生产部署或 LAN 联调时通过 SERVER_HOST=0.0.0.0 显式覆盖。
func GetServerHost() string {
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		return "127.0.0.1"
	}
	return host
}
