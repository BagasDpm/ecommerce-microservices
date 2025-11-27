package repositories

import "ecommerce-microservices/product-service/domain/entities"

type ProductRepository interface {
	FindAll(category string) ([]entities.Product, error)
	FindByID(id uint) (*entities.Product, error)
	UpdateStock(id uint, quantity int) error
}
