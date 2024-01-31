package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
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
	DB     *gorm.DB
	Logger *logrus.Logger
}

func (r *Repository) CreateBook(context *fiber.Ctx) error {
	book := Book{}
	err := context.BodyParser(&book)
	if err != nil {
		r.Logger.WithError(err).Error("failed to parse request body")
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	err = r.DB.Create(&book).Error
	if err != nil {
		r.Logger.WithError(err).Error("failed to create book")
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "couldn't create a book"})
		return err
	}

	r.Logger.Info("Book has been created", logrus.Fields{"author": book.Author, "title": book.Title, "publisher": book.Publisher, "userID": book.UserID})

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book has been created"})

	return nil
}

func (r *Repository) CreateUser(context *fiber.Ctx) error {
	user := User{}
	err := context.BodyParser(&user)
	if err != nil {
		r.Logger.WithError(err).Error("failed to parse request body for CreateUser")
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	err = r.DB.Create(&user).Error
	if err != nil {
		r.Logger.WithError(err).Error("failed to create user")
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "couldn't create a user"})
		return err
	}
	r.Logger.Info("User has been created", logrus.Fields{"name": user.Name, "email": user.Email, "age": user.Age})
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "user has been created"})
	return nil
}

func (r *Repository) DeleteBook(context *fiber.Ctx) error {
	bookModel := models.Books{}
	id := context.Params("id")
	if id == "" {
		r.Logger.Error("id cannot be empty for DeleteBook")
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty"})
		return nil
	}
	err := r.DB.Delete(bookModel, id)

	if err.Error != nil {
		r.Logger.WithError(err.Error).Error("could not delete book for DeleteBook")
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not delete book",
		})
		return err.Error
	}
	r.Logger.Info("Book successfully deleted", logrus.Fields{"bookID": id})
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book succesfully deleted",
	})
	return nil
}

func (r *Repository) DeleteUser(context *fiber.Ctx) error {
	userModel := models.Users{}
	id := context.Params("id")
	if id == "" {
		r.Logger.Error("id cannot be empty for DeleteUser")
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty"})
		return nil
	}

	err := r.DB.Delete(userModel, id).Error

	if err != nil {
		r.Logger.WithError(err).Error("could not delete book for DeleteBook")
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not delete user",
		})
		return err
	}
	r.Logger.Info("User successfully deleted", logrus.Fields{"userID": id})
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "user successfully deleted",
	})
	return nil
}

func (r *Repository) GetBooks(context *fiber.Ctx) error {
	bookModels := &[]models.Books{}

	err := r.DB.Preload("User").Find(bookModels).Error
	if err != nil {
		r.Logger.WithError(err).Error("failed to fetch books")
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get books"})
		return err
	}

	r.Logger.WithFields(logrus.Fields{
		"bookCount": len(*bookModels),
	}).Info("books fetched successfully")

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
		r.Logger.WithError(err).Error("failed to fetch users")
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get users"})
		return err
	}
	r.Logger.WithFields(logrus.Fields{
		"bookCount": len(*userModels),
	}).Info("users fetched successfully")

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
		r.Logger.WithField("id", id).Error("ID cannot be empty")
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id can not be empty",
		})
		return nil
	}
	fmt.Println("the ID is ", id)

	err := r.DB.Preload("User").Where("id = ?", id).First(bookModel).Error
	if err != nil {
		r.Logger.WithError(err).WithField("id", id).Error("Failed to get book by ID")
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "can not get book by id",
		})
		return err
	}
	r.Logger.WithField("id", id).Info("Book ID fetched successfully")
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
		r.Logger.WithField("id", id).Error("ID cannot be empty")
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}

	err := r.DB.Where("id = ?", id).First(bookModel).Error
	if err != nil {
		r.Logger.WithError(err).WithField("id", id).Error("Failed to get book by ID")
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "can not get book by id",
		})
		return err
	}

	newBook := &models.Books{}
	err = context.BodyParser(&newBook)
	if err != nil {
		r.Logger.WithError(err).Error("Failed to parse request body")
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	bookModel.Author = newBook.Author
	bookModel.Title = newBook.Title
	bookModel.Publisher = newBook.Publisher
	bookModel.UserID = newBook.UserID

	err = r.DB.Save(bookModel).Error
	if err != nil {
		r.Logger.WithError(err).Error("Failed to update book")
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not update book"})
		return err
	}
	r.Logger.WithField("id", id).Info("Book successfully updated")
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
		r.Logger.WithField("id", id).Error("ID cannot be empty")
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}

	err := r.DB.Where("id = ?", id).Preload("Books").First(userModel).Error
	if err != nil {
		r.Logger.WithError(err).WithField("id", id).Error("Failed to get user by ID")
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not get user by id",
		})
		return err
	}
	r.Logger.WithField("id", id).Info("User ID fetched successfully")
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
		r.Logger.WithField("id", id).Error("ID cannot be empty")
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}

	err := r.DB.Where("user_id = ?", id).Find(userBooks).Error
	if err != nil {
		r.Logger.WithError(err).WithField("id", id).Error("Failed to get book by user ID")
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

func LoggerMiddleware(logger *logrus.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		logger.WithFields(logrus.Fields{
			"method":  c.Method(),
			"path":    c.Path(),
			"status":  c.Response().StatusCode(),
			"latency": latency,
		}).Info("request processed")

		return err
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

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
		logger.WithError(err).Fatal("Could not load the database")
	}

	// err = models.MigrateBooks(db)
	// if err != nil {
	// 	log.Fatal("could not migrate db")
	// }

	r := Repository{
		DB:     db,
		Logger: logger,
	}

	app := fiber.New(fiber.Config{
		ReadTimeout: 3 * time.Second,
	})

	app.Use(LoggerMiddleware(logger))
	r.SetupRoutes(app)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGHUP)
	go func() {
		if err := app.Listen(":8080"); err != nil {
			logger.WithError(err).Fatal("Error starting server")
		}
	}()

	fmt.Println("Server is running on :8080")
	fmt.Println("Process id is", os.Getpid())

	select {
	case <-stopChan:
		fmt.Println("Received termination signal. Shutting down...")
		_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := app.Shutdown(); err != nil {
			logger.WithError(err).Error("Server shutdown failed")
		} else {
			logger.Info("Server shutdown gracefully")
		}
		time.Sleep(500 * time.Millisecond)
	}

}
