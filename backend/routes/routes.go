package routes

import (
	"ZADANIE-6105/controllers"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes инициализирует все маршруты
func SetupRoutes(router *gin.Engine, db *gorm.DB) {

	router.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	api := router.Group("/api")
	{
		api.POST("/tenders/new", controllers.CreateTender)
		api.GET("/tenders", controllers.GetTenders)
		api.GET("/tenders/my", controllers.GetUserTenders)
		api.POST("/employees/new", controllers.CreateEmployee)
		api.PATCH("/tenders/:tenderId/edit", controllers.UpdateTender)
		api.GET("/tenders/:tenderId/status", controllers.GetTenderStatus)
		api.PUT("/tenders/:tenderId/status", controllers.UpdateTenderStatus)
		api.PUT("/tenders/:tenderId/rollback/:version", controllers.RollbackTender)
		api.POST("/bids/new", controllers.CreateBid)
		api.GET("/bids/my", controllers.GetMyBids)
		api.GET("/bids/:tenderId/:action", func(c *gin.Context) {
			action := c.Param("action")

			if action == "list" {
				// Вызов GetBidsByTender без явного использования переменной tenderId
				controllers.GetBidsByTender(c)
			} else if action == "status" {
				bidId := c.Param("tenderId") // Используем параметр как bidId для статуса
				c.Set("bidId", bidId)
				controllers.GetBidStatus(c)
			} else {
				c.JSON(404, gin.H{"reason": "Not found"})
			}
		})
		api.PUT("/bids/:bidId/status", controllers.UpdateBidStatus)
		api.PATCH("/bids/:bidId/edit", controllers.EditBid)
		api.PUT("/bids/:bidId/rollback/:version", controllers.RollbackBid)
		api.PUT("/bids/:bidId/submit_decision", controllers.SubmitDecision)
		api.PUT("/bids/:bidId/feedback", controllers.SendFeedback)
		api.GET("/bids/:tenderId/reviews", controllers.GetBidReviews)
	}
}
