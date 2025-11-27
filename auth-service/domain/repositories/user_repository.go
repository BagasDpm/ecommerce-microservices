package repositories

import "ecommerce-microservices/auth-service/domain/entities"

type UserRepository interface {
	Create(user *entities.User) error
	FindByEmail(email string) (*entities.User, error)
	FindByID(id uint) (*entities.User, error)
	Update(user *entities.User) error
}
