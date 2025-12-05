package db

import (
	"ecommerce-microservices/inventory-service/domain/entities"

	"gorm.io/gorm"
)

type OrderItemRepository struct {
	db *gorm.DB
}

func NewOrderItemRepository(db *gorm.DB) *OrderItemRepository {
	return &OrderItemRepository{db: db}
}


func (r *OrderItemRepository) GetByOrderId(ordeId uint) ([]*entities.OrderItem, error) {
	// find all order items by order id
	var orderItems []*entities.OrderItem
	err := r.db.Preload("Product").Where("order_id = ?", ordeId).Find(&orderItems).Error
	if err != nil {
		return nil, err
	}

	return orderItems, nil
}
