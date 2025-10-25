package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/jamoowen/reminiscer/internal/config"
	"github.com/jamoowen/reminiscer/internal/database"
	"github.com/jamoowen/reminiscer/internal/handlers"
	"github.com/jamoowen/reminiscer/internal/middleware"
	"github.com/jamoowen/reminiscer/internal/models"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

func main() {
	// Initialize configuration
	cfg := config.GetConfig()

	// Set up database
	dbPath := filepath.Join(".", "data", "reminiscer.db")
	err := os.MkdirAll(filepath.Dir(dbPath), 0755)
	if err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate("./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize store
	store := models.NewSQLiteStore(db.DB)

	// Initialize Echo
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	// Initialize auth middleware
	authMid := middleware.NewAuthMiddleware(cfg, store)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(store, authMid)
	quoteHandler := handlers.NewQuoteHandler(store, authMid)
	groupHandler := handlers.NewGroupHandler(store, authMid)

	// Set up routes
	authHandler.SetupRoutes(e)
	quoteHandler.SetupRoutes(e)
	groupHandler.SetupRoutes(e)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
