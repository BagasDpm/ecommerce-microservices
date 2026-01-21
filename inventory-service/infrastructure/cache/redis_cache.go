package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ecommerce-microservices/inventory-service/domain/entities"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string, db int) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{client: rdb}
}

func (r *RedisCache) GetProduct(ctx context.Context, productID uint) (*entities.Product, error) {
	key := fmt.Sprintf("product:%d", productID)

	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}

	var product entities.Product
	err = json.Unmarshal([]byte(val), &product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *RedisCache) SetProduct(ctx context.Context, product *entities.Product) error {
	key := fmt.Sprintf("product:%d", product.ID)

	data, err := json.Marshal(product)
	if err != nil {
		return err
	}

	// Cache for 5 minutes as specified
	return r.client.Set(ctx, key, data, 5*time.Minute).Err()
}

func (r *RedisCache) DeleteProduct(ctx context.Context, productID uint) error {
	key := fmt.Sprintf("product:%d", productID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) GetProducts(ctx context.Context, cacheKey string) ([]*entities.Product, error) {
	val, err := r.client.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}

	var products []*entities.Product
	err = json.Unmarshal([]byte(val), &products)
	return products, err
}

func (r *RedisCache) SetProducts(ctx context.Context, cacheKey string, products []*entities.Product) error {
	data, err := json.Marshal(products)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, cacheKey, data, 5*time.Minute).Err()
}