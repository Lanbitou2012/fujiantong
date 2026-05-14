package config

import (
	"log"
	"os"

	"fujiantong/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化 MySQL 数据库连接
func InitDB() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Println("Warning: DB_DSN is not set in .env")
		return
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: Failed to connect to MySQL database: %v. Please check your local MySQL service and .env config.\n", err)
		return
	}

	// 自动迁移 (AutoMigrate)
	// GORM 会自动读取 model 目录下的结构体，并在 MySQL 中自动创建或更新表结构
	// 这样我们就彻底告别了手动写 SQL 建表的痛苦
	err = db.AutoMigrate(
		&model.User{},
		&model.UserRoleHistory{},
		&model.File{},
		&model.FileShareCard{},
		&model.Authorization{},
		&model.DailyAdStat{},
		&model.SettlementRecord{},
		&model.CommissionSettlement{},
		&model.WxToken{},
		&model.MpTemplateVersion{},
		&model.MpDeployment{},
		&model.MpAudit{},
		&model.AdminActionLog{},
		&model.SystemSetting{},
	)
	if err != nil {
		log.Printf("Warning: Failed to AutoMigrate tables: %v\n", err)
	} else {
		log.Println("Database AutoMigrate successfully!")
	}

	DB = db
	log.Println("MySQL Database connection established successfully")
	SeedInitialData()
}
