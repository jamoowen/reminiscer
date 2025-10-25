package handlers

import (
	"time"

	"github.com/jamoowen/reminiscer/internal/models"
)

// AuthRequest represents authentication request data
type AuthRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest represents user registration request data
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=6"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// CreateQuoteRequest represents the request to create a quote
type CreateQuoteRequest struct {
	Text    string `json:"text" validate:"required"`
	Author  string `json:"author" validate:"required"`
	GroupID string `json:"group_id" validate:"required"`
}

// UpdateQuoteRequest represents the request to update a quote
type UpdateQuoteRequest struct {
	Text   string `json:"text" validate:"required"`
	Author string `json:"author" validate:"required"`
}

// CreateGroupRequest represents the request to create a group
type CreateGroupRequest struct {
	Name    string   `json:"name" validate:"required"`
	GroupID string   `json:"group_id" validate:"required"`
	Members []string `json:"members" validate:"required,min=1"`
}

// UpdateGroupRequest represents the request to update a group
type UpdateGroupRequest struct {
	Name    string   `json:"name" validate:"required"`
	Members []string `json:"members" validate:"required,min=1"`
}

// QuoteResponse represents a quote with additional metadata
type QuoteResponse struct {
	ID         string    `json:"id"`
	Text       string    `json:"text"`
	Author     string    `json:"author"`
	UploaderID string    `json:"uploader_id"`
	GroupID    string    `json:"group_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Uploader   string    `json:"uploader"` // Username of uploader
}

// GroupResponse represents a group with additional metadata
type GroupResponse struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"group_id"`
	Name      string    `json:"name"`
	Members   []string  `json:"members"` // List of member usernames
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListQuotesParams represents the query parameters for listing quotes
type ListQuotesParams struct {
	Author string `query:"author"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

// toQuoteResponse converts a models.Quote to a QuoteResponse
func toQuoteResponse(q *models.Quote, uploaderUsername string) *QuoteResponse {
	return &QuoteResponse{
		ID:         q.ID,
		Text:       q.Text,
		Author:     q.Author,
		UploaderID: q.UploaderID,
		GroupID:    q.GroupID,
		CreatedAt:  q.CreatedAt,
		UpdatedAt:  q.UpdatedAt,
		Uploader:   uploaderUsername,
	}
}

// toQuoteResponses converts a slice of models.Quote to QuoteResponses
func toQuoteResponses(quotes []*models.Quote, getUsernameFn func(string) string) []*QuoteResponse {
	responses := make([]*QuoteResponse, len(quotes))
	for i, q := range quotes {
		responses[i] = toQuoteResponse(q, getUsernameFn(q.UploaderID))
	}
	return responses
}

// toGroupResponse converts models.Group slice to a GroupResponse
func toGroupResponse(groups []*models.Group, getUsernameFn func(string) string) *GroupResponse {
	if len(groups) == 0 {
		return nil
	}

	// All groups in the slice should have the same GroupID
	first := groups[0]
	members := make([]string, len(groups))
	for i, g := range groups {
		members[i] = getUsernameFn(g.MemberID)
	}

	return &GroupResponse{
		ID:        first.ID,
		GroupID:   first.GroupID,
		Name:      first.Name,
		Members:   members,
		CreatedAt: first.CreatedAt,
		UpdatedAt: first.UpdatedAt,
	}
}
