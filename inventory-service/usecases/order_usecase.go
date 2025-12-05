package usecases

import (
	"ecommerce-microservices/inventory-service/domain/entities"
	"ecommerce-microservices/inventory-service/infrastructure/db"
)

type OrderUsecase interface {
	GetOrder(id uint) (*entities.Order, error)
	UpdateStatus(id uint, status entities.OrderStatus) error
}

type OrderUseCase struct {
	orderRepo *db.OrderRepository
}

func NewOrderUsecase(orderRepo *db.OrderRepository) *OrderUseCase {
	return &OrderUseCase{
		orderRepo: orderRepo,
	}
}

func (uc *OrderUseCase) GetOrder(id uint) (*entities.Order, error) {
	return uc.orderRepo.FindByID(id)
}

func (uc *OrderUseCase) UpdateStatus(id uint, status entities.OrderStatus) error {
	return uc.orderRepo.UpdateStatus(id, status)
}