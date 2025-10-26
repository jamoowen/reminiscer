package handlers

import (
	"net/http"

	"github.com/jamoowen/reminiscer/internal/api"
	"github.com/jamoowen/reminiscer/internal/errors"
	"github.com/jamoowen/reminiscer/internal/middleware"
	"github.com/jamoowen/reminiscer/internal/models"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	store   models.Store
	authMid *middleware.AuthMiddleware
}

func NewAuthHandler(store models.Store, authMid *middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		store:   store,
		authMid: authMid,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Invalid request format")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Invalid request data")
	}

	// Check if user already exists
	existingUser, err := h.store.Users().GetByEmail(req.Email)
	if err != nil {
		if errors.IsCode(err, errors.CodeDatabaseError) {
			return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Internal server error")
		}
		return api.SendError(c, http.StatusInternalServerError, errors.CodeInternalError, "Error checking existing user")
	}
	if existingUser != nil {
		return api.SendError(c, http.StatusConflict, errors.CodeAlreadyExists, "Email already registered")
	}

	// Create new user
	user := &models.User{
		Email:         req.Email,
		Username:      req.Username,
		Password:      req.Password,
		Authenticated: true,
	}

	if err := h.store.Users().Create(user); err != nil {
		if errors.IsCode(err, errors.CodeDatabaseError) {
			return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to create user in database")
		}
		return api.SendError(c, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create user")
	}

	// Generate token
	token, err := h.authMid.GenerateToken(user)
	if err != nil {
		return api.SendError(c, http.StatusInternalServerError, errors.CodeInternalError, "Failed to generate token")
	}

	return api.SendSuccess(c, http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login handles user authentication
func (h *AuthHandler) Login(c echo.Context) error {
	var req AuthRequest
	if err := c.Bind(&req); err != nil {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Invalid request format")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Invalid request data")
	}

	// Authenticate user
	user, err := h.store.Users().Authenticate(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.IsCode(err, errors.CodeInvalidCredentials):
			return api.SendError(c, http.StatusUnauthorized, errors.CodeInvalidCredentials, "Invalid email or password")
		case errors.IsCode(err, errors.CodeDatabaseError):
			return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Database error occurred")
		default:
			return api.SendError(c, http.StatusInternalServerError, errors.CodeInternalError, "Authentication failed")
		}
	}

	// Generate token
	token, err := h.authMid.GenerateToken(user)
	if err != nil {
		return api.SendError(c, http.StatusInternalServerError, errors.CodeInternalError, "Failed to generate token")
	}

	return api.SendSuccess(c, http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

// Me returns the current authenticated user
func (h *AuthHandler) Me(c echo.Context) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return api.SendError(c, http.StatusUnauthorized, errors.CodeUnauthorized, "Not authenticated")
	}

	return api.SendSuccess(c, http.StatusOK, user)
}

// SetupRoutes sets up the authentication routes
func (h *AuthHandler) SetupRoutes(e *echo.Echo) {
	e.POST("/auth/register", h.Register)
	e.POST("/auth/login", h.Login)
	e.GET("/auth/me", h.Me, h.authMid.Authenticate)
}
