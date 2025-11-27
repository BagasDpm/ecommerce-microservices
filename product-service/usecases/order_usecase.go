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

	order := &entities.Order{
		UserID:     userID,
		Status:     entities.OrderStatusPending,
		TotalPrice: totalPrice,
		Items:      orderItems,
	}

	if err := u.orderRepo.Create(order); err != nil {
		return nil, err
	}

	if err := u.publisher.PublishOrderPlaced(order.ID); err != nil {
		log.Printf("Failed to publish order placed message: %v", err)
	}

	return order, nil
}
