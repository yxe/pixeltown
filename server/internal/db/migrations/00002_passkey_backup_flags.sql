-- +goose Up
-- +goose StatementBegin

ALTER TABLE passkeys
  ADD COLUMN backup_eligible boolean NOT NULL DEFAULT false,
  ADD COLUMN backup_state    boolean NOT NULL DEFAULT false;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE passkeys
  DROP COLUMN backup_state,
  DROP COLUMN backup_eligible;

-- +goose StatementEnd
