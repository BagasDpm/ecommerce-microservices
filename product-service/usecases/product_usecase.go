package usecases

import (
	"ecommerce-microservices/product-service/domain/entities"
	"ecommerce-microservices/product-service/domain/repositories"
	"ecommerce-microservices/product-service/infrastructure/cache"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type ProductUsecase interface {
	GetProducts(category string) ([]entities.Product, error)
	GetProductByID(id uint) (*entities.Product, error)
}

type productUsecase struct {
	productRepo repositories.ProductRepository
	cache       *cache.RedisCache
}

func NewProductUsecase(productRepo repositories.ProductRepository, cache *cache.RedisCache) ProductUsecase {
	return &productUsecase{
		productRepo: productRepo,
		cache:       cache,
	}
}

func (u *productUsecase) GetProducts(category string) ([]entities.Product, error) {
	cacheKey := fmt.Sprintf("products:%s", category)

	var products []entities.Product
	err := u.cache.Get(cacheKey, &products)
	if err == nil {
		return products, nil
	}

	if err != redis.Nil {
		return nil, err
	}

	products, err = u.productRepo.FindAll(category)
	if err != nil {
		return nil, err
	}

	u.cache.Set(cacheKey, products, 5*time.Minute)

	return products, nil
}

func (u *productUsecase) GetProductByID(id uint) (*entities.Product, error) {
	cacheKey := fmt.Sprintf("product:%d", id)

	var product entities.Product
	err := u.cache.Get(cacheKey, &product)
	if err == nil {
		return &product, nil
	}

	if err != redis.Nil {
		return nil, err
	}

	productData, err := u.productRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("product not found")
	}

	u.cache.Set(cacheKey, productData, 5*time.Minute)

	return productData, nil
}
