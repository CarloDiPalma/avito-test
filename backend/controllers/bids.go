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
		Version:     1,
		Status:      "CREATED",
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
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Tender ID is required"})
		return
	}

	// Проверка существования тендера
	var tender models.Tender
	if err := database.First(&tender, "id = ?", tenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to retrieve bids"})
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
	bidID := c.GetString("bidId")

	// Проверка наличия bidID
	if bidID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Bid ID is required"})
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

	// Поиск предложения по ID
	var bid models.Bid
	if err := database.First(&bid, "id = ?", bidID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	// Проверка прав доступа: пользователь должен быть создателем предложения
	if bid.AuthorID != employee.ID {
		// Проверка, является ли пользователь ответственным за организацию, к которой принадлежит предложение
		var orgResponsible models.OrganizationResponsible
		if err := database.Where("organization_id = ? AND employee_id = ?", bid.TenderID, employee.ID).First(&orgResponsible).Error; err != nil {
			// Если пользователь не имеет прав ни как создатель, ни как ответственный за организацию, возвращаем 403
			c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized to access this bid"})
			return
		}
	}

	// Если пользователь имеет доступ, возвращаем статус предложения
	c.JSON(http.StatusOK, gin.H{
		"status": bid.Status,
	})
}

func UpdateBidStatus(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)

	// Получение идентификатора предложения из параметров пути
	bidID := c.Param("bidId")

	// Проверка наличия bidID
	if bidID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Bid ID is required"})
		return
	}

	// Получение обязательных параметров запроса: status и username
	status := c.Query("status")
	username := c.Query("username")

	// Проверка обязательных параметров
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Status is required"})
		return
	}

	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	// Проверка допустимых значений статуса
	validStatuses := map[string]bool{
		"Created":   true,
		"Published": true,
		"Canceled":  true,
	}

	if !validStatuses[status] {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid status value"})
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
		c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	// Проверка прав доступа: пользователь должен быть создателем предложения
	if bid.AuthorID != employee.ID {
		// Проверка, является ли пользователь ответственным за организацию, к которой принадлежит предложение
		var orgResponsible models.OrganizationResponsible
		if err := database.Where("organization_id = ? AND employee_id = ?", bid.TenderID, employee.ID).First(&orgResponsible).Error; err != nil {
			// Если пользователь не имеет прав ни как создатель, ни как ответственный за организацию, возвращаем 403
			c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized to update this bid"})
			return
		}
	}

	// Обновление статуса предложения
	bid.Status = status

	// Сохранение изменений в базе данных
	if err := database.Save(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to update bid status"})
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

	// Возврат успешного ответа
	c.JSON(http.StatusOK, response)
}

func EditBid(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)

	// Получение идентификатора предложения и данных для обновления
	bidID := c.Param("bidId")
	var bidInput schemas.BidEditRequest
	if err := c.ShouldBindJSON(&bidInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid input"})
		return
	}

	// Проверка наличия предложения
	var bid models.Bid
	if err := database.First(&bid, "id = ?", bidID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	// Проверка прав доступа
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

	// Сохранение текущей версии в BidHistory
	history := models.BidHistory{
		ID:          uuid.New(), // или другой метод генерации UUID
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

	// Обновление предложения
	bid.Name = bidInput.Name
	bid.Description = bidInput.Description
	bid.Version += 1
	if err := database.Save(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to update bid"})
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

func RollbackBid(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)

	// Получение идентификатора предложения, версии и username
	bidID := c.Param("bidId")
	versionStr := c.Param("version")
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	// Проверка наличия предложения
	var bid models.Bid
	if err := database.First(&bid, "id = ?", bidID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	// Проверка прав доступа
	var employee models.Employee
	if err := database.First(&employee, "username = ?", username).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	if bid.AuthorID != employee.ID {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized to rollback this bid"})
		return
	}

	// Поиск записи в BidHistory по версии
	var history models.BidHistory
	if err := database.First(&history, "bid_id = ? AND version = ?", bidID, versionStr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Version not found in history"})
		return
	}

	// Инкрементирование версии и запись текущего состояния в BidHistory
	currentHistory := models.BidHistory{
		ID:          uuid.New(), // Новый UUID для истории
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

	// Обновление предложения на основе данных из BidHistory
	bid.Name = history.Name
	bid.Description = history.Description
	bid.Status = history.Status
	bid.TenderID = history.TenderID
	bid.AuthorType = history.AuthorType
	bid.AuthorID = history.AuthorID
	bid.Version += 1           // Инкремент версии
	bid.CreatedAt = time.Now() // Обновляем CreatedAt на текущее время
	bid.Decision = history.Decision

	if err := database.Save(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to rollback bid"})
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

func SubmitDecision(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)

	// Получение идентификатора предложения, решения и username из query параметров
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

	// Проверка существования предложения
	var bid models.Bid
	if err := database.First(&bid, "id = ?", bidID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	// Проверка существования пользователя
	var employee models.Employee
	if err := database.First(&employee, "username = ?", username).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	// Проверка прав доступа: пользователь должен быть автором предложения
	if bid.AuthorID != employee.ID {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized to submit decision for this bid"})
		return
	}

	// Обновление предложения с учетом принятого решения
	bid.Decision = &decision
	if err := database.Save(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Failed to submit decision"})
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

func SendFeedback(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)

	// Получение идентификатора предложения, отзыва и username из query параметров
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

	// Проверка длины отзыва
	if len(feedback) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Feedback exceeds maximum length of 1000 characters"})
		return
	}

	// Проверка существования предложения
	var bid models.Bid
	if err := database.First(&bid, "id = ?", bidID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	// Проверка существования тендера
	var tender models.Tender
	if err := database.First(&tender, "id = ?", bid.TenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	// Проверка существования пользователя
	var employee models.Employee
	if err := database.First(&employee, "username = ?", username).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent user"})
		return
	}

	// Проверка прав доступа: пользователь должен быть ответственным за организацию, связанную с тендером
	var orgResp models.OrganizationResponsible
	if err := database.First(&orgResp, "organization_id = ? AND employee_id = ?", tender.OrganizationID, employee.ID).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "User is not authorized to submit feedback for this bid"})
		return
	}

	// Создание отзыва
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

func GetBidReviews(c *gin.Context) {
	db, _ := c.Get("db")
	database := db.(*gorm.DB)

	// Получение параметров из запроса
	tenderID := c.Param("tenderId")
	authorUsername := c.Query("authorUsername")
	requesterUsername := c.Query("requesterUsername")
	limitStr := c.DefaultQuery("limit", "5")
	offsetStr := c.DefaultQuery("offset", "0")

	// Проверка наличия обязательных параметров
	if authorUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "authorUsername is required"})
		return
	}
	if requesterUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "requesterUsername is required"})
		return
	}

	// Преобразование limit и offset в int
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

	// Проверка существования пользователя-запрашивающего
	var requesterEmployee models.Employee
	if err := database.First(&requesterEmployee, "username = ?", requesterUsername).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Invalid or non-existent requester user"})
		return
	}

	// Проверка существования тендера
	var tender models.Tender
	if err := database.First(&tender, "id = ?", tenderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return
	}

	// Проверка, является ли requester ответственным за организацию
	var organizationResponsible models.OrganizationResponsible
	if err := database.First(&organizationResponsible, "organization_id = ? AND employee_id = ?", tender.OrganizationID, requesterEmployee.ID).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Requester does not have permission to view reviews for this tender"})
		return
	}

	// Проверка существования автора предложений
	var authorEmployee models.Employee
	if err := database.First(&authorEmployee, "username = ?", authorUsername).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Invalid or non-existent author user"})
		return
	}
	// Получение отзывов из таблицы bid_feedbacks для данного тендера и автора
	var reviews []schemas.BidReviewResponse
	if err := database.Table("bid_feedbacks").Select("id, feedback as description, created_at").
		Where("bid_id IN (SELECT id FROM bids WHERE tender_id = ? AND author_id = ?)", tenderID, authorEmployee.ID).
		Limit(limit).Offset(offset).Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "Error retrieving reviews"})
		return
	}

	// Проверка наличия отзывов
	if len(reviews) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"reason": "No reviews found"})
		return
	}

	// Формирование ответа
	c.JSON(http.StatusOK, reviews)
}
