package main

import (
	"ecommerce-microservices/auth-service/handlers"
	"ecommerce-microservices/auth-service/infrastructure/db"
	"ecommerce-microservices/auth-service/shared/middleware"
	"ecommerce-microservices/auth-service/usecases"
	"ecommerce-microservices/config"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	userRepo := db.NewUserRepository(database)
	authUsecase := usecases.NewAuthUsecase(userRepo, cfg.JWT.Secret)
	authHandler := handlers.NewAuthHandler(authUsecase)

	router := gin.Default()

	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		protected.GET("/profile", authHandler.GetProfile)
		protected.PUT("/profile", authHandler.UpdateProfile)
	}

	log.Printf("Auth Service running on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
