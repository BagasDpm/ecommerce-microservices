package main

import (
	"log"

	"ecommerce-microservices/config"
	"ecommerce-microservices/inventory-service/handlers"
	"ecommerce-microservices/inventory-service/infrastructure/cache"
	"ecommerce-microservices/inventory-service/infrastructure/db"
	"ecommerce-microservices/inventory-service/infrastructure/messaging"
	"ecommerce-microservices/inventory-service/services"
	"ecommerce-microservices/inventory-service/usecases"
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

	// Initialize Redis cache
	redisCache := cache.NewRedisCache("localhost:6379", "", 0)

	// Initialize repositories
	orderRepo := db.NewOrderRepository(database)
	orderItemRepo := db.NewOrderItemRepository(database)
	productRepo := db.NewProductRepository(database)

	// Initialize usecases
	orderUsecase := usecases.NewOrderUsecase(orderRepo)
	orderItemUsecase := usecases.NewOrderItemUsecase(orderItemRepo)
	productUsecase := usecases.NewProductUsecase(productRepo, redisCache)

	// Initialize email service
	emailService := services.NewMockEmailService()

	// Initialize RabbitMQ publisher
	publisher, err := messaging.NewRabbitMQPublisher("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ publisher: %v", err)
	}
	defer publisher.Close()

	// Initialize handlers
	inventoryHandler := handlers.NewInventoryHandler(orderUsecase, orderItemUsecase, productUsecase, publisher)
	notificationHandler := handlers.NewNotificationHandler(orderUsecase, emailService)

	// Start services
	log.Println("Starting inventory and notification services...")

	// Start inventory worker
	go inventoryHandler.OrderPlacedWorker()

	// Start notification workers
	notificationHandler.StartNotificationWorkers()
}