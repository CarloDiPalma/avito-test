package routes

import (
	"ZADANIE-6105/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {

	router.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	api := router.Group("/api")
	{
		api.POST("/tenders/new", handlers.CreateTender)
		api.GET("/tenders", handlers.GetTenders)
		api.GET("/tenders/my", handlers.GetUserTenders)
		api.POST("/employees/new", handlers.CreateEmployee)
		api.PATCH("/tenders/:tenderId/edit", handlers.UpdateTender)
		api.GET("/tenders/:tenderId/status", handlers.GetTenderStatus)
		api.PUT("/tenders/:tenderId/status", handlers.UpdateTenderStatus)
		api.PUT("/tenders/:tenderId/rollback/:version", handlers.RollbackTender)
		api.POST("/bids/new", handlers.CreateBid)
		api.GET("/bids/my", handlers.GetMyBids)
		api.GET("/bids/:tenderId/:action", func(c *gin.Context) {
			action := c.Param("action")

			if action == "list" {
				handlers.GetBidsByTender(c)
			} else if action == "status" {
				bidId := c.Param("tenderId")
				c.Set("bidId", bidId)
				handlers.GetBidStatus(c)
			} else {
				c.JSON(404, gin.H{"reason": "Not found"})
			}
		})
		api.PUT("/bids/:bidId/status", handlers.UpdateBidStatus)
		api.PATCH("/bids/:bidId/edit", handlers.EditBid)
		api.PUT("/bids/:bidId/rollback/:version", handlers.RollbackBid)
		api.PUT("/bids/:bidId/submit_decision", handlers.SubmitDecision)
		api.PUT("/bids/:bidId/feedback", handlers.SendFeedback)
		api.GET("/bids/:tenderId/reviews", handlers.GetBidReviews)
	}
}
