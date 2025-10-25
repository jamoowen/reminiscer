package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jamoowen/reminiscer/internal/api"
	"github.com/jamoowen/reminiscer/internal/config"
	"github.com/jamoowen/reminiscer/internal/errors"
	"github.com/jamoowen/reminiscer/internal/models"
	"github.com/labstack/echo/v4"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type AuthMiddleware struct {
	config *config.Config
	store  models.Store
}

func NewAuthMiddleware(cfg *config.Config, store models.Store) *AuthMiddleware {
	return &AuthMiddleware{
		config: cfg,
		store:  store,
	}
}

// GenerateToken creates a new JWT token for a user
func (m *AuthMiddleware) GenerateToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.config.JWT.TokenExpiration())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(m.config.JWT.Secret))
	if err != nil {
		return "", errors.InternalError("Failed to generate token")
	}

	return signedToken, nil
}

// Authenticate middleware function
func (m *AuthMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return api.SendError(c, http.StatusUnauthorized, errors.CodeUnauthorized, "Missing authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return api.SendError(c, http.StatusUnauthorized, errors.CodeUnauthorized, "Invalid authorization format")
		}
		tokenString := parts[1]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.config.JWT.Secret), nil
		})

		if err != nil {
			if err == jwt.ErrTokenExpired {
				return api.SendError(c, http.StatusUnauthorized, errors.CodeInvalidToken, "Token has expired")
			}
			return api.SendError(c, http.StatusUnauthorized, errors.CodeInvalidToken, "Invalid token")
		}

		if !token.Valid {
			return api.SendError(c, http.StatusUnauthorized, errors.CodeInvalidToken, "Invalid token")
		}

		user, err := m.store.Users().GetByID(claims.UserID)
		if err != nil {
			if errors.IsCode(err, errors.CodeNotFound) {
				return api.SendError(c, http.StatusUnauthorized, errors.CodeUnauthorized, "User not found")
			}
			return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Error verifying user")
		}

		if !user.Authenticated {
			return api.SendError(c, http.StatusUnauthorized, errors.CodeUnauthorized, "User not authenticated")
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("email", user.Email)

		return next(c)
	}
}

// OptionalAuth middleware that doesn't require authentication but will process it if present
func (m *AuthMiddleware) OptionalAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return next(c)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return next(c)
		}

		tokenString := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.config.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			return next(c)
		}

		user, err := m.store.Users().GetByID(claims.UserID)
		if err != nil || user == nil || !user.Authenticated {
			return next(c)
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("email", user.Email)

		return next(c)
	}
}

// GetUserFromContext retrieves the authenticated user from the context
func GetUserFromContext(c echo.Context) *models.User {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		return nil
	}
	return user
}
