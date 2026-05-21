package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Username  string
	Email     string
	CreatedAt time.Time
}

func (s *Store) CreateUser(
	ctx context.Context,
	username, email string,
) (*User, error) {
	return nil, nil
}

func (s *Store) FindUserByEmail(
	ctx context.Context,
	email string,
) (*User, error) {
	return nil, nil
}
