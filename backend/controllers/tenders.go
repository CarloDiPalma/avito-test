package controllers

import (
	"ZADANIE-6105/models"
	"ZADANIE-6105/schemas"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Проверка валидности пользователя по имени пользователя
func isValidUser(db *gorm.DB, username string) bool {
	var employee models.Employee
	result := db.Where("username = ?", username).First(&employee)

	// Если результат не найден или произошла ошибка, пользователь невалиден
	return result.Error == nil && result.RowsAffected > 0
}

// Проверка, является ли пользователь ответственным за организацию
func isResponsibleForOrganization(db *gorm.DB, organizationID uuid.UUID, employeeID uuid.UUID) bool {
	var responsibility models.OrganizationResponsible
	result := db.Where("organization_id = ? AND employee_id = ?", organizationID, employeeID).First(&responsibility)
	return result.Error == nil && result.RowsAffected > 0
}

func CreateTender(c *gin.Context) {
	var tender models.Tender

	// Извлечение данных из тела запроса
	if err := c.ShouldBindJSON(&tender); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка, что creatorUsername не пуст
	if tender.CreatorUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "creatorUsername cannot be empty"})
		return
	}

	// Получите доступ к базе данных
	db := c.MustGet("db").(*gorm.DB)

	// Получите ID сотрудника по имени пользователя
	var employee models.Employee
	result := db.Where("username = ?", tender.CreatorUsername).First(&employee)
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	// Проверьте, является ли сотрудник ответственным за организацию
	if !isResponsibleForOrganization(db, tender.OrganizationID, employee.ID) {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not responsible for the organization"})
		return
	}

	if err := tender.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Create(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tender"})
		return
	}

	response := schemas.CreateTenderResponse{
		ID:          tender.ID,
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
		Status:      tender.Status,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
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

	// Поиск тендера по ID
	var tender models.Tender
	if err := database.First(&tender, "id = ?", tenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tender not found"})
		return
	}

	// Проверка имени пользователя
	username := c.Query("username")
	if username != "" && tender.CreatorUsername != username {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to update this tender"})
		return
	}

	// Обновление только разрешенных полей
	var updatedFields struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		ServiceType *string `json:"serviceType"`
	}

	if err := c.ShouldBindJSON(&updatedFields); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Сохранение текущего состояния тендера в историю
	history := models.TenderHistory{
		TenderID:        tender.ID,
		Name:            tender.Name,
		Description:     tender.Description,
		ServiceType:     tender.ServiceType,
		Status:          tender.Status,
		Version:         tender.Version,
		CreatorUsername: tender.CreatorUsername,
		OrganizationID:  tender.OrganizationID,
		CreatedAt:       tender.CreatedAt,
	}

	// Сохранение истории тендера
	if err := database.Create(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save tender history"})
		return
	}

	// Обновление полей тендера
	if updatedFields.Name != nil {
		tender.Name = *updatedFields.Name
	}
	if updatedFields.Description != nil {
		tender.Description = *updatedFields.Description
	}
	if updatedFields.ServiceType != nil {
		tender.ServiceType = *updatedFields.ServiceType
	}
	tender.Version++

	// Сохранение обновленного тендера
	if err := database.Save(&tender).Error; err != nil {
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

	if err := database.Save(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tender status"})
		return
	}

	c.JSON(http.StatusOK, tender)
}

func RollbackTender(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)

	tenderID := c.Param("tenderId")
	versionStr := c.Param("version")

	// Преобразование версии в int
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version"})
		return
	}

	// Проверка имени пользователя
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	// Поиск тендера по ID
	var tender models.Tender
	if err := database.First(&tender, "id = ?", tenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tender not found"})
		return
	}

	// Проверка прав на изменение тендера
	if tender.CreatorUsername != username {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to rollback this tender"})
		return
	}

	// Поиск версии в истории
	var history models.TenderHistory
	if err := database.First(&history, "tender_id = ? AND version = ?", tenderID, version).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tender version not found"})
		return
	}

	// Сохранение текущего состояния тендера в историю
	currentHistory := models.TenderHistory{
		TenderID:        tender.ID,
		Name:            tender.Name,
		Description:     tender.Description,
		ServiceType:     tender.ServiceType,
		Status:          tender.Status,
		Version:         tender.Version,
		CreatorUsername: tender.CreatorUsername,
		OrganizationID:  tender.OrganizationID,
		CreatedAt:       tender.CreatedAt,
	}

	if err := database.Create(&currentHistory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save current tender history"})
		return
	}

	// Обновление тендера данными из истории
	tender.Name = history.Name
	tender.Description = history.Description
	tender.ServiceType = history.ServiceType
	tender.Status = history.Status
	tender.Version++

	// Сохранение обновленного тендера
	if err := database.Save(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tender"})
		return
	}

	c.JSON(http.StatusOK, tender)
}
