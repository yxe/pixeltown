-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
  id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  username   citext      NOT NULL UNIQUE,
  email      citext      NOT NULL UNIQUE,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE passkeys (
  id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  credential_id bytea       NOT NULL UNIQUE,
  public_key    bytea       NOT NULL,
  sign_count    bigint      NOT NULL DEFAULT 0,
  transports    text[]      NOT NULL DEFAULT '{}',
  nickname      text,
  created_at    timestamptz NOT NULL DEFAULT now(),
  last_used_at  timestamptz
);

CREATE INDEX passkeys_user_id_idx ON passkeys(user_id);

CREATE TABLE recovery_codes (
  id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  code_hash  text        NOT NULL,
  used_at    timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX recovery_codes_user_id_idx ON recovery_codes(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS recovery_codes;
DROP TABLE IF EXISTS passkeys;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd
