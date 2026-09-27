package config

import "os"

// 配置项均支持环境变量覆盖，默认连接本地 MySQL
var (
	MySQLDSN   = getEnv("MYSQL_DSN", "root:123456@tcp(127.0.0.1:3306)/taste_for_you?charset=utf8mb4&parseTime=True&loc=Local")
	JWTSecret  = getEnv("JWT_SECRET", "taste-for-you-secret-key")
	ListenAddr = getEnv("LISTEN_ADDR", ":8081")
	UploadDir  = getEnv("UPLOAD_DIR", "uploads")
)

// 注册即赠送的初始尝鲜币
const RegisterBonus = 100

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
