package main

import "os"

// 配置项均可通过环境变量覆盖，默认值适配本地开发环境。
var (
	// MySQL DSN，例如 root:123456@tcp(127.0.0.1:3306)/taste_foryou?charset=utf8mb4&parseTime=True&loc=Local
	mysqlDSN = getenv("MYSQL_DSN", "root:123456@tcp(127.0.0.1:3306)/taste_foryou?charset=utf8mb4&parseTime=True&loc=Local")
	// 后端监听端口
	serverPort = getenv("PORT", "8080")
	// JWT 签名密钥
	jwtSecret = getenv("JWT_SECRET", "taste-foryou-dev-secret")
	// 上传文件保存目录
	uploadDir = getenv("UPLOAD_DIR", "uploads")
)

// 注册即赠送的初始尝鲜币
const initialCoins = 100

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
