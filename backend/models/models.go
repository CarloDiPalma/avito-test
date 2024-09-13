package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Employee struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Username  string    `gorm:"unique;not null"`
	FirstName string
	LastName  string
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (Employee) TableName() string {
	return "employee"
}

type OrganizationType string

const (
	IE  OrganizationType = "IE"
	LLC OrganizationType = "LLC"
	JSC OrganizationType = "JSC"
)

type Organization struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Description string    `gorm:"type:text"`
	Type        string    `gorm:"type:organization_type"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (Organization) TableName() string {
	return "organization"
}

type OrganizationResponsible struct {
	ID             uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;foreignKey:OrganizationID"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;foreignKey:EmployeeID"`
}

func (OrganizationResponsible) TableName() string {
	return "organization_responsible"
}

var validServiceTypes = []string{"Construction", "Delivery", "Manufacture"}

type Tender struct {
	ID              uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	Name            string    `json:"name" binding:"required,max=100"`
	Description     string    `json:"description" binding:"required,max=500"`
	ServiceType     string    `json:"serviceType" gorm:"column:service_type" validate:"required,oneof=Construction Delivery Manufacture"`
	Status          string    `json:"status" binding:"omitempty,oneof=Created Published Closed" gorm:"default:'Created'"`
	OrganizationID  uuid.UUID `json:"organizationId" binding:"required,max=100"`
	CreatorUsername string    `json:"creatorUsername" binding:"required"`
	Version         int       `json:"version" gorm:"default:1"`
	CreatedAt       time.Time `json:"createdAt"`
}

func ValidateServiceType(serviceType string) bool {
	for _, v := range validServiceTypes {
		if serviceType == v {
			return true
		}
	}
	return false
}

func (t *Tender) Validate() error {
	validate := validator.New()
	err := validate.RegisterValidation("oneof", func(fl validator.FieldLevel) bool {
		return ValidateServiceType(fl.Field().String())
	})

	if err != nil {
		return err
	}

	return validate.Struct(t)
}

type TenderHistory struct {
	ID              uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	TenderID        uuid.UUID `gorm:"type:uuid" json:"tender_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ServiceType     string    `json:"service_type"`
	Status          string    `json:"status"`
	Version         int       `json:"version"`
	CreatorUsername string    `json:"creator_username"`
	OrganizationID  uuid.UUID `gorm:"type:uuid" json:"organization_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Bid struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	Name        string    `json:"name" binding:"required,max=100"`
	Description string    `json:"description" binding:"required,max=500"`
	Status      string    `json:"status" binding:"required,oneof=Created Published Canceled"`
	TenderID    uuid.UUID `gorm:"type:uuid;not null" json:"tender_id"`
	AuthorType  string    `json:"authorType" binding:"required,oneof=Organisation, User"`
	AuthorID    uuid.UUID `json:"authorId" binding:"required"`
	Version     int       `gorm:"default:1" json:"version" binding:"required,min=1"`
	CreatedAt   time.Time `json:"createdAt" binding:"required"`
	Decision    *string   `gorm:"type:bid_decision;default:NULL"`
}

type BidHistory struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	BidID       uuid.UUID `gorm:"type:uuid;not null" json:"bidId"`
	Name        string    `json:"name" binding:"required,max=100"`
	Description string    `json:"description" binding:"required,max=500"`
	Status      string    `json:"status" binding:"required,oneof=Created Published Canceled"`
	TenderID    uuid.UUID `gorm:"type:uuid;not null" json:"tender_id"`
	AuthorType  string    `json:"authorType" binding:"required,oneof=Organisation, User"`
	AuthorID    uuid.UUID `json:"authorId" binding:"required"`
	Version     int       `gorm:"default:1" json:"version" binding:"required,min=1"`
	CreatedAt   time.Time `json:"createdAt"`
	Decision    *string   `gorm:"type:bid_decision;default:NULL"`
}

type BidFeedback struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	BidID     uuid.UUID `gorm:"type:uuid;not null" json:"bid_id"`
	Feedback  string    `json:"feedback" binding:"required,max=1000"`
	AuthorID  uuid.UUID `json:"author_id" binding:"required"`
	CreatedAt time.Time `json:"created_at" binding:"required"`
}
