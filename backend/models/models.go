package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Employee модель для таблицы employee
type Employee struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"unique;not null"`
	FirstName string
	LastName  string
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// OrganizationType определяет тип организации
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

type OrganizationResponsible struct {
	ID             uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrganizationID uuid.UUID    `gorm:"type:uuid;not null" gorm:"foreignKey:OrganizationID"`
	EmployeeID     uuid.UUID    `gorm:"type:uuid;not null" gorm:"foreignKey:EmployeeID"`
	Organization   Organization `gorm:"constraint:OnDelete:CASCADE"`
	Employee       Employee     `gorm:"constraint:OnDelete:CASCADE"`
}

var validServiceTypes = []string{"Construction", "Delivery", "Manufacture"}

type Tender struct {
	ID              uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	Name            string    `json:"name" binding:"required"`
	Description     string    `json:"description" binding:"required"`
	ServiceType     string    `json:"serviceType" gorm:"column:service_type" validate:"required,oneof=Construction Delivery Manufacture"`
	Status          string    `json:"status" binding:"required,oneof=Created Pending Completed"`
	OrganizationID  uuid.UUID `json:"organizationId" binding:"required"`
	CreatorUsername string    `json:"creatorUsername" binding:"required"`
	Version         int       `json:"version" gorm:"default:1"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// ValidateServiceType проверяет, что значение ServiceType является одним из допустимых значений
func ValidateServiceType(serviceType string) bool {
	for _, v := range validServiceTypes {
		if serviceType == v {
			return true
		}
	}
	return false
}

// Валидация модели
func (t *Tender) Validate() error {
	validate := validator.New()
	validate.RegisterValidation("oneof", func(fl validator.FieldLevel) bool {
		return ValidateServiceType(fl.Field().String())
	})
	return validate.Struct(t)
}

type TenderHistory struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	TenderID    uuid.UUID `json:"tenderId"` // Ссылка на основной тендер
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ServiceType string    `json:"serviceType"`
	Status      string    `json:"status"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Proposal модель для таблицы proposal
type Proposal struct {
	ID         uint      `gorm:"primaryKey"`
	TenderID   uint      `gorm:"not null"`
	EmployeeID uint      `gorm:"not null"`
	Offer      string    `gorm:"not null"`
	Status     string    `gorm:"not null"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	Tender     Tender    `gorm:"foreignKey:TenderID"`
	Employee   Employee  `gorm:"foreignKey:EmployeeID"`
}
