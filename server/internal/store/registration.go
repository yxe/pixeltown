package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// CreateUserWithPasskey atomically inserts a user and their first
// passkey in a single transaction. Returns ErrUserExists on a
// username or email collision. On success u.CreatedAt, p.ID, and
// p.CreatedAt are populated.
func (s *Store) CreateUserWithPasskey(
	ctx context.Context,
	u *User,
	p *Passkey,
) error {
	tx, err := s.pool.Begin(ctx)

	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx, `
		INSERT INTO users (id, username, email)
		VALUES ($1, $2, $3)
		RETURNING created_at
	`, u.ID, u.Username, u.Email).Scan(&u.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrUserExists
		}

		return err
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO passkeys
			(user_id, credential_id, public_key, sign_count,
			 transports, backup_eligible, backup_state, nickname)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`,
		p.UserID, p.CredentialID, p.PublicKey, p.SignCount,
		p.Transports, p.BackupEligible, p.BackupState, p.Nickname,
	).Scan(&p.ID, &p.CreatedAt)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
