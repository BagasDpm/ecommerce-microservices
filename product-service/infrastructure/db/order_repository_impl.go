package db

import (
	"ecommerce-microservices/product-service/domain/entities"
	"ecommerce-microservices/product-service/domain/repositories"

	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) repositories.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *entities.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) CreateOrderItems(orderItems []entities.OrderItem) error {
	if len(orderItems) == 0 {
		return nil
	}
	return r.db.Create(&orderItems).Error
}

func (r *orderRepository) FindByID(id uint) (*entities.Order, error) {
	var order entities.Order
	err := r.db.Preload("Items.Product").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) UpdateStatus(id uint, status entities.OrderStatus) error {
	return r.db.Model(&entities.Order{}).
		Where("id = ?", id).
		Update("status", status).
		Error
}
