package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jamoowen/reminiscer/internal/errors"
)

// SQLiteGroupStore implements GroupStore interface
type SQLiteGroupStore struct {
	db *sql.DB
}

// NewSQLiteGroupStore creates a new SQLite group store
func NewSQLiteGroupStore(db *sql.DB) *SQLiteGroupStore {
	return &SQLiteGroupStore{db: db}
}

// Create inserts a new group into the database
func (s *SQLiteGroupStore) Create(group *Group) error {
	if group.ID == "" {
		group.ID = uuid.New().String()
	}

	now := time.Now()
	group.CreatedAt = now
	group.UpdatedAt = now

	query := `
		INSERT INTO groups (id, group_id, name, member_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		group.ID,
		group.GroupID,
		group.Name,
		group.MemberID,
		group.CreatedAt,
		group.UpdatedAt,
	)

	if err != nil {
		return errors.DatabaseError("Failed to create group")
	}

	return nil
}

// GetByID retrieves a group by its ID
func (s *SQLiteGroupStore) GetByID(id string) (*Group, error) {
	var group Group
	query := `
		SELECT id, group_id, name, member_id, created_at, updated_at
		FROM groups
		WHERE id = ?
	`

	err := s.db.QueryRow(query, id).Scan(
		&group.ID,
		&group.GroupID,
		&group.Name,
		&group.MemberID,
		&group.CreatedAt,
		&group.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Group not found")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get group")
	}

	return &group, nil
}

// GetByGroupID retrieves all groups with the given group_id
func (s *SQLiteGroupStore) GetByGroupID(groupID string) ([]*Group, error) {
	query := `
		SELECT id, group_id, name, member_id, created_at, updated_at
		FROM groups
		WHERE group_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, groupID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to get groups")
	}
	defer rows.Close()

	var groups []*Group
	for rows.Next() {
		var group Group
		err := rows.Scan(
			&group.ID,
			&group.GroupID,
			&group.Name,
			&group.MemberID,
			&group.CreatedAt,
			&group.UpdatedAt,
		)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan group data")
		}
		groups = append(groups, &group)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseError("Error iterating through groups")
	}

	if len(groups) == 0 {
		return nil, errors.NotFound("No groups found")
	}

	return groups, nil
}

// GetByMemberID retrieves all groups for a specific member
func (s *SQLiteGroupStore) GetByMemberID(memberID string) ([]*Group, error) {
	query := `
		SELECT id, group_id, name, member_id, created_at, updated_at
		FROM groups
		WHERE member_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, memberID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to get member groups")
	}
	defer rows.Close()

	var groups []*Group
	for rows.Next() {
		var group Group
		err := rows.Scan(
			&group.ID,
			&group.GroupID,
			&group.Name,
			&group.MemberID,
			&group.CreatedAt,
			&group.UpdatedAt,
		)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan group data")
		}
		groups = append(groups, &group)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseError("Error iterating through groups")
	}

	if len(groups) == 0 {
		return nil, errors.NotFound("No groups found for member")
	}

	return groups, nil
}

// Update updates an existing group
func (s *SQLiteGroupStore) Update(group *Group) error {
	group.UpdatedAt = time.Now()

	query := `
		UPDATE groups
		SET name = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.Exec(query,
		group.Name,
		group.UpdatedAt,
		group.ID,
	)
	if err != nil {
		return errors.DatabaseError("Failed to update group")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to check update result")
	}

	if rows == 0 {
		return errors.NotFound("Group not found")
	}

	return nil
}

// Delete removes a group from the database
func (s *SQLiteGroupStore) Delete(id string) error {
	query := `DELETE FROM groups WHERE id = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return errors.DatabaseError("Failed to delete group")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to check delete result")
	}

	if rows == 0 {
		return errors.NotFound("Group not found")
	}

	return nil
}
