package repositories

import "ecommerce-microservices/inventory-service/domain/entities"

type OrderItemRepository interface {
    GetByOrderId(orderID uint) ([]*entities.OrderItem, error)
    Create(orderItem *entities.OrderItem) error
}