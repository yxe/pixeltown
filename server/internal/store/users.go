package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User is one row of the users table.
type User struct {
	ID        uuid.UUID
	Username  string
	Email     string
	CreatedAt time.Time
}

// CreateUser inserts a new row and returns the populated User.
func (s *Store) CreateUser(
	ctx context.Context,
	username, email string,
) (*User, error) {
	return nil, nil
}

// FindUserByEmail returns the user with the given email, or
// (nil, nil) when no row matches.
func (s *Store) FindUserByEmail(
	ctx context.Context,
	email string,
) (*User, error) {
	return nil, nil
}
