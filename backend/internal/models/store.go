package models

import (
	"database/sql"
)

// SQLiteStore implements Store interface and combines all SQLite store implementations
type SQLiteStore struct {
	db         *sql.DB
	userStore  *SQLiteUserStore
	groupStore *SQLiteGroupStore
	quoteStore *SQLiteQuoteStore
}

// NewSQLiteStore creates a new SQLite store that implements the Store interface
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{
		db:         db,
		userStore:  NewSQLiteUserStore(db),
		groupStore: NewSQLiteGroupStore(db),
		quoteStore: NewSQLiteQuoteStore(db),
	}
}

// Users returns the UserStore implementation
func (s *SQLiteStore) Users() UserStore {
	return s.userStore
}

// Groups returns the GroupStore implementation
func (s *SQLiteStore) Groups() GroupStore {
	return s.groupStore
}

// Quotes returns the QuoteStore implementation
func (s *SQLiteStore) Quotes() QuoteStore {
	return s.quoteStore
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
