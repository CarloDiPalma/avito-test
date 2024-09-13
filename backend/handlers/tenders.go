package handlers

import (
	"ZADANIE-6105/models"
	"ZADANIE-6105/schemas"
	"ZADANIE-6105/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func isResponsibleForOrganization(db *gorm.DB, organizationID uuid.UUID, userID uuid.UUID) bool {
	var responsibility models.OrganizationResponsible
	result := db.Where("organization_id = ? AND user_id = ?", organizationID, userID).First(&responsibility)
	return result.Error == nil && result.RowsAffected > 0
}

func CreateTender(c *gin.Context) {
	var tender models.Tender

	if err := c.ShouldBindJSON(&tender); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
		return
	}

	if tender.CreatorUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "creatorUsername cannot be empty"})
		return
	}

	db, ok := utils.GetDB(c)
	if !ok {
		return
	}

	var employee models.Employee
	result := db.Where("username = ?", tender.CreatorUsername).First(&employee)
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	if !isResponsibleForOrganization(db, tender.OrganizationID, employee.ID) {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not responsible for the organization"})
		return
	}

	if err := tender.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
		return
	}

	if err := db.Create(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to create tender"})
		return
	}

	response := schemas.TenderResponse{
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
	db, ok := utils.GetDB(c)
	if !ok {
		return
	}

	validServiceTypes := map[string]bool{
		"Construction": true,
		"Delivery":     true,
		"Manufacture":  true,
	}

	serviceTypes := c.QueryArray("service_type")
	limitStr := c.DefaultQuery("limit", "5")
	offsetStr := c.DefaultQuery("offset", "0")

	for _, st := range serviceTypes {
		if !validServiceTypes[st] {
			c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid service_type value"})
			return
		}
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var tenders []models.Tender

	query := db.Model(&models.Tender{})

	if len(serviceTypes) > 0 {
		query = query.Where("service_type IN ?", serviceTypes)
	}

	query = query.Order("name ASC").Limit(limit).Offset(offset)

	if err := query.Find(&tenders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	var responses []schemas.TenderResponse
	for _, tender := range tenders {
		responses = append(responses, schemas.TenderResponse{
			ID:          tender.ID,
			Name:        tender.Name,
			Description: tender.Description,
			ServiceType: tender.ServiceType,
			Status:      tender.Status,
			Version:     tender.Version,
			CreatedAt:   tender.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, responses)
}

func GetUserTenders(c *gin.Context) {
	db, ok := utils.GetDB(c)
	if !ok {
		return
	}

	username := c.Query("username")

	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "username parameter cannot be empty"})
		return
	}

	var employee models.Employee
	result := db.Where("username = ?", username).First(&employee)
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	var tenders []models.Tender

	query := db.Model(&models.Tender{})
	query = query.Where("creator_username = ?", username)

	limitStr := c.DefaultQuery("limit", "5")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	query = query.Order("name ASC").Limit(limit).Offset(offset)

	if err := query.Find(&tenders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	var responses []schemas.TenderResponse
	for _, tender := range tenders {
		responses = append(responses, schemas.TenderResponse{
			ID:          tender.ID,
			Name:        tender.Name,
			Description: tender.Description,
			ServiceType: tender.ServiceType,
			Status:      tender.Status,
			Version:     tender.Version,
			CreatedAt:   tender.CreatedAt.Format(time.RFC3339),
		})
	}

	if len(responses) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"reason": "No tenders found"})
		return
	}

	c.JSON(http.StatusOK, responses)
}

func UpdateTender(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}
	tenderID := c.Param("tenderId")

	var tender models.Tender
	if err := database.First(&tender, "id = ?", tenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "username is required"})
		return
	}

	var employee models.Employee
	if err := database.First(&employee, "username = ?", username).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "User not found"})
		return
	}

	if tender.CreatorUsername != username {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Unauthorized to update this tender"})
		return
	}

	var updateRequest schemas.TenderUpdateRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
		return
	}

	if updateRequest.Name != nil {
		tender.Name = *updateRequest.Name
	}
	if updateRequest.Description != nil {
		tender.Description = *updateRequest.Description
	}
	if updateRequest.ServiceType != nil {
		tender.ServiceType = *updateRequest.ServiceType
	}
	tender.Version++

	if err := database.Save(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to update tender"})
		return
	}

	c.JSON(http.StatusOK, schemas.TenderResponse{
		ID:          tender.ID,
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
		Status:      tender.Status,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt.Format(time.RFC3339),
	})
}

func GetTenderStatus(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}
	tenderID := c.Param("tenderId")
	username := c.Query("username")

	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Username must be provided"})
		return
	}

	var tender models.Tender
	if err := database.First(&tender, "id = ?", tenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	var employee models.Employee
	result := database.Where("username = ?", username).First(&employee)
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	if tender.CreatorUsername != username {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Unauthorized to view this tender"})
		return
	}

	c.JSON(http.StatusOK, tender.Status)
}

func UpdateTenderStatus(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}
	tenderID := c.Param("tenderId")

	var tender models.Tender
	if err := database.First(&tender, "id = ?", tenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	var employee models.Employee
	if err := database.First(&employee, "username = ?", username).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "User not found"})
		return
	}

	if tender.CreatorUsername != username {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Unauthorized to update this tender"})
		return
	}

	newStatus := c.Query("status")
	validStatuses := map[string]bool{"Created": true, "Published": true, "Closed": true}
	if !validStatuses[newStatus] {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid status"})
		return
	}

	tender.Status = newStatus
	if err := database.Save(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to update tender status"})
		return
	}

	response := schemas.TenderResponse{
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

func RollbackTender(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}

	var uriParams struct {
		TenderID      string `uri:"tenderId" binding:"required,max=100"`
		TenderVersion int    `uri:"version" binding:"required,min=1"`
	}
	if err := c.ShouldBindUri(&uriParams); err != nil {
		if len(uriParams.TenderID) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"reason": "TenderID cannot be longer than 100 characters"})
			return
		}
		if uriParams.TenderVersion < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"reason": "Version must be greater than 0"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
		return
	}

	var queryParams struct {
		Username string `form:"username" binding:"required"`
	}
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
		return
	}

	var user models.Employee
	if err := database.Where("username = ?", queryParams.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	var tender models.Tender
	if err := database.First(&tender, "id = ?", uriParams.TenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	if tender.CreatorUsername != queryParams.Username {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Unauthorized to rollback this tender"})
		return
	}

	var history models.TenderHistory
	if err := database.First(&history, "tender_id = ? AND version = ?", uriParams.TenderID, uriParams.TenderVersion).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender version not found"})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to save current tender history"})
		return
	}

	tender.Name = history.Name
	tender.Description = history.Description
	tender.ServiceType = history.ServiceType
	tender.Status = history.Status
	tender.Version++

	if err := database.Save(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to update tender"})
		return
	}

	tenderResponse := schemas.TenderResponse{
		ID:          tender.ID,
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
		Status:      tender.Status,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, tenderResponse)
}
