package models

type Books struct {
	ID        uint   `gorm:"primary_key;autoIncrement" json:"id"`
	Author    *string `json:"author"`
	Title     *string `json:"title"`
	Publisher *string `json:"publisher"`
	UserID    uint   `json:"user_id" gorm:"index"` // Foreign key referencing Users table
	User      Users  `json:"user" gorm:"foreignKey:UserID"`
 }


// func MigrateBooks(db *gorm.DB)error{
// 	err := db.AutoMigrate(&Books{} , &Users{})
// 	return err
// } 