package entities

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(200);not null" json:"name"`
	Category  string         `gorm:"type:varchar(100);not null" json:"category"`
	Price     float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock     int            `gorm:"not null;default:0" json:"stock"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
