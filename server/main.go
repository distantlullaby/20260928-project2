package main

import (
	"log"

	"taste-server/config"
	"taste-server/database"
	"taste-server/handlers"
	"taste-server/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	database.Init()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5180", "http://127.0.0.1:5180", "http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	r.Static("/uploads", config.UploadDir)

	api := r.Group("/api")
	{
		api.POST("/register", handlers.Register)
		api.POST("/login", handlers.Login)

		// Feed 流：登录可选（登录后会标记哪些是我发布的）
		api.GET("/wishes", middleware.OptionalAuth(), handlers.ListWishes)
		api.GET("/wishes/:id", middleware.OptionalAuth(), handlers.GetWish)

		auth := api.Group("")
		auth.Use(middleware.Auth())
		{
			auth.GET("/me", handlers.Me)
			auth.POST("/upload", handlers.Upload)

			auth.POST("/wishes", handlers.CreateWish)
			auth.POST("/wishes/:id/bounty", handlers.AddBounty)
			auth.POST("/wishes/:id/cancel", handlers.CancelWish)
			auth.POST("/wishes/:id/responses", handlers.CreateResponse)
			auth.POST("/wishes/:id/accept/:responseId", handlers.AcceptWish)

			auth.GET("/my/coins", handlers.MyCoinLogs)
			auth.GET("/my/wishes", handlers.MyWishes)
			auth.GET("/my/responses", handlers.MyResponses)
		}
	}

	log.Printf("替你尝一口服务已启动: %s", config.ListenAddr)
	if err := r.Run(config.ListenAddr); err != nil {
		log.Fatal(err)
	}
}
