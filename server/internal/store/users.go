package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrUserExists is returned by CreateUser when the username or
// email already exists.
var ErrUserExists = errors.New("user already exists")

// User is one row of the users table.
type User struct {
	ID        uuid.UUID
	Username  string
	Email     string
	CreatedAt time.Time
}

// CreateUser inserts a new row and returns the populated User.
// Returns ErrUserExists on a unique-constraint violation.
func (s *Store) CreateUser(
	ctx context.Context,
	username, email string,
) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (username, email)
		VALUES ($1, $2)
		RETURNING id, username, email, created_at
	`, username, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUserExists
		}

		return nil, err
	}

	return &u, nil
}

// FindUserByEmail returns the user with the given email, or
// (nil, nil) when no row matches.
func (s *Store) FindUserByEmail(
	ctx context.Context,
	email string,
) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT id, username, email, created_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &u, nil
}

// FindUserByID returns the user with the given id, or (nil, nil)
// when no row matches.
func (s *Store) FindUserByID(
	ctx context.Context,
	id uuid.UUID,
) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT id, username, email, created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &u, nil
}
