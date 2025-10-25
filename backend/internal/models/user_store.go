package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jamoowen/reminiscer/internal/errors"
	"golang.org/x/crypto/bcrypt"
)

// SQLiteUserStore implements UserStore interface
type SQLiteUserStore struct {
	db *sql.DB
}

// NewSQLiteUserStore creates a new SQLite user store
func NewSQLiteUserStore(db *sql.DB) *SQLiteUserStore {
	return &SQLiteUserStore{db: db}
}

// Create inserts a new user into the database
func (s *SQLiteUserStore) Create(user *User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.InternalError("Failed to hash password")
	}

	query := `
		INSERT INTO users (id, email, username, hashed_password, authenticated, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		user.ID,
		user.Email,
		user.Username,
		hashedPassword,
		user.Authenticated,
		time.Now(),
	)

	if err != nil {
		return errors.DatabaseError("Failed to create user")
	}

	return nil
}

// GetByID retrieves a user by their ID
func (s *SQLiteUserStore) GetByID(id string) (*User, error) {
	var user User
	query := `
		SELECT id, email, username, authenticated, created_at
		FROM users
		WHERE id = ?
	`

	err := s.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Authenticated,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("User not found")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get user")
	}

	return &user, nil
}

// GetByEmail retrieves a user by their email
func (s *SQLiteUserStore) GetByEmail(email string) (*User, error) {
	var user User
	query := `
		SELECT id, email, username, authenticated, created_at
		FROM users
		WHERE email = ?
	`

	err := s.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Authenticated,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // User not found is not an error in this case
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get user by email")
	}

	return &user, nil
}

// Update updates an existing user
func (s *SQLiteUserStore) Update(user *User) error {
	var query string
	var args []interface{}

	// If password is provided, update it
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.InternalError("Failed to hash password")
		}

		query = `
			UPDATE users
			SET email = ?, username = ?, hashed_password = ?, authenticated = ?
			WHERE id = ?
		`
		args = []interface{}{
			user.Email,
			user.Username,
			hashedPassword,
			user.Authenticated,
			user.ID,
		}
	} else {
		query = `
			UPDATE users
			SET email = ?, username = ?, authenticated = ?
			WHERE id = ?
		`
		args = []interface{}{
			user.Email,
			user.Username,
			user.Authenticated,
			user.ID,
		}
	}

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return errors.DatabaseError("Failed to update user")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to check update result")
	}

	if rows == 0 {
		return errors.NotFound("User not found")
	}

	return nil
}

// Authenticate verifies user credentials and returns the user if valid
func (s *SQLiteUserStore) Authenticate(email, password string) (*User, error) {
	var user User
	var hashedPassword string

	query := `
		SELECT id, email, username, hashed_password, authenticated, created_at
		FROM users
		WHERE email = ?
	`

	err := s.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&hashedPassword,
		&user.Authenticated,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.InvalidCredentials("Invalid email or password")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to authenticate user")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return nil, errors.InvalidCredentials("Invalid email or password")
	}

	return &user, nil
}
