package handlers

import (
	"ZADANIE-6105/models"
	"ZADANIE-6105/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateEmployee(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}

	var employee models.Employee
	if err := c.ShouldBindJSON(&employee); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.Create(&employee).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Employee created successfully", "employee": employee})
}
