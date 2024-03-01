package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"github.com/dgrijalva/jwt-go"
	"os/signal"
	"github.com/gofiber/template/html/v2"
	"syscall"
	"time"
	"golang.org/x/crypto/bcrypt"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"strconv"
	"golang.org/x/time/rate"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/yerlan/go-fiber-postgres/models"

	"github.com/yerlan/go-fiber-postgres/storage"
	"gorm.io/gorm"
	"errors"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
	UserID    uint   `json:"user_id"`
}

type User struct {
	ID uint `json:id`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role string `json:"role"`
	Password string `json:"password"`
	Age   int    `json:"age"`
}

type Repository struct {
	DB     *gorm.DB
	Logger *logrus.Logger
	JWTSecret       string 
   ProtectedResource func(r *Repository, context *fiber.Ctx) error
}


func (r *Repository) ProfileHandler(context *fiber.Ctx) error {
    
    tokenString := context.Cookies("jwt")

    if tokenString == "" {
        
        return context.Redirect("/api/login")
    }

    // Parse the JWT token
    token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(r.JWTSecret), nil
    })

    if err != nil || !token.Valid {
        // Redirect to login page if the token is invalid
        return context.Redirect("/api/login")
    }

    // Extract user information from the token claims
    claims, ok := token.Claims.(*jwt.MapClaims)
    if !ok {
        return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse token claims"})
    }

	fmt.Println(claims)

    user := User{
        Name:  (*claims)["name"].(string),
        Email: (*claims)["email"].(string),
        // Role:  (*claims)["role"].(string),
        // Add other user information as needed
    }
	fmt.Println(user)


    // Render the profile page with user data using HTML template
    return context.Render("profile", fiber.Map{"user": user})
}



func (r *Repository) AdminHandler(context *fiber.Ctx) error {
    // Check if the user is authenticated
 
    users, err := r.GetAllUsers()
	
    if err != nil {
        return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch users"})
    }

    // Render the admin page with the list of users
    return context.Render("admin", fiber.Map{"users": users})
}

func (r *Repository) GetAllUsers() ([]User, error) {
    // Create a slice to store the users
    var users []User

    // Use GORM's Find method to fetch all users
    if err := r.DB.Find(&users).Error; err != nil {
        return nil, err
    }

    return users, nil
}



func (r *Repository) ProtectedBooks(context *fiber.Ctx) error {
	return context.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Protected /books resource accessed"})
 }

func (r *Repository) Signup(context *fiber.Ctx) error {
    user := User{}
    if err := context.BodyParser(&user); err != nil {
        return context.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error hashing password"})
    }
    user.Password = string(hashedPassword)

    if err := r.DB.Create(&user).Error; err != nil {
        return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error creating user"})
    }

	if err := r.sendWelcomeEmail(user.Email); err != nil {
        r.Logger.WithError(err).Error("Error sending welcome email")
    }

    return context.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User created successfully"})
}

func (r *Repository) sendWelcomeEmail(toEmail string) error {
    auth := smtp.PlainAuth(
        "",
        "egisbekovim0@gmail.com",
		"bnku nhpu rmic xbkb",     
        "smtp.gmail.com",
    )

    subject := "Welcome to Your App"
    body := "Thank you for signing up! We're excited to have you on board."

    msg := fmt.Sprintf("Subject: %s\n%s", subject, body)

    err := smtp.SendMail(
        "smtp.gmail.com:587",
        auth,
        "egisbekovim0@gmail.com",   // replace with your SMTP email
        []string{toEmail},
        []byte(msg),
    )

    return err
}

func (r *Repository) SendEmail(context *fiber.Ctx) error {
    emailRequest := struct {
        ToEmail string `json:"to_email"`
    }{}

	fmt.Println(emailRequest.ToEmail)

    if err := context.BodyParser(&emailRequest); err != nil {
        return context.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }

    if err := r.sendWelcomeEmail(emailRequest.ToEmail); err != nil {
        return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error sending email"})
    }

    return context.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Email sent successfully"})
}

func (r *Repository) Login(context *fiber.Ctx) error {
    loginRequest := new(struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    })

    if err := context.BodyParser(loginRequest); err != nil {
        return context.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }

    user := User{}
    if err := r.DB.Where("email = ?", loginRequest.Email).First(&user).Error; err != nil {
        return context.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
        return context.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
    }

	token, err := GenerateJWT(&user)
    if err != nil {
        return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error generating JWT"})
    }


	context.Cookie(&fiber.Cookie{
        Name:  "jwt",
        Value: token,
    })


    return context.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Login successful"})
}

func GenerateJWT(user *User) (string, error) {
    token := jwt.New(jwt.SigningMethodHS256)
    claims := token.Claims.(jwt.MapClaims)
    claims["email"] = user.Email
	claims["name"] = user.Name
	// claims["role"] = user.Role

    tokenString, err := token.SignedString([]byte("your-secret-key"))
    if err != nil {
        return "", err
    }

    return tokenString, nil
}


func Authenticate() fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := c.Get("Authorization")
        if token == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
        }

        claims := jwt.MapClaims{}
        _, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
            return []byte("your-secret-key"), nil
        })

        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
        }

        return c.Next()
    }
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


const defaultPageSize = 5
var limiter = rate.NewLimiter(1, 1)

func (r *Repository) GetBooks(context *fiber.Ctx) error {
	if !limiter.Allow() {
		context.Status(http.StatusTooManyRequests).JSON(&fiber.Map{
		  "message": "To Many Request",
		})
		return nil
	  }

    page, err := strconv.Atoi(context.Query("page", "1"))
    if err != nil || page <= 0 {
        page = 1
    }

    pageSize, err := strconv.Atoi(context.Query("pageSize", strconv.Itoa(defaultPageSize)))
    if err != nil || pageSize <= 0 {
        pageSize = defaultPageSize
    }

    authorFilter := context.Query("author")
    titleFilter := context.Query("title")
    sortBy := context.Query("sortBy", "id")
    sortOrder := context.Query("sortOrder", "asc")

    bookModels := &[]models.Books{}

    query := r.DB.Model(&models.Books{})

    if authorFilter != "" {
        query = query.Where("author LIKE ?", "%"+authorFilter+"%")
    }

    if titleFilter != "" {
        query = query.Where("title LIKE ?", "%"+titleFilter+"%")
    }

    query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder)).
        Limit(pageSize).Offset((page - 1) * pageSize).
        Preload("User").Find(bookModels)

    if err := query.Error; err != nil {
        context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "could not get books"})
        return err
    }

    context.Status(http.StatusOK).JSON(&fiber.Map{
        "message": "books fetched successfully",
        "data":    bookModels,
    })

    return nil
}


func (r *Repository) GetUsers(context *fiber.Ctx) error {
	if !limiter.Allow() {
		context.Status(http.StatusTooManyRequests).JSON(&fiber.Map{
		  "message": "Too Many Requests",
		})
		return nil
	}
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
	api.Post("/signup", r.Signup)
    api.Post("/login", r.Login)
	app.Get("/profile", r.ProfileHandler)
	app.Get("/admin", r.AdminHandler)
	app.Post("/sendEmail", r.SendEmail)
	// protected := api.Use(Authenticate())

	// protected.Get("/books", r.ProtectedBooks)
}
var loge = logrus.New()

func init() {

	loge.SetFormatter(&logrus.JSONFormatter{})

	loge.SetLevel(logrus.DebugLevel)
  }

func LoggerMiddleware(logger *logrus.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		loge.WithFields(logrus.Fields{
			"method":  c.Method(),
			"path":    c.Path(),
			"status":  c.Response().StatusCode(),
			"latency": latency,
		}).Info("request processed")

		return err
	}
}


func main() {
	ctx := "LogrusToLogFile"
	loge.Out = os.Stdout
	file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		loge.Out = file
	   } else {
		 loge.Info("Failed to log to file, using default stderr")
	   }
	err = errors.New("math: square root of negative number")
	   if err != nil {
		 loge.WithFields(logrus.Fields{
		   "ctx": ctx,
		 }).Error("Write to file")
	   }
	err1 := godotenv.Load(".env")
	if err1 != nil {
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
		JWTSecret:       "your-secret-key", 
		ProtectedResource: (*Repository).ProtectedBooks, 
	}

	engine := html.NewFileSystem(http.Dir("../frontend"), ".html")

  // Reload the templates on each render, good for development
	engine.Reload(true)

	// Debug will print each template that is parsed, good for debugging
	engine.Debug(true) // Optional. Default: false

	// Layout defines the variable name that is used to yield templates within layouts
	engine.Layout("embed") 

	engine.Delims("{{", "}}") 

	app := fiber.New(fiber.Config{
		Views: engine,
		ReadTimeout: 3 * time.Second,
	})

	app.Get("/login", func(c *fiber.Ctx) error {
        return c.SendFile("../frontend/signin.html")
    })

	// app.Get("/profile", func(c *fiber.Ctx) error {
    //     return c.SendFile("../frontend/profile.html")
    // })

    app.Get("/signup", func(c *fiber.Ctx) error {
        return c.SendFile("../frontend/signup.html")
    })

	// for i := 0; i < 100; i++ {
	// 	book := Book{}
	// 	err := faker.FakeData(&book)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	if err := db.Create(&book).Error; err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Printf("Added Book: %s\n", book.Title)
	// }

	// app.Get("/login", func(c *fiber.Ctx) error {
    //     return c.Render("login", nil)
    // })

    // app.Get("/signup", func(c *fiber.Ctx) error {
    //     return c.Render("signup", nil)
    // })


	app.Use(cors.New(cors.Config{
        AllowOrigins: "*",
        AllowMethods: "GET,POST,PUT,DELETE",
        AllowHeaders: "Origin, Content-Type, Accept",
    }))
	app.Static("/", "../frontend")
	app.Static("/сss", "../frontend/css")
	app.Static("/assets3", "../frontend/assets3")

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
