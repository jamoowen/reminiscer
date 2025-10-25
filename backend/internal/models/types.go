package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	Password      string    `json:"-"` // Never send password in JSON
	Authenticated bool      `json:"authenticated"`
	CreatedAt     time.Time `json:"created_at"`
}

// UserStore handles all database operations for users
type UserStore interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Authenticate(email, password string) (*User, error)
}

// Group represents a group in the system
type Group struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"group_id"`
	Name      string    `json:"name"`
	MemberID  string    `json:"member_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GroupStore handles all database operations for groups
type GroupStore interface {
	Create(group *Group) error
	GetByID(id string) (*Group, error)
	GetByGroupID(groupID string) ([]*Group, error)
	GetByMemberID(memberID string) ([]*Group, error)
	Update(group *Group) error
	Delete(id string) error
}

// Quote represents a quote in the system
type Quote struct {
	ID         string    `json:"id"`
	Text       string    `json:"text"`
	Author     string    `json:"author,omitempty"`
	UploaderID string    `json:"uploader_id"`
	GroupID    string    `json:"group_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// QuoteFilter represents the filtering options for quotes
type QuoteFilter struct {
	Author string
	Page   int
	Limit  int
}

// QuoteStore handles all database operations for quotes
type QuoteStore interface {
	Create(quote *Quote) error
	GetByID(id string) (*Quote, error)
	GetRandom(filter QuoteFilter) (*Quote, error)
	List(filter QuoteFilter) ([]*Quote, error)
	Update(quote *Quote) error
	Delete(id string) error
}

// Store combines all storage interfaces
type Store interface {
	Users() UserStore
	Groups() GroupStore
	Quotes() QuoteStore
}
