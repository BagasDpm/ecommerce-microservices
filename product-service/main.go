package main

import (
	"ecommerce-microservices/config"
	"ecommerce-microservices/product-service/handlers"
	"ecommerce-microservices/product-service/infrastructure/cache"
	"ecommerce-microservices/product-service/infrastructure/db"
	"ecommerce-microservices/product-service/infrastructure/messaging"
	"ecommerce-microservices/product-service/shared/middleware"
	"ecommerce-microservices/product-service/usecases"
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

	redisCache, err := cache.NewRedisCache(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	publisher, err := messaging.NewRabbitMQPublisher(&cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer publisher.Close()

	productRepo := db.NewProductRepository(database)
	orderRepo := db.NewOrderRepository(database)

	productUsecase := usecases.NewProductUsecase(productRepo, redisCache)
	orderUsecase := usecases.NewOrderUsecase(orderRepo, productRepo, publisher)

	productHandler := handlers.NewProductHandler(productUsecase)
	orderHandler := handlers.NewOrderHandler(orderUsecase)

	router := gin.Default()

	router.GET("/products", productHandler.GetProducts)
	router.GET("/products/:id", productHandler.GetProductByID)

	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		protected.POST("/orders", orderHandler.CreateOrder)
	}

	log.Printf("Product Service running on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
