package db

import (
	"fmt"

	"ecommerce-microservices/inventory-service/domain/entities"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) FindByID(id uint) (*entities.Product, error) {
	var product entities.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) DecreaseStockByIds(updates []struct {
	ProductID uint
	Quantity  int
}) error {
	if len(updates) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		// Use SELECT FOR UPDATE to lock rows atomically
		for _, update := range updates {
			var product entities.Product

			// Lock the row for update to prevent race conditions
			err := tx.Set("gorm:query_option", "FOR UPDATE").
				Select("id, stock").
				Where("id = ?", update.ProductID).
				First(&product).Error
			if err != nil {
				return fmt.Errorf("product not found: %d", update.ProductID)
			}

			// Check stock availability
			if product.Stock < update.Quantity {
				return fmt.Errorf("insufficient stock for product %d: available=%d, required=%d",
					update.ProductID, product.Stock, update.Quantity)
			}
		}

		// Atomically update all stocks
		for _, update := range updates {
			result := tx.Model(&entities.Product{}).
				Where("id = ?", update.ProductID).
				Update("stock", gorm.Expr("stock - ?", update.Quantity))

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected == 0 {
				return fmt.Errorf("failed to update stock for product %d", update.ProductID)
			}
		}

		return nil
	})
}
