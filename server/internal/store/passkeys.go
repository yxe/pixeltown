package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Passkey is one row of the passkeys table.
type Passkey struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	CredentialID   []byte
	PublicKey      []byte
	SignCount      uint32
	Transports     []string
	BackupEligible bool
	BackupState    bool
	Nickname       *string
	CreatedAt      time.Time
	LastUsedAt     *time.Time
}

// CreatePasskey inserts a new credential for an existing user. On
// success the ID and CreatedAt fields are filled in on p.
func (s *Store) CreatePasskey(ctx context.Context, p *Passkey) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO passkeys
			(user_id, credential_id, public_key, sign_count,
			 transports, backup_eligible, backup_state, nickname)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`,
		p.UserID, p.CredentialID, p.PublicKey, p.SignCount,
		p.Transports, p.BackupEligible, p.BackupState, p.Nickname,
	).Scan(&p.ID, &p.CreatedAt)
}

// FindPasskeysForUser returns every credential belonging to userID.
func (s *Store) FindPasskeysForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]Passkey, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, credential_id, public_key, sign_count,
		       transports, backup_eligible, backup_state, nickname,
		       created_at, last_used_at
		FROM passkeys WHERE user_id = $1
	`, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var out []Passkey
	for rows.Next() {
		var p Passkey

		if err := rows.Scan(
			&p.ID, &p.UserID, &p.CredentialID, &p.PublicKey,
			&p.SignCount, &p.Transports,
			&p.BackupEligible, &p.BackupState,
			&p.Nickname, &p.CreatedAt, &p.LastUsedAt,
		); err != nil {
			return nil, err
		}

		out = append(out, p)
	}

	return out, rows.Err()
}

// FindPasskeyByCredentialID returns the passkey matching credID,
// or (nil, nil) when no row matches.
func (s *Store) FindPasskeyByCredentialID(
	ctx context.Context,
	credID []byte,
) (*Passkey, error) {
	var p Passkey
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, credential_id, public_key, sign_count,
		       transports, backup_eligible, backup_state, nickname,
		       created_at, last_used_at
		FROM passkeys WHERE credential_id = $1
	`, credID).Scan(
		&p.ID, &p.UserID, &p.CredentialID, &p.PublicKey,
		&p.SignCount, &p.Transports,
		&p.BackupEligible, &p.BackupState,
		&p.Nickname, &p.CreatedAt, &p.LastUsedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &p, nil
}

// UpdatePasskeyAfterLogin bumps sign_count, refreshes backup_state
// (it can toggle when a device backs up or restores a credential),
// and stamps last_used_at.
func (s *Store) UpdatePasskeyAfterLogin(
	ctx context.Context,
	id uuid.UUID,
	newCount uint32,
	backupState bool,
) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE passkeys
		SET sign_count = $1, backup_state = $2, last_used_at = now()
		WHERE id = $3
	`, newCount, backupState, id)
	return err
}
