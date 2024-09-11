package controllers

import (
	"net/http"
	"strconv"
	"time"

	"ZADANIE-6105/models"
	"ZADANIE-6105/schemas"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateBid(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)

	// Входные данные с использованием схемы BidCreateRequest
	var bidInput schemas.BidCreateRequest

	if err := c.ShouldBindJSON(&bidInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка существования тендера по TenderID
	tenderID, err := uuid.Parse(bidInput.TenderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid TenderID format"})
		return
	}

	var tender models.Tender
	if err := database.First(&tender, "id = ?", tenderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"reason": "Tender with the given ID does not exist"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"reason": "Database error while checking tender"})
		}
		return
	}

	// Проверка существования пользователя по AuthorID
	var employee models.Employee
	if err := database.Where("id = ?", bidInput.AuthorID).First(&employee).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Unauthorized, user does not exist"})
		return
	}

	// Проверка прав с использованием модели OrganizationResponsible
	if bidInput.AuthorType == "Organization" {
		var orgResponsible models.OrganizationResponsible
		if err := database.Where("organization_id = ? AND employee_id = ?", tender.OrganizationID, employee.ID).First(&orgResponsible).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusForbidden, gin.H{"reason": "Unauthorized to create bid as Organization"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"reason": "Database error while checking organization responsible"})
			return
		}
	}

	// Создание нового предложения
	bid := models.Bid{
		Name:        bidInput.Name,
		Description: bidInput.Description,
		TenderID:    tenderID,
		AuthorType:  bidInput.AuthorType,
		AuthorID:    employee.ID,
		Version:     1, // Начальная версия
		CreatedAt:   time.Now(),
	}

	if err := database.Create(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to create bid"})
		return
	}

	// Формирование ответа
	response := schemas.BidCreateResponse{
		ID:         bid.ID,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorID:   bid.AuthorID,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

func GetMyBids(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)

	// Получение username из query
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	// Проверка, существует ли пользователь с данным username
	var employee models.Employee
	if err := database.Where("username = ?", username).First(&employee).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid username"})
		return
	}

	// Параметры пагинации (по умолчанию limit = 5, offset = 0)
	limit := 5
	offset := 0

	if l, err := strconv.Atoi(c.DefaultQuery("limit", "5")); err == nil {
		limit = l
	}
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil {
		offset = o
	}

	// Получение списка предложений текущего пользователя
	var bids []models.Bid
	if err := database.Where("author_id = ?", employee.ID).Limit(limit).Offset(offset).Find(&bids).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve bids"})
		return
	}

	// Преобразуем bids в BidCreateResponse
	var responseBids []schemas.BidCreateResponse
	for _, bid := range bids {
		responseBids = append(responseBids, schemas.BidCreateResponse{
			ID:         bid.ID,
			Name:       bid.Name,
			Status:     bid.Status,
			AuthorID:   bid.AuthorID,
			AuthorType: bid.AuthorType,
			Version:    bid.Version,
			CreatedAt:  bid.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, responseBids)
}

func GetBidsByTender(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)

	// Получение идентификатора тендера из параметров запроса
	tenderID := c.Param("tenderId")
	if tenderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tender ID is required"})
		return
	}

	// Проверка существования тендера
	var tender models.Tender
	if err := database.First(&tender, "id = ?", tenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tender not found"})
		return
	}

	// Получение username из query
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	// Проверка существования пользователя
	var employee models.Employee
	if err := database.First(&employee, "username = ?", username).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	// Проверка прав доступа: пользователь должен быть создателем тендера
	if tender.CreatorUsername != username {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized to access bids for this tender"})
		return
	}

	// Параметры пагинации (по умолчанию limit = 5, offset = 0)
	limit := 5
	offset := 0

	if l, err := strconv.Atoi(c.DefaultQuery("limit", "5")); err == nil {
		limit = l
	}
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil {
		offset = o
	}

	// Получение списка предложений, связанных с указанным тендером
	var bids []models.Bid
	if err := database.Where("tender_id = ?", tenderID).Limit(limit).Offset(offset).Find(&bids).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve bids"})
		return
	}

	// Преобразование bids в BidCreateResponse
	var responseBids []schemas.BidCreateResponse
	for _, bid := range bids {
		responseBids = append(responseBids, schemas.BidCreateResponse{
			ID:         bid.ID,
			Name:       bid.Name,
			Status:     bid.Status,
			AuthorType: bid.AuthorType,
			AuthorID:   bid.AuthorID,
			Version:    bid.Version,
			CreatedAt:  bid.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, responseBids)
}

func GetBidStatus(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)

	// Получение идентификатора предложения из параметров запроса
	bidID := c.Param("bidId")

	// Получение username из query
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	// Проверка существования пользователя
	var employee models.Employee
	if err := database.First(&employee, "username = ?", username).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	// Поиск предложения по ID
	var bid models.Bid
	if err := database.First(&bid, "id = ?", bidID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bid not found"})
		return
	}

	// Формирование ответа
	response := gin.H{
		"id":        bid.ID,
		"status":    bid.Status,
		"version":   bid.Version,
		"createdAt": bid.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}
