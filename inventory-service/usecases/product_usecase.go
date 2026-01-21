package usecases

import (
	"context"
	"log"

	"ecommerce-microservices/inventory-service/domain/entities"
	"ecommerce-microservices/inventory-service/infrastructure/cache"
	"ecommerce-microservices/inventory-service/infrastructure/db"
)

type ProductUsecase interface {
    GetByID(id uint) (*entities.Product, error)
    GetByIDWithCache(ctx context.Context, id uint) (*entities.Product, error)
    DecreaseStock(updates []struct {
        ProductID uint
        Quantity  int
    }) error
}

type ProductUseCase struct {
    productRepo *db.ProductRepository
    cache       *cache.RedisCache
}

func NewProductUsecase(productRepo *db.ProductRepository, cache *cache.RedisCache) *ProductUseCase {
    return &ProductUseCase{
        productRepo: productRepo,
        cache:       cache,
    }
}

func (uc *ProductUseCase) GetByID(id uint) (*entities.Product, error) {
    return uc.productRepo.FindByID(id)
}

func (uc *ProductUseCase) GetByIDWithCache(ctx context.Context, id uint) (*entities.Product, error) {
    // Try cache first
    if uc.cache != nil {
        product, err := uc.cache.GetProduct(ctx, id)
        if err != nil {
            log.Printf("Cache error for product %d: %v", id, err)
        }
        if product != nil {
            log.Printf("Cache hit for product %d", id)
            return product, nil
        }
    }

    // Cache miss, get from database
    log.Printf("Cache miss for product %d, fetching from DB", id)
    product, err := uc.productRepo.FindByID(id)
    if err != nil {
        return nil, err
    }

    // Update cache
    if uc.cache != nil {
        if cacheErr := uc.cache.SetProduct(ctx, product); cacheErr != nil {
            log.Printf("Failed to cache product %d: %v", id, cacheErr)
        }
    }

    return product, nil
}

func (uc *ProductUseCase) DecreaseStock(updates []struct {
    ProductID uint
    Quantity  int
}) error {
    err := uc.productRepo.DecreaseStockByIds(updates)
    if err != nil {
        return err
    }

    // Invalidate cache for updated products
    if uc.cache != nil {
        ctx := context.Background()
        for _, update := range updates {
            if cacheErr := uc.cache.DeleteProduct(ctx, update.ProductID); cacheErr != nil {
                log.Printf("Failed to invalidate cache for product %d: %v", update.ProductID, cacheErr)
            }
        }
    }

    return nil
}
