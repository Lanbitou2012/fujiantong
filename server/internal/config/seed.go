package config

import (
	"log"
	"os"

	"fujiantong/internal/model"

	"golang.org/x/crypto/bcrypt"
)

func SeedInitialData() {
	if DB == nil {
		return
	}
	seedInitialAdmin()
}

// seedInitialAdmin 首次启动时，若 users 表中无任何 is_admin=true 的记录，
// 则按 ADMIN_USERNAME / ADMIN_PASSWORD 创建一个管理员。
func seedInitialAdmin() {
	var count int64
	if err := DB.Model(&model.User{}).Where("is_admin = ?", true).Count(&count).Error; err != nil {
		log.Printf("Warning: failed to count admins: %v", err)
		return
	}
	if count > 0 {
		return
	}

	username := os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "admin123456"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Warning: failed to hash initial admin password: %v", err)
		return
	}

	admin := model.User{
		AdminUsername:       username,
		Nickname:            "平台管理员",
		LoginPassphraseHash: string(hash),
		IsAdmin:             true,
		IsPromoter:          true, // 主理人通常同时也是推广员
		Status:              1,
	}
	if err := DB.Create(&admin).Error; err != nil {
		log.Printf("Warning: failed to create initial admin: %v", err)
		return
	}
	log.Printf("Initial admin created: %s", username)
}
