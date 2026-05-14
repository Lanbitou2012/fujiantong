package config

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client
var Ctx = context.Background()

// InitRedis 初始化 Redis 连接（应对十万级瞬间高并发下载的防击穿设计）
func InitRedis() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // 默认地址
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	RDB = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword, // no password set
		DB:       0,             // use default DB
		PoolSize: 100,           // 连接池大小，应对高并发
	})

	_, err := RDB.Ping(Ctx).Result()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis database: %v. Cache breakdown protection will be disabled.\n", err)
		RDB = nil // 置空，防止业务代码里报错崩溃，让系统降级
		return
	}

	log.Println("Redis connection established successfully. Hot-file cache is ready!")
}
