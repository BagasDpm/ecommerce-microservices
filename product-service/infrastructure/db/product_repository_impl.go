package db

import (
	"ecommerce-microservices/product-service/domain/entities"
	"ecommerce-microservices/product-service/domain/repositories"

	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) repositories.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) FindAll(category string) ([]entities.Product, error) {
	var products []entities.Product
	query := r.db

	if category != "" {
		query = query.Where("category = ?", category)
	}

	err := query.Find(&products).Error
	return products, err
}

func (r *productRepository) FindByID(id uint) (*entities.Product, error) {
	var product entities.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) UpdateStock(id uint, quantity int) error {
	return r.db.Model(&entities.Product{}).
		Where("id = ?", id).
		Update("stock", gorm.Expr("stock - ?", quantity)).
		Error
}
