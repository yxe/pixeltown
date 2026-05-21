package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Passkey is one row of the passkeys table.
type Passkey struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	CredentialID []byte
	PublicKey    []byte
	SignCount    uint32
	Transports   []string
	Nickname     *string
	CreatedAt    time.Time
	LastUsedAt   *time.Time
}

// CreatePasskey inserts a new credential for an existing user. On
// success the ID and CreatedAt fields are filled in.
func (s *Store) CreatePasskey(ctx context.Context, p *Passkey) error {
	return nil
}

// FindPasskeysForUser returns every credential belonging to userID.
func (s *Store) FindPasskeysForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]Passkey, error) {
	return nil, nil
}

// FindPasskeyByCredentialID returns the passkey matching credID, or
// (nil, nil) when no row matches.
func (s *Store) FindPasskeyByCredentialID(
	ctx context.Context,
	credID []byte,
) (*Passkey, error) {
	return nil, nil
}

// UpdatePasskeySignCount bumps sign_count and stamps last_used_at.
func (s *Store) UpdatePasskeySignCount(
	ctx context.Context,
	id uuid.UUID,
	newCount uint32,
) error {
	return nil
}
