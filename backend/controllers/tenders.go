package controllers

import (
	"ZADANIE-6105/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateTender(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)

	var tender models.Tender

	if err := c.ShouldBindJSON(&tender); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.Create(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tender"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tender created successfully", "tender": tender})
}

// GetTenders возвращает список тендеров с возможностью фильтрации по типу услуг
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

// GetUserTenders возвращает список тендеров текущего пользователя
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
	username := c.Query("username")

	// Поиск тендера по идентификатору
	var tender models.Tender
	if err := database.First(&tender, "id = ?", tenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tender not found"})
		return
	}

	// Проверка, что переданный username соответствует создателю тендера
	if tender.CreatorUsername != username {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to update this tender"})
		return
	}

	// Получение новых данных для обновления
	var updatedTender struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ServiceType string `json:"serviceType"`
	}
	if err := c.ShouldBindJSON(&updatedTender); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновление полей тендера (разрешенные поля)
	if updatedTender.Name != "" {
		tender.Name = updatedTender.Name
	}
	if updatedTender.Description != "" {
		tender.Description = updatedTender.Description
	}
	if updatedTender.ServiceType != "" {
		tender.ServiceType = updatedTender.ServiceType
	}

	// Сохранение обновленного тендера в базе данных
	if err := database.Save(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tender"})
		return
	}

	c.JSON(http.StatusOK, tender)
}

func GetTenderStatus(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)

	// Получение идентификатора тендера из параметров запроса
	tenderID := c.Param("tenderId")

	// Получение username из query parameters (необязательно)
	username := c.Query("username")

	// Поиск тендера по идентификатору
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

	// Возвращение статуса тендера
	c.JSON(http.StatusOK, tender.Status)
}
