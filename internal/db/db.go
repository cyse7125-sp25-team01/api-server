package db

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB initializes the database using GORM
func ConnectDB() (*gorm.DB, error) {
	// Get individual environment variables with default values
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost" // Default value
	}

	portStr := os.Getenv("DB_PORT")
	if portStr == "" {
		portStr = "5432" // Default value
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres" // Default value
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres" // Default value
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "api" // Default value
	}

	// Convert DB_PORT to integer and handle potential errors
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid port value: %v", err)
		return nil, err
	}

	// Construct the DSN using individual environment variables
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		host, port, user, password, dbName)

	// Open connection using GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}
	db.Exec("SET search_path To api")
	fmt.Println("✅ Database connected successfully!")

	return db, nil
}
