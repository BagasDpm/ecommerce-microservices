package repositories

import "ecommerce-microservices/inventory-service/domain/entities"

type ProductRepository interface {
    FindByID(id uint) (*entities.Product, error)
    DecreaseStockByIds(updates []struct {
        ProductID uint
        Quantity  int
    }) error
}
