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
	// loadEnvFile := os.Getenv("LOAD_ENV_FILE")
	// if loadEnvFile == "" {
	// 	loadEnvFile = "false"
	// }

	// if loadEnvFile == "true" {
	// 	err := godotenv.Load()
	// 	if err != nil {
	// 		log.Fatalf("Error loading .env file")
	// 	}
	// }

	postgresConn := os.Getenv("POSTGRES_CONN")
	if postgresConn == "" {
		postgresConn = "postgres://cnrprod1725742191-team-77945:cnrprod1725742191-team-77945@rc1b-5xmqy6bq501kls4m.mdb.yandexcloud.net:6432/cnrprod1725742191-team-77945?target_session_attrs=read-write"
		// postgresConn = "postgres://postgres:4824@localhost:5432/avito?sslmode=disable"
	}

	// Выполнение миграций
	migrations.RunMigrations(postgresConn)
	db, err := gorm.Open(postgres.Open(postgresConn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Инициализация и запуск сервера
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "PUT"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
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
