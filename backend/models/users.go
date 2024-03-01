package models

import (
	"time"
)

type Users struct {
	ID        uint   `gorm:"primary key;autoIncrement" json:"id"`
	Name      *string `json:"name"`
	Email     *string `json:"email"`
	Password  *string `json:"password"`
	Role      *string   `json:"role" gorm:"default:user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Age       int       `json:"age" gorm:"default:18"`
	Books     []Books `json:"books" gorm:"foreignKey:UserID"`
}