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

func CreateBid(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}

	var bidInput schemas.BidCreateRequest

	if err := c.ShouldBindJSON(&bidInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	var employee models.Employee
	if err := database.Where("id = ?", bidInput.AuthorID).First(&employee).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Unauthorized, user does not exist"})
		return
	}

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

	bid := models.Bid{
		Name:        bidInput.Name,
		Description: bidInput.Description,
		TenderID:    tenderID,
		AuthorType:  bidInput.AuthorType,
		AuthorID:    employee.ID,
		Version:     1,
		Status:      "Created",
		CreatedAt:   time.Now(),
	}

	if err := database.Create(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to create bid"})
		return
	}

	response := schemas.BidCreateResponse{
		ID:         bid.ID,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorID:   bid.AuthorID,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
	}

	c.JSON(http.StatusOK, response)
}

func GetMyBids(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}

	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	var employee models.Employee
	if err := database.Where("username = ?", username).First(&employee).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid username"})
		return
	}

	limit := 5
	offset := 0

	if l, err := strconv.Atoi(c.DefaultQuery("limit", "5")); err == nil {
		limit = l
	}
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil {
		offset = o
	}

	var bids []models.Bid
	if err := database.Where("author_id = ?", employee.ID).Limit(limit).Offset(offset).Find(&bids).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve bids"})
		return
	}

	var responseBids []schemas.BidCreateResponse
	for _, bid := range bids {
		responseBids = append(responseBids, schemas.BidCreateResponse{
			ID:         bid.ID,
			Name:       bid.Name,
			Status:     bid.Status,
			AuthorID:   bid.AuthorID,
			AuthorType: bid.AuthorType,
			Version:    bid.Version,
			CreatedAt:  bid.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
		})
	}

	c.JSON(http.StatusOK, responseBids)
}

func GetBidsByTender(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}

	tenderID := c.Param("tenderId")
	if tenderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Tender ID is required"})
		return
	}

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
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	if tender.CreatorUsername != username {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized to access bids for this tender"})
		return
	}

	limit := 5
	offset := 0

	if l, err := strconv.Atoi(c.DefaultQuery("limit", "5")); err == nil {
		limit = l
	}
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil {
		offset = o
	}

	var bids []models.Bid
	if err := database.Where("tender_id = ?", tenderID).Limit(limit).Offset(offset).Find(&bids).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to retrieve bids"})
		return
	}

	var responseBids []schemas.BidCreateResponse
	for _, bid := range bids {
		responseBids = append(responseBids, schemas.BidCreateResponse{
			ID:         bid.ID,
			Name:       bid.Name,
			Status:     bid.Status,
			AuthorType: bid.AuthorType,
			AuthorID:   bid.AuthorID,
			Version:    bid.Version,
			CreatedAt:  bid.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
		})
	}

	c.JSON(http.StatusOK, responseBids)
}

func GetBidStatus(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}

	bidID := c.GetString("bidId")

	if bidID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Bid ID is required"})
		return
	}

	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	var employee models.Employee
	if err := database.First(&employee, "username = ?", username).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	var bid models.Bid
	if err := database.First(&bid, "id = ?", bidID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	if bid.AuthorID != employee.ID {
		var orgResponsible models.OrganizationResponsible
		if err := database.Where("organization_id = ? AND employee_id = ?", bid.TenderID, employee.ID).First(&orgResponsible).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized to access this bid"})
			return
		}
	}

	c.JSON(http.StatusOK, bid.Status)
}

func UpdateBidStatus(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}

	bidID := c.Param("bidId")

	if bidID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Bid ID is required"})
		return
	}

	status := c.Query("status")
	username := c.Query("username")

	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Status is required"})
		return
	}

	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	validStatuses := map[string]bool{
		"Created":   true,
		"Published": true,
		"Canceled":  true,
	}

	if !validStatuses[status] {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid status value"})
		return
	}

	var employee models.Employee
	if err := database.First(&employee, "username = ?", username).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	var bid models.Bid
	if err := database.First(&bid, "id = ?", bidID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	if bid.AuthorID != employee.ID {
		var orgResponsible models.OrganizationResponsible
		if err := database.Where("organization_id = ? AND employee_id = ?", bid.TenderID, employee.ID).First(&orgResponsible).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized to update this bid"})
			return
		}
	}

	bid.Status = status

	if err := database.Save(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to update bid status"})
		return
	}

	response := schemas.BidCreateResponse{
		ID:         bid.ID,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorID:   bid.AuthorID,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
	}

	c.JSON(http.StatusOK, response)
}

func EditBid(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}

	bidID := c.Param("bidId")
	var bidInput schemas.BidEditRequest
	if err := c.ShouldBindJSON(&bidInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid input"})
		return
	}

	var bid models.Bid
	if err := database.First(&bid, "id = ?", bidID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	var employee models.Employee
	if err := database.First(&employee, "username = ?", username).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	if bid.AuthorID != employee.ID {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized to edit this bid"})
		return
	}

	history := models.BidHistory{
		ID:          uuid.New(),
		BidID:       bid.ID,
		Name:        bid.Name,
		Description: bid.Description,
		Status:      bid.Status,
		TenderID:    bid.TenderID,
		AuthorType:  bid.AuthorType,
		AuthorID:    bid.AuthorID,
		Version:     bid.Version,
		CreatedAt:   bid.CreatedAt,
		Decision:    bid.Decision,
	}
	database.Create(&history)

	bid.Name = bidInput.Name
	bid.Description = bidInput.Description
	bid.Version += 1
	if err := database.Save(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to update bid"})
		return
	}

	response := schemas.BidCreateResponse{
		ID:         bid.ID,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorID:   bid.AuthorID,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
	}
	c.JSON(http.StatusOK, response)
}

func RollbackBid(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}

	bidID := c.Param("bidId")
	versionStr := c.Param("version")
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	var bid models.Bid
	if err := database.First(&bid, "id = ?", bidID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	var employee models.Employee
	if err := database.First(&employee, "username = ?", username).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	if bid.AuthorID != employee.ID {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized to rollback this bid"})
		return
	}

	var history models.BidHistory
	if err := database.First(&history, "bid_id = ? AND version = ?", bidID, versionStr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Version not found in history"})
		return
	}

	currentHistory := models.BidHistory{
		ID:          uuid.New(),
		BidID:       bid.ID,
		Name:        bid.Name,
		Description: bid.Description,
		Status:      bid.Status,
		TenderID:    bid.TenderID,
		AuthorType:  bid.AuthorType,
		AuthorID:    bid.AuthorID,
		Version:     bid.Version,
		CreatedAt:   bid.CreatedAt,
		Decision:    bid.Decision,
	}
	if err := database.Create(&currentHistory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to save bid history"})
		return
	}

	bid.Name = history.Name
	bid.Description = history.Description
	bid.Status = history.Status
	bid.TenderID = history.TenderID
	bid.AuthorType = history.AuthorType
	bid.AuthorID = history.AuthorID
	bid.Version += 1
	bid.CreatedAt = time.Now()
	bid.Decision = history.Decision

	if err := database.Save(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to rollback bid"})
		return
	}

	response := schemas.BidCreateResponse{
		ID:         bid.ID,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorID:   bid.AuthorID,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
	}
	c.JSON(http.StatusOK, response)
}

func SubmitDecision(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}

	bidID := c.Param("bidId")
	decision := c.Query("decision")
	username := c.Query("username")

	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	if decision != "Approved" && decision != "Rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid decision value"})
		return
	}

	var bid models.Bid
	if err := database.First(&bid, "id = ?", bidID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	var employee models.Employee
	if err := database.First(&employee, "username = ?", username).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	if bid.AuthorID != employee.ID {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized to submit decision for this bid"})
		return
	}

	bid.Decision = &decision
	if err := database.Save(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to submit decision"})
		return
	}

	response := schemas.BidCreateResponse{
		ID:         bid.ID,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorID:   bid.AuthorID,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
	}
	c.JSON(http.StatusOK, response)
}

func SendFeedback(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}

	bidID := c.Param("bidId")
	feedback := c.Query("bidFeedback")
	username := c.Query("username")

	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	if feedback == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Feedback is required"})
		return
	}

	if len(feedback) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Feedback exceeds maximum length of 1000 characters"})
		return
	}

	var bid models.Bid
	if err := database.First(&bid, "id = ?", bidID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	var tender models.Tender
	if err := database.First(&tender, "id = ?", bid.TenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	var employee models.Employee
	if err := database.First(&employee, "username = ?", username).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	var orgResp models.OrganizationResponsible
	if err := database.First(&orgResp, "organization_id = ? AND employee_id = ?", tender.OrganizationID, employee.ID).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized to submit feedback for this bid"})
		return
	}

	feedbackRecord := models.BidFeedback{
		BidID:     bid.ID,
		Feedback:  feedback,
		AuthorID:  employee.ID,
		CreatedAt: time.Now(),
	}
	if err := database.Create(&feedbackRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to submit feedback"})
		return
	}

	response := schemas.BidCreateResponse{
		ID:         bid.ID,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorID:   bid.AuthorID,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
	}
	c.JSON(http.StatusOK, response)
}

func GetBidReviews(c *gin.Context) {
	database, ok := utils.GetDB(c)
	if !ok {
		return
	}

	tenderID := c.Param("tenderId")
	authorUsername := c.Query("authorUsername")
	requesterUsername := c.Query("requesterUsername")
	limitStr := c.DefaultQuery("limit", "5")
	offsetStr := c.DefaultQuery("offset", "0")

	if authorUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "authorUsername is required"})
		return
	}
	if requesterUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "requesterUsername is required"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid limit value"})
		return
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid offset value"})
		return
	}

	var requesterEmployee models.Employee
	if err := database.First(&requesterEmployee, "username = ?", requesterUsername).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent requester user"})
		return
	}

	var tender models.Tender
	if err := database.First(&tender, "id = ?", tenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	var organizationResponsible models.OrganizationResponsible
	if err := database.First(&organizationResponsible, "organization_id = ? AND employee_id = ?", tender.OrganizationID, requesterEmployee.ID).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Requester does not have permission to view reviews for this tender"})
		return
	}

	var authorEmployee models.Employee
	if err := database.First(&authorEmployee, "username = ?", authorUsername).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Invalid or non-existent author user"})
		return
	}

	var reviews []schemas.BidReviewResponse
	if err := database.Table("bid_feedbacks").Select("id, feedback as description, created_at").
		Where("bid_id IN (SELECT id FROM bids WHERE tender_id = ? AND author_id = ?)", tenderID, authorEmployee.ID).
		Limit(limit).Offset(offset).Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Error retrieving reviews"})
		return
	}

	if len(reviews) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"reason": "No reviews found"})
		return
	}

	c.JSON(http.StatusOK, reviews)
}
