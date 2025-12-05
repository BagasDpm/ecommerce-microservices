package repositories

import "ecommerce-microservices/inventory-service/domain/entities"

type OrderRepository interface {
    Create(order *entities.Order) error
    FindByID(id uint) (*entities.Order, error)
    UpdateStatus(id uint, status entities.OrderStatus) error
}
