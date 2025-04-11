package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/csye7125/team01/internal/db"
	"github.com/csye7125/team01/internal/store"
)

func main() {
	fmt.Println("🚀 Starting API Server...")

	// Initialize OpenTelemetry
	shutdown, err := InitTracer()
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			log.Printf("Error shutting down OpenTelemetry: %v", err)
		}
	}()

	// ✅ Connect to DB using GORM
	database, err := db.ConnectDB()
	if err != nil {
		log.Fatal("❌ Could not connect to the database")
	}

	// ✅ Run automatic migrations
	database.AutoMigrate(&store.User{})

	fmt.Println("✅ Database migrations completed!")

	// ✅ Fix: Use `NewStorage(database)` correctly
	storage := store.NewStorage(database)
	app := NewApplication(storage) // ✅ Fix: app.store is now correctly initialized

	mux := app.mount()
	log.Fatal(app.run(mux))
}