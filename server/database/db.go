package database

import (
	"log"

	"taste-server/config"
	"taste-server/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Init 连接数据库并自动建表
func Init() {
	db, err := gorm.Open(mysql.Open(config.MySQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接 MySQL 失败: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Wish{},
		&models.TastingResponse{},
		&models.CoinTransaction{},
	); err != nil {
		log.Fatalf("自动建表失败: %v", err)
	}
	DB = db
	log.Println("数据库初始化完成")
}
