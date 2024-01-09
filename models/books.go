package models

import "gorm.io/gorm"

type Books struct {
	ID        uint   `gorm:"primary_key;autoIncrement" json:"id"`
	Author    *string `json:"author"`
	Title     *string `json:"title"`
	Publisher *string `json:"publisher"`
	UserID    uint   `json:"user_id" gorm:"index"` // Foreign key referencing Users table
	User      Users  `json:"user" gorm:"foreignKey:UserID"`
 }

type Users struct {
	ID        uint   `gorm:"primary key;autoIncrement" json:"id"`
	Name      *string `json:"name"`
	Email     *string `json:"email"`
	CreatedAt *gorm.DeletedAt `json:"created_at"`
	UpdatedAt *gorm.DeletedAt `json:"updated_at"`
	Age       int    `json:"age"`
	Books     []Books `json:"books" gorm:"foreignKey:UserID"`
}

// func MigrateBooks(db *gorm.DB)error{
// 	err := db.AutoMigrate(&Books{} , &Users{})
// 	return err
// } 