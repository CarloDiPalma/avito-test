package main

import (
	"ZADANIE-6105/migrations"
	"ZADANIE-6105/routes"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Выполнение миграций
	migrations.RunMigrations()
	postgresConn := os.Getenv("POSTGRES_CONN")
	if postgresConn == "" {
		postgresConn = "postgres://postgres:7744@localhost:5432/avito?sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(postgresConn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Инициализация и запуск сервера
	r := gin.Default()

	// Настройка CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"https://editor.swagger.io"},
		AllowMethods: []string{"GET", "POST", "PATCH", "DELETE", "PUT"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))

	// Middleware для установки базы данных в контекст
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Настройка маршрутов
	routes.SetupRoutes(r, db)

	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = "0.0.0.0:8080"
	}
	if err := r.Run(serverAddress); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
