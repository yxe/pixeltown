package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

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

func (s *Store) CreatePasskey(ctx context.Context, p *Passkey) error {
	return nil
}

func (s *Store) FindPasskeysForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]Passkey, error) {
	return nil, nil
}

func (s *Store) FindPasskeyByCredentialID(
	ctx context.Context,
	credID []byte,
) (*Passkey, error) {
	return nil, nil
}

func (s *Store) UpdatePasskeySignCount(
	ctx context.Context,
	id uuid.UUID,
	newCount uint32,
) error {
	return nil
}
