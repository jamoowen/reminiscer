package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jamoowen/reminiscer/internal/errors"
)

// SQLiteQuoteStore implements QuoteStore interface
type SQLiteQuoteStore struct {
	db *sql.DB
}

// NewSQLiteQuoteStore creates a new SQLite quote store
func NewSQLiteQuoteStore(db *sql.DB) *SQLiteQuoteStore {
	return &SQLiteQuoteStore{db: db}
}

// Create inserts a new quote into the database
func (s *SQLiteQuoteStore) Create(quote *Quote) error {
	if quote.ID == "" {
		quote.ID = uuid.New().String()
	}

	now := time.Now()
	quote.CreatedAt = now
	quote.UpdatedAt = now

	query := `
		INSERT INTO quotes (id, text, author, uploader_id, group_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		quote.ID,
		quote.Text,
		quote.Author,
		quote.UploaderID,
		quote.GroupID,
		quote.CreatedAt,
		quote.UpdatedAt,
	)

	if err != nil {
		return errors.DatabaseError("Failed to create quote")
	}

	return nil
}

// GetByID retrieves a quote by its ID
func (s *SQLiteQuoteStore) GetByID(id string) (*Quote, error) {
	var quote Quote
	query := `
		SELECT id, text, author, uploader_id, group_id, created_at, updated_at
		FROM quotes
		WHERE id = ?
	`

	err := s.db.QueryRow(query, id).Scan(
		&quote.ID,
		&quote.Text,
		&quote.Author,
		&quote.UploaderID,
		&quote.GroupID,
		&quote.CreatedAt,
		&quote.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Quote not found")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get quote")
	}

	return &quote, nil
}

// GetRandom retrieves a random quote, optionally filtered by author
func (s *SQLiteQuoteStore) GetRandom(filter QuoteFilter) (*Quote, error) {
	var query string
	var args []interface{}

	if filter.Author != "" {
		query = `
			SELECT id, text, author, uploader_id, group_id, created_at, updated_at
			FROM quotes
			WHERE author = ?
			ORDER BY RANDOM()
			LIMIT 1
		`
		args = append(args, filter.Author)
	} else {
		query = `
			SELECT id, text, author, uploader_id, group_id, created_at, updated_at
			FROM quotes
			ORDER BY RANDOM()
			LIMIT 1
		`
	}

	var quote Quote
	err := s.db.QueryRow(query, args...).Scan(
		&quote.ID,
		&quote.Text,
		&quote.Author,
		&quote.UploaderID,
		&quote.GroupID,
		&quote.CreatedAt,
		&quote.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("No quotes found")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get random quote")
	}

	return &quote, nil
}

// List retrieves quotes with pagination and optional author filter
func (s *SQLiteQuoteStore) List(filter QuoteFilter) ([]*Quote, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}

	offset := (filter.Page - 1) * filter.Limit

	var query string
	var args []interface{}

	if filter.Author != "" {
		query = `
			SELECT id, text, author, uploader_id, group_id, created_at, updated_at
			FROM quotes
			WHERE author = ?
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`
		args = append(args, filter.Author, filter.Limit, offset)
	} else {
		query = `
			SELECT id, text, author, uploader_id, group_id, created_at, updated_at
			FROM quotes
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`
		args = append(args, filter.Limit, offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list quotes")
	}
	defer rows.Close()

	var quotes []*Quote
	for rows.Next() {
		var quote Quote
		err := rows.Scan(
			&quote.ID,
			&quote.Text,
			&quote.Author,
			&quote.UploaderID,
			&quote.GroupID,
			&quote.CreatedAt,
			&quote.UpdatedAt,
		)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan quote data")
		}
		quotes = append(quotes, &quote)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseError("Error iterating through quotes")
	}

	return quotes, nil
}

// Update updates an existing quote
func (s *SQLiteQuoteStore) Update(quote *Quote) error {
	quote.UpdatedAt = time.Now()

	query := `
		UPDATE quotes
		SET text = ?, author = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.Exec(query,
		quote.Text,
		quote.Author,
		quote.UpdatedAt,
		quote.ID,
	)
	if err != nil {
		return errors.DatabaseError("Failed to update quote")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to check update result")
	}

	if rows == 0 {
		return errors.NotFound("Quote not found")
	}

	return nil
}

// Delete removes a quote from the database
func (s *SQLiteQuoteStore) Delete(id string) error {
	query := `DELETE FROM quotes WHERE id = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return errors.DatabaseError("Failed to delete quote")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to check delete result")
	}

	if rows == 0 {
		return errors.NotFound("Quote not found")
	}

	return nil
}
