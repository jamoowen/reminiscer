package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator/v10"
	"github.com/jamoowen/reminiscer/internal/config"
	"github.com/jamoowen/reminiscer/internal/database"
	"github.com/jamoowen/reminiscer/internal/handlers"
	"github.com/jamoowen/reminiscer/internal/middleware"
	"github.com/jamoowen/reminiscer/internal/models"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
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
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	log.Printf("Initializing database at %s\n", cfg.Database.Path)
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	log.Print("Running migrations")
	if err := db.Migrate("./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize store
	fmt.Print("Initializing store")
	store := models.NewSQLiteStore(db.DB)

	// Initialize Echo
	fmt.Print("Initializing Echo")
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())

	// Configure CORS
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: []string{cfg.Security.AllowedOrigins},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE, echo.PATCH},
	}))

	// Rate limiting
	if !cfg.IsDevelopment() {
		e.Use(echoMiddleware.RateLimiter(echoMiddleware.NewRateLimiterMemoryStore(
			rate.Limit(cfg.Security.RateLimitRequests),
		)))
	}

	// Initialize auth middleware
	fmt.Print("Initializing auth middleware")
	authMid := middleware.NewAuthMiddleware(cfg, store)

	// Initialize handlers
	fmt.Print("Initializing handlers")
	authHandler := handlers.NewAuthHandler(store, authMid)
	quoteHandler := handlers.NewQuoteHandler(store, authMid)
	groupHandler := handlers.NewGroupHandler(store, authMid)

	// Set up routes
	fmt.Print("Setting up routes")
	authHandler.SetupRoutes(e)
	quoteHandler.SetupRoutes(e)
	groupHandler.SetupRoutes(e)

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		if err := e.Shutdown(nil); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
	}()

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	if cfg.IsDevelopment() {
		log.Printf("Server starting in development mode on http://localhost%s", serverAddr)
	} else {
		log.Printf("Server starting in production mode on port %s", cfg.Server.Port)
	}

	if err := e.Start(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
