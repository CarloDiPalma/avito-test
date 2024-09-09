package controllers

import (
	"ZADANIE-6105/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateTender(c *gin.Context) {
	var tender models.Tender

	if err := c.ShouldBindJSON(&tender); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := tender.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	if err := db.Create(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tender"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tender created successfully", "tender": tender})
}

func GetTenders(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	serviceType := c.Query("serviceType")
	var tenders []models.Tender

	query := db.Model(&models.Tender{})
	if serviceType != "" {
		query = query.Where("service_type = ?", serviceType)
	}

	if err := query.Find(&tenders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tenders)
}

func GetUserTenders(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	var user models.Employee
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var tenders []models.Tender
	if err := db.Where("creator_username = ?", username).Find(&tenders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tenders)
}

func UpdateTender(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)
	tenderID := c.Param("tenderId")

	var tender models.Tender
	if err := database.First(&tender, "id = ?", tenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tender not found"})
		return
	}

	username := c.Query("username")
	if username != "" && tender.CreatorUsername != username {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to update this tender"})
		return
	}

	var updatedFields map[string]interface{}
	if err := c.ShouldBindJSON(&updatedFields); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tender.Version++
	tender.UpdatedAt = time.Now()

	if err := database.Model(&tender).Updates(updatedFields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tender"})
		return
	}

	c.JSON(http.StatusOK, tender)
}

func GetTenderStatus(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)
	tenderID := c.Param("tenderId")
	username := c.Query("username")

	var tender models.Tender
	if err := database.First(&tender, "id = ?", tenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tender not found"})
		return
	}

	// (Необязательно) Можно добавить логику проверки username, если это необходимо
	if username != "" && tender.CreatorUsername != username {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to view this tender"})
		return
	}

	c.JSON(http.StatusOK, tender.Status)
}

func UpdateTenderStatus(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)
	tenderID := c.Param("tenderId")

	var tender models.Tender
	if err := database.First(&tender, "id = ?", tenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tender not found"})
		return
	}

	username := c.Query("username")
	if username != "" && tender.CreatorUsername != username {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to update this tender"})
		return
	}

	newStatus := c.Query("status")
	if newStatus == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	tender.Status = newStatus
	tender.UpdatedAt = time.Now()

	if err := database.Save(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tender status"})
		return
	}

	c.JSON(http.StatusOK, tender)
}
