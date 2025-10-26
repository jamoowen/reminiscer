package handlers

import (
	"net/http"

	"github.com/jamoowen/reminiscer/internal/api"
	"github.com/jamoowen/reminiscer/internal/errors"
	"github.com/jamoowen/reminiscer/internal/middleware"
	"github.com/jamoowen/reminiscer/internal/models"
	"github.com/labstack/echo/v4"
)

type QuoteHandler struct {
	store   models.Store
	authMid *middleware.AuthMiddleware
}

func NewQuoteHandler(store models.Store, authMid *middleware.AuthMiddleware) *QuoteHandler {
	return &QuoteHandler{
		store:   store,
		authMid: authMid,
	}
}

// SetupRoutes sets up the quote routes
func (h *QuoteHandler) SetupRoutes(e *echo.Echo) {
	quotes := e.Group("/quotes", h.authMid.Authenticate)
	quotes.POST("", h.Create)
	quotes.GET("", h.List)
	quotes.GET("/random", h.GetRandom)
	quotes.PATCH("/:id", h.Update)
	quotes.DELETE("/:id", h.Delete)
}

// Create handles creating a new quote
func (h *QuoteHandler) Create(c echo.Context) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return api.SendError(c, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required")
	}

	var req CreateQuoteRequest
	if err := c.Bind(&req); err != nil {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Invalid request data")
	}

	// Verify group exists and user is a member
	groups, err := h.store.Groups().GetByGroupID(req.GroupID)
	if err != nil {
		if errors.IsCode(err, errors.CodeNotFound) {
			return api.SendError(c, http.StatusNotFound, errors.CodeNotFound, "Group not found")
		}
		return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to verify group")
	}

	isMember := false
	for _, g := range groups {
		if g.MemberID == user.ID {
			isMember = true
			break
		}
	}
	if !isMember {
		return api.SendError(c, http.StatusForbidden, errors.CodeForbidden, "Not a member of this group")
	}

	quote := &models.Quote{
		Text:       req.Text,
		Author:     req.Author,
		UploaderID: user.ID,
		GroupID:    req.GroupID,
	}

	if err := h.store.Quotes().Create(quote); err != nil {
		return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to create quote")
	}

	return api.SendSuccess(c, http.StatusCreated, toQuoteResponse(quote, user.Username))
}

// List handles retrieving quotes with optional filtering
func (h *QuoteHandler) List(c echo.Context) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return api.SendError(c, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required")
	}

	params := ListQuotesParams{}
	if err := c.Bind(&params); err != nil {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Invalid query parameters")
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 50 {
		params.Limit = 10
	}

	filter := models.QuoteFilter{
		Author: params.Author,
		Page:   params.Page,
		Limit:  params.Limit,
	}

	quotes, err := h.store.Quotes().List(filter)
	if err != nil {
		if errors.IsCode(err, errors.CodeNotFound) {
			return api.SendSuccess(c, http.StatusOK, []interface{}{}) // Empty list instead of error
		}
		return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to retrieve quotes")
	}

	// Get username lookup function
	getUsernameFn := func(userID string) string {
		u, err := h.store.Users().GetByID(userID)
		if err != nil || u == nil {
			return "Unknown"
		}
		return u.Username
	}

	return api.SendSuccess(c, http.StatusOK, toQuoteResponses(quotes, getUsernameFn))
}

// GetRandom handles retrieving a random quote
func (h *QuoteHandler) GetRandom(c echo.Context) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return api.SendError(c, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required")
	}

	filter := models.QuoteFilter{
		Author: c.QueryParam("author"),
	}

	quote, err := h.store.Quotes().GetRandom(filter)
	if err != nil {
		if errors.IsCode(err, errors.CodeNotFound) {
			return api.SendError(c, http.StatusNotFound, errors.CodeNotFound, "No quotes found")
		}
		return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to get random quote")
	}

	uploader, err := h.store.Users().GetByID(quote.UploaderID)
	if err != nil {
		return api.SendSuccess(c, http.StatusOK, toQuoteResponse(quote, "Unknown"))
	}

	return api.SendSuccess(c, http.StatusOK, toQuoteResponse(quote, uploader.Username))
}

// Update handles updating a quote
func (h *QuoteHandler) Update(c echo.Context) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return api.SendError(c, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required")
	}

	id := c.Param("id")
	if id == "" {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Quote ID is required")
	}

	quote, err := h.store.Quotes().GetByID(id)
	if err != nil {
		if errors.IsCode(err, errors.CodeNotFound) {
			return api.SendError(c, http.StatusNotFound, errors.CodeNotFound, "Quote not found")
		}
		return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to retrieve quote")
	}

	if quote.UploaderID != user.ID {
		return api.SendError(c, http.StatusForbidden, errors.CodeForbidden, "Not authorized to update this quote")
	}

	var req UpdateQuoteRequest
	if err := c.Bind(&req); err != nil {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Invalid request data")
	}

	quote.Text = req.Text
	quote.Author = req.Author

	if err := h.store.Quotes().Update(quote); err != nil {
		return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to update quote")
	}

	return api.SendSuccess(c, http.StatusOK, toQuoteResponse(quote, user.Username))
}

// Delete handles deleting a quote
func (h *QuoteHandler) Delete(c echo.Context) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return api.SendError(c, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required")
	}

	id := c.Param("id")
	if id == "" {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Quote ID is required")
	}

	quote, err := h.store.Quotes().GetByID(id)
	if err != nil {
		if errors.IsCode(err, errors.CodeNotFound) {
			return api.SendError(c, http.StatusNotFound, errors.CodeNotFound, "Quote not found")
		}
		return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to retrieve quote")
	}

	if quote.UploaderID != user.ID {
		return api.SendError(c, http.StatusForbidden, errors.CodeForbidden, "Not authorized to delete this quote")
	}

	if err := h.store.Quotes().Delete(id); err != nil {
		if errors.IsCode(err, errors.CodeNotFound) {
			return api.SendError(c, http.StatusNotFound, errors.CodeNotFound, "Quote not found")
		}
		return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to delete quote")
	}

	return api.SendSuccess(c, http.StatusOK, nil)
}
