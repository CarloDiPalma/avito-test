package schemas

import (
	"time"

	"github.com/google/uuid"
)

type CreateTenderResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ServiceType string    `json:"serviceType"`
	Status      string    `json:"status"`
	Version     int       `json:"version"`
	CreatedAt   string    `json:"createdAt"`
}

type CreateBidRequest struct {
	Name            string `json:"name" binding:"required,max=100"`
	Description     string `json:"description" binding:"required,max=500"`
	Status          string `json:"status" binding:"required,oneof=Created Published Canceled Approved Rejected"`
	TenderID        string `json:"tenderId" binding:"required"`
	OrganizationID  string `json:"organizationId" binding:"required"`
	CreatorUsername string `json:"creatorUsername" binding:"required"`
}

type CreateBidResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	AuthorType string    `json:"authorType"`
	AuthorID   uuid.UUID `json:"authorId"`
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"createdAt"`
}
