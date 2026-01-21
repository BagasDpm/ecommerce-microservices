package usecases

import (
	"ecommerce-microservices/inventory-service/domain/entities"
	"ecommerce-microservices/inventory-service/infrastructure/db"
)

type OrderItemUsecase interface {
	GetByOrderID(orderID uint) ([]*entities.OrderItem, error)
}

type OrderItemUseCase struct {
	orderItemRepo *db.OrderItemRepository
}

func NewOrderItemUsecase(orderItemRepo *db.OrderItemRepository) *OrderItemUseCase {
	return &OrderItemUseCase{
		orderItemRepo: orderItemRepo,
	}
}

func (uc *OrderItemUseCase) GetByOrderID(orderID uint) ([]*entities.OrderItem, error) {
	return uc.orderItemRepo.GetByOrderId(orderID)
}