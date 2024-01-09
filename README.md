# Book Reservation Web Application

## Overview
This web application is expertly crafted for the efficient management of a comprehensive book collection and user profiles. It's an all-encompassing solution that allows users to seamlessly add, edit, and delete books from the collection. Additionally, users can create and manage their own profiles, ensuring a personalized and user-friendly experience.

## Team
Developed by the talented trio: Amangeldi Diyar, Egisbekov Erlan, and Raisov Raiymbek, this application represents a collaborative effort combining expertise and passion for technology.

## Getting Started
To set up the project, follow these steps:

1. **Create Project Folder**: Begin by creating your project folder and initialize it using Go modules:
`go mod init github`
2. **Install Dependencies**: Next, install the necessary Go packages:
`go get gorm.io`
`go get -u github.com/gofiber/fiber/v2`
`go get -u github.com/joho/godotenv`

3. **Set Up the Environment**: In Visual Studio Code, create the folders 'models', 'storage', and a '.env' file. In the '.env' file, include all your database details. Then, create a 'main.go' file.
4. **Develop Models and Storage**: In the 'models' folder, start by creating your book and user models. You can choose to have both models in one file or separate them. In the 'storage' folder, create 'postgres.go' for managing database connections.

## Key Technologies
- **Fiber**: A Go web framework for swiftly building HTTP servers. [Fiber GitHub Repository](https://github.com/gofiber/fiber)
- **Gorm**: An ORM for Go, used for database operations. [Gorm Official Website](https://gorm.io/)
- **Godotenv**: A library for loading environment variables from a '.env' file. [Godotenv GitHub Repository](https://github.com/joho/godotenv)
- **Golang-Migrate**: A tool for managing database migrations in Go. [Golang-Migrate GitHub Repository](https://github.com/golang-migrate/migrate)

Embark on your journey to manage and organize books efficiently with our Book Reservation Web Application. Happy Coding!
