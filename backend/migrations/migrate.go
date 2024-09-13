package migrations

import (
	"ZADANIE-6105/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func RunMigrations(postgresConn string) {
	db, err := gorm.Open(postgres.Open(postgresConn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Удаление таблиц перед миграцией
	err = dropTables(db)
	if err != nil {
		log.Fatalf("Error dropping tables: %v", err)
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

	// Создание расширения uuid-ossp, если оно не существует
	err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error
	if err != nil {
		log.Fatalf("Error creating extension uuid-ossp: %v", err)
	}

	// Создание типа данных bid_decision, если он не существует
	err = db.Exec(`DO $$ BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'bid_decision') THEN
            CREATE TYPE bid_decision AS ENUM ('Approved', 'Rejected');
        END IF;
    END $$`).Error
	if err != nil {
		log.Fatalf("Error creating type bid_decision: %v", err)
	}

	err = db.AutoMigrate(
		// &models.Employee{},
		// &models.Organization{},
		// &models.OrganizationResponsible{},
		&models.Tender{},
		&models.TenderHistory{},
		&models.Bid{},
		&models.BidHistory{},
		&models.BidFeedback{},
	)
	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}

	log.Println("Database migrated successfully")
}

func dropTables(db *gorm.DB) error {
	tables := []string{
		"tenders",
		"tender_histories",
		"bids",
		"bid_histories",
		"bid_feedbacks",
	}

	for _, table := range tables {
		err := db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE;").Error
		if err != nil {
			return err
		}
	}

	log.Println("Tables dropped successfully")
	return nil
}
