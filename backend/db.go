package main

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// clauseForUpdate 返回行级排他锁子句（需配合事务使用）。
func clauseForUpdate() clause.Locking {
	return clause.Locking{Strength: "UPDATE"}
}

func initDB() {
	var err error
	DB, err = gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("连接 MySQL 失败: %v", err)
	}
	if err := DB.AutoMigrate(&User{}, &Wish{}, &TasteResponse{}, &CoinTransaction{}); err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}
	log.Println("MySQL 连接成功，数据表已就绪")
}
