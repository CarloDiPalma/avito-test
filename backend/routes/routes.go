package routes

import (
	"ZADANIE-6105/controllers"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes инициализирует все маршруты
func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Пинг эндпоинт
	router.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	api := router.Group("/api")
	{
		api.POST("/tenders/new", controllers.CreateTender)
		api.GET("/tenders", controllers.GetTenders)
		api.GET("/tenders/my", controllers.GetUserTenders)
		api.POST("/employees/new", controllers.CreateEmployee)
		api.PATCH("/api/tenders/:tenderId/edit", controllers.UpdateTender)
		api.GET("/tenders/:tenderId/status", controllers.GetTenderStatus)
		api.PUT("/tenders/:tenderId/status", controllers.UpdateTenderStatus)
	}
}
