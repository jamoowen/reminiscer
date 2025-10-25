package api

import (
	"github.com/labstack/echo/v4"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SendSuccess sends a success response
func SendSuccess(c echo.Context, status int, data interface{}) error {
	return c.JSON(status, &Response{
		Success: true,
		Data:    data,
	})
}

// SendError sends an error response
func SendError(c echo.Context, status int, code string, message string) error {
	return c.JSON(status, &ErrorResponse{
		Success: false,
		Code:    code,
		Message: message,
	})
}
