package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB initializes the database using GORM
func ConnectDB() (*gorm.DB, error) {
	// Get database connection string from environment variables
	dsn := os.Getenv("DB_DSN") // Example: "host=localhost user=admin password=adminpassword dbname=social port=5432 sslmode=disable TimeZone=UTC"

	// Open connection using GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}

	fmt.Println("✅ Database connected successfully!")

	return db, nil
}
