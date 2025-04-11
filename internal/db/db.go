package db

import (
	"context" // Added this import
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes" // Added this import
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	// Create a custom logger that creates spans for database operations
	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // Slow SQL threshold
			LogLevel:                  logger.Info,            // Log level
			IgnoreRecordNotFoundError: false,                  // Log not found error
			Colorful:                  true,                   // Disable color
		},
	)

	// Open connection using GORM with our custom callbacks
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}

	// Add custom callbacks for OpenTelemetry tracing
	_ = db.Callback().Create().Before("gorm:create").Register("before_create", beforeCallback)
	_ = db.Callback().Create().After("gorm:create").Register("after_create", afterCallback)
	_ = db.Callback().Query().Before("gorm:query").Register("before_query", beforeCallback)
	_ = db.Callback().Query().After("gorm:query").Register("after_query", afterCallback)
	_ = db.Callback().Update().Before("gorm:update").Register("before_update", beforeCallback)
	_ = db.Callback().Update().After("gorm:update").Register("after_update", afterCallback)
	_ = db.Callback().Delete().Before("gorm:delete").Register("before_delete", beforeCallback)
	_ = db.Callback().Delete().After("gorm:delete").Register("after_delete", afterCallback)

	db.Exec("SET search_path To api")
	fmt.Println("✅ Database connected successfully!")

	return db, nil
}

// beforeCallback creates a span before database operations
func beforeCallback(db *gorm.DB) {
	ctx := db.Statement.Context
	tracer := otel.Tracer("gorm")
	
	// Extract operation name from the gorm scope
	operation := "database"
	if db.Statement.Schema != nil {
		operation = db.Statement.Schema.Table
	}
	
	// Create a new span for this database operation
	spanName := fmt.Sprintf("%s.%s", "gorm", operation)
	newCtx, span := tracer.Start(ctx, spanName, 
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", operation),
		),
	)
	
	// Store the span in the db context for later retrieval
	db.Statement.Context = context.WithValue(newCtx, "span", span)
}

// afterCallback ends the span after database operations
func afterCallback(db *gorm.DB) {
	// Get the span from context
	ctx := db.Statement.Context
	spanValue := ctx.Value("span")
	if spanValue == nil {
		return
	}
	
	span, ok := spanValue.(trace.Span)
	if !ok {
		return
	}
	
	// Add SQL to span attributes
	if db.Statement.SQL.String() != "" {
		span.SetAttributes(
			attribute.String("db.statement", db.Statement.SQL.String()),
			attribute.Int64("db.rows_affected", db.Statement.RowsAffected),
		)
	}
	
	// Record error if present
	if db.Statement.Error != nil {
		span.RecordError(db.Statement.Error)
		span.SetStatus(codes.Error, db.Statement.Error.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	
	// End the span
	span.End()
}