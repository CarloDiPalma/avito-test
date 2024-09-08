package migrations

import (
	"ZADANIE-6105/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func RunMigrations() {
	postgresConn := os.Getenv("POSTGRES_CONN")
	if postgresConn == "" {
		postgresConn = "postgres://postgres:7744@localhost:5432/avito"
	}
	db, err := gorm.Open(postgres.Open(postgresConn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Создание типа данных organization_type, если он не существует
	err = db.Exec(`DO $$ BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'organization_type') THEN
            CREATE TYPE organization_type AS ENUM ('IE', 'LLC', 'JSC');
        END IF;
    END $$`).Error
	if err != nil {
		log.Fatalf("Error creating type organization_type: %v", err)
	}

	// Автоматическая миграция всех моделей
	err = db.AutoMigrate(
		&models.Employee{},
		&models.Organization{},
		&models.OrganizationResponsible{},
		&models.Tender{},
		&models.Proposal{},
	)
	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}

	log.Println("Database migrated successfully")
}
