package schemas

import (
	"github.com/google/uuid"
)

type TenderResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ServiceType string    `json:"serviceType"`
	Status      string    `json:"status"`
	Version     int       `json:"version"`
	CreatedAt   string    `json:"createdAt"`
}

type TenderUpdateRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	ServiceType *string `json:"serviceType" binding:"omitempty,oneof=Construction Delivery Manufacture"`
}

type RollbackTenderRequest struct {
	TenderID      string `uri:"tenderId" binding:"required,max=100"`
	TenderVersion int    `uri:"version" binding:"required,min=1"`
	Username      string `form:"username" binding:"required"`
}

type CreateBidRequest struct {
	Name            string `json:"name" binding:"required,max=100"`
	Description     string `json:"description" binding:"required,max=500"`
	Status          string `json:"status" binding:"required,oneof=Created Published Canceled Approved Rejected"`
	TenderID        string `json:"tenderId" binding:"required"`
	OrganizationID  string `json:"organizationId" binding:"required"`
	CreatorUsername string `json:"creatorUsername" binding:"required"`
}

type BidCreateRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"required,max=500"`
	TenderID    string `json:"tenderId" binding:"required,uuid"`
	AuthorType  string `json:"authorType" binding:"required,oneof=User Organization"`
	AuthorID    string `json:"authorId" binding:"required,uuid"`
}

type BidCreateResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	AuthorType string    `json:"authorType"`
	AuthorID   uuid.UUID `json:"authorId"`
	Version    int       `json:"version"`
	CreatedAt  string    `json:"createdAt"`
}
