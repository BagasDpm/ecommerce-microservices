package usecases

import (
	"ecommerce-microservices/product-service/domain/entities"
	"ecommerce-microservices/product-service/domain/repositories"
	"ecommerce-microservices/product-service/infrastructure/messaging"
	"errors"
	"log"
)

type OrderUsecase interface {
	CreateOrder(userID uint, req *entities.CreateOrderRequest) (*entities.Order, error)
}

type orderUsecase struct {
	orderRepo   repositories.OrderRepository
	productRepo repositories.ProductRepository
	publisher   *messaging.RabbitMQPublisher
}

func NewOrderUsecase(
	orderRepo repositories.OrderRepository,
	productRepo repositories.ProductRepository,
	publisher *messaging.RabbitMQPublisher,
) OrderUsecase {
	return &orderUsecase{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		publisher:   publisher,
	}
}

func (u *orderUsecase) CreateOrder(userID uint, req *entities.CreateOrderRequest) (*entities.Order, error) {
	var totalPrice float64
	var orderItems []entities.OrderItem

	// Validate products and calculate total
	for _, item := range req.Items {
		product, err := u.productRepo.FindByID(item.ProductID)
		if err != nil {
			return nil, errors.New("product not found")
		}

		if product.Stock < item.Quantity {
			return nil, errors.New("insufficient stock for product: " + product.Name)
		}

		orderItem := entities.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		}
		orderItems = append(orderItems, orderItem)
		totalPrice += product.Price * float64(item.Quantity)
	}

	// Create order first
	order := &entities.Order{
		UserID:     userID,
		Status:     entities.OrderStatusPending,
		TotalPrice: totalPrice,
	}

	if err := u.orderRepo.Create(order); err != nil {
		return nil, err
	}

	// Now set OrderID and create items explicitly
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
	}

	// Create order items (you'll need to add this method to your repository)
	if err := u.orderRepo.CreateOrderItems(orderItems); err != nil {
		// If items creation fails, you might want to rollback the order
		return nil, err
	}

	// Set items back to order for response
	order.Items = orderItems

	if err := u.publisher.PublishOrderPlaced(order.ID); err != nil {
		log.Printf("Failed to publish order placed message: %v", err)
	}

	return order, nil
}
