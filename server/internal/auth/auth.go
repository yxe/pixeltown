// Package auth wires the WebAuthn library to the store + holds
// an in-memory challenge store keyed by opaque session id.
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"pixeltown/server/internal/store"
)

// Auth bundles the dependencies the HTTP handlers need.
type Auth struct {
	Store    *store.Store
	WebAuthn *webauthn.WebAuthn
	Sessions *SessionStore
}

// New constructs an Auth from a Store, reading RP config from env
// (RPID, RP_DISPLAY_NAME, RP_ORIGIN_WEB, RP_ORIGIN_API) with
// dev-friendly defaults.
func New(s *store.Store) (*Auth, error) {
	wa, err := webauthn.New(&webauthn.Config{
		RPID:          env("RPID", "localhost"),
		RPDisplayName: env("RP_DISPLAY_NAME", "Pixeltown (dev)"),
		RPOrigins: []string{
			env("RP_ORIGIN_WEB", "http://localhost:5173"),
			env("RP_ORIGIN_API", "http://localhost:4000"),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("webauthn config: %w", err)
	}

	return &Auth{
		Store:    s,
		WebAuthn: wa,
		Sessions: NewSessionStore(),
	}, nil
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

type webauthnUser struct {
	User  *store.User
	Creds []webauthn.Credential
}

func (u *webauthnUser) WebAuthnID() []byte {
	b, _ := u.User.ID.MarshalBinary()
	return b
}

func (u *webauthnUser) WebAuthnName() string {
	return u.User.Username
}

func (u *webauthnUser) WebAuthnDisplayName() string {
	return u.User.Username
}

func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.Creds
}

// passkeysToCredentials adapts persisted Passkey rows to the
// Credential shape the webauthn lib expects for login.
func passkeysToCredentials(ps []store.Passkey) []webauthn.Credential {
	out := make([]webauthn.Credential, len(ps))

	for i, p := range ps {
		transports := make([]protocol.AuthenticatorTransport, len(p.Transports))

		for j, t := range p.Transports {
			transports[j] = protocol.AuthenticatorTransport(t)
		}

		out[i] = webauthn.Credential{
			ID:        p.CredentialID,
			PublicKey: p.PublicKey,
			Transport: transports,
			Flags: webauthn.CredentialFlags{
				BackupEligible: p.BackupEligible,
				BackupState:    p.BackupState,
			},
			Authenticator: webauthn.Authenticator{
				SignCount: p.SignCount,
			},
		}
	}

	return out
}

// SessionStore holds in-flight WebAuthn challenges between the
// begin and finish steps of a registration or login.
type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]sessionEntry
}

type sessionEntry struct {
	Data    webauthn.SessionData
	User    *store.User
	Expires time.Time
}

// NewSessionStore returns an empty SessionStore.
func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]sessionEntry)}
}

// Put stores data + the staged user under a fresh opaque id and
// returns it. The entry expires after ttl.
func (s *SessionStore) Put(
	data webauthn.SessionData,
	user *store.User,
	ttl time.Duration,
) string {
	id := randomID()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = sessionEntry{
		Data:    data,
		User:    user,
		Expires: time.Now().Add(ttl),
	}
	s.gc()
	return id
}

// Take removes and returns the entry for id, or (nil, false) on
// miss or expiry. Entries are single-use.
func (s *SessionStore) Take(id string) (*sessionEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.sessions[id]

	if !ok {
		return nil, false
	}

	delete(s.sessions, id)

	if time.Now().After(entry.Expires) {
		return nil, false
	}

	return &entry, true
}

func (s *SessionStore) gc() {
	now := time.Now()

	for id, e := range s.sessions {
		if now.After(e.Expires) {
			delete(s.sessions, id)
		}
	}
}

func randomID() string {
	var b [24]byte
	_, _ = rand.Read(b[:])
	return base64.RawURLEncoding.EncodeToString(b[:])
}
