package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/yerlan/go-fiber-postgres/models"
	"github.com/yerlan/go-fiber-postgres/storage"
	"gorm.io/gorm"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
	UserID    uint   `json:"user_id"`
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) CreateBook(context *fiber.Ctx) error {
	book := Book{}
	err := context.BodyParser(&book)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	err = r.DB.Create(&book).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "couldn't create a book"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book has been created"})

	return nil
}

func (r *Repository) CreateUser(context *fiber.Ctx) error {
	user := User{}
	err := context.BodyParser(&user)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	err = r.DB.Create(&user).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "couldn't create a user"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "user has been created"})
	return nil
}

func (r *Repository) DeleteBook(context *fiber.Ctx) error {
	bookModel := models.Books{}
	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty"})
		return nil
	}
	err := r.DB.Delete(bookModel, id)

	if err.Error != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not delete book",
		})
		return err.Error
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book succesfully deleted",
	})
	return nil
}

func (r *Repository) DeleteUser(context *fiber.Ctx) error {
	userModel := models.Users{}
	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty"})
		return nil
	}

	err := r.DB.Delete(userModel, id).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not delete user",
		})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "user successfully deleted",
	})
	return nil
}

func (r *Repository) GetBooks(context *fiber.Ctx) error {
	bookModels := &[]models.Books{}

	err := r.DB.Preload("User").Find(bookModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get books"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "books fetched good ",
		"data":    bookModels,
	})
	return nil
}

func (r *Repository) GetUsers(context *fiber.Ctx) error {
	userModels := &[]models.Users{}

	err := r.DB.Preload("Books").Find(userModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get users"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "users fetched successfully",
		"data":    userModels,
	})
	return nil
}

func (r *Repository) GetBookByID(context *fiber.Ctx) error {
	id := context.Params("id")
	bookModel := &models.Books{}

	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id can not be empty",
		})
		return nil
	}
	fmt.Println("the ID is ", id)

	err := r.DB.Preload("User").Where("id = ?", id).First(bookModel).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "can not get book by id",
		})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book id fetched goodly",
		"data":    bookModel,
	})
	return nil
}

func (r *Repository) UpdateBook(context *fiber.Ctx) error {
	id := context.Params("id")
	bookModel := &models.Books{}

	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}

	// Retrieve the book by ID
	err := r.DB.Where("id = ?", id).First(bookModel).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "can not get book by id",
		})
		return err
	}

	// Parse the request body to update the book
	newBook := &models.Books{} // Use the correct struct name here
	err = context.BodyParser(&newBook)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	// Update the book fields
	bookModel.Author = newBook.Author
	bookModel.Title = newBook.Title
	bookModel.Publisher = newBook.Publisher
	bookModel.UserID = newBook.UserID

	// Save the updated book
	err = r.DB.Save(bookModel).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not update book"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book successfully updated",
		"data":    bookModel,
	})
	return nil
}

func (r *Repository) GetUserByID(context *fiber.Ctx) error {
	id := context.Params("id")
	userModel := &models.Users{}

	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}

	err := r.DB.Where("id = ?", id).Preload("Books").First(userModel).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not get user by id",
		})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "user id fetched successfully",
		"data":    userModel,
	})
	return nil
}

func (r *Repository) GetBooksByUserID(context *fiber.Ctx) error {
    id := context.Params("id")
    userBooks := &[]models.Books{}

    if id == "" {
        context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
            "message": "id cannot be empty",
        })
        return nil
    }

    // Retrieve books with the specified user ID
    err := r.DB.Where("user_id = ?", id).Find(userBooks).Error
    if err != nil {
        context.Status(http.StatusBadRequest).JSON(&fiber.Map{
            "message": "could not get books for user",
        })
        return err
    }

    context.Status(http.StatusOK).JSON(&fiber.Map{
        "message": "books for user fetched successfully",
        "data":    userBooks,
    })
    return nil
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_books", r.CreateBook)
	api.Delete("/delete_book/:id", r.DeleteBook)
	api.Get("/get_books/:id", r.GetBookByID)
	api.Get("/books", r.GetBooks)
	api.Post("/create_users", r.CreateUser)
	api.Delete("/delete_user/:id", r.DeleteUser)
	api.Get("/get_users/:id", r.GetUserByID)
	api.Get("/users", r.GetUsers)
	api.Put("/update_book/:id", r.UpdateBook)
	api.Get("/get_books_by_user/:id", r.GetBooksByUserID)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("could not load the database")
	}

	// err = models.MigrateBooks(db)
	// if err != nil {
	// 	log.Fatal("could not migrate db")
	// }

	r := Repository{
		DB: db,
	}

	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":8080")
}
