package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// corsMiddleware 开发期允许 Vite 开发服务器跨域访问。
func corsMiddleware() gin.HandlerFunc {
	allowed := map[string]bool{
		"http://localhost:5173": true,
		"http://127.0.0.1:5173": true,
		"http://localhost:5174": true,
		"http://127.0.0.1:5174": true,
		"http://localhost:5175": true,
		"http://127.0.0.1:5175": true,
	}
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func main() {
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		log.Fatalf("上传目录初始化失败: %v", err)
	}

	initDB()
	seed()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 开发期允许 Vite 开发服务器跨域访问
	r.Use(corsMiddleware())

	// 上传图片静态访问
	r.Static("/uploads", uploadDir)

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "替你尝一口 API 运行中"})
	})

	api := r.Group("/api")
	{
		api.POST("/register", handleRegister)
		api.POST("/login", handleLogin)

		// 心愿 Feed（公开浏览）
		api.GET("/wishes", handleListWishes)
		api.GET("/wishes/:id", handleGetWish)
		api.GET("/images", func(c *gin.Context) {
			// 兼容预留：图片通过 /uploads 静态服务
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		auth := api.Group("")
		auth.Use(authRequired())
		{
			auth.GET("/me", handleMe)

			auth.POST("/wishes", handleCreateWish)
			auth.POST("/wishes/:id/append", handleAppendReward)
			auth.POST("/wishes/:id/cancel", handleCancelWish)
			auth.POST("/wishes/:id/settle", handleSettle)
			auth.POST("/wishes/:id/responses", handleCreateResponse)

			auth.GET("/my/wishes", handleMyWishes)
			auth.GET("/my/responses", handleMyResponses)
			auth.GET("/my/transactions", handleMyTransactions)

			auth.POST("/upload", handleUpload)
		}
	}

	log.Printf("后端服务启动：http://localhost:%s", serverPort)
	if err := r.Run(":" + serverPort); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务启动失败: %v", err)
	}
}
