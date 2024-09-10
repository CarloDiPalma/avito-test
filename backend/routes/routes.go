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
		// api.GET("/bids/:tenderId/list", controllers.GetBidsByTender)
		// api.GET("/bids/:bidId/status", controllers.GetBidStatus)
		// subResources := api.GET("/bids/:bidId")
		// {
		// 	subResources.GET("/status", controllers.GetBidStatus)
		// 	// subResources.GET("/ssub-resources", GetSSubResources)
		// 	// subResources.GET("/ssub-resources/:ssrid", GetSSubResourcesByID)
		// }
		api.GET("/bids/:id/:action", func(c *gin.Context) {
			id := c.Param("id")
			action := c.Param("action")

			if action == "list" {
				c.Set("tenderId", id)
				controllers.GetBidsByTender(c)
			} else if action == "status" {
				c.Set("bidId", id)
				controllers.GetBidStatus(c)
			} else {
				c.JSON(404, gin.H{"message": "Not found"})
			}
		})

	}
}
