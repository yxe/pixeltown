package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"

	"pixeltown/server/internal/store"
)

type Auth struct {
	Store    *store.Store
	WebAuthn *webauthn.WebAuthn
	Sessions *SessionStore
}

func New(s *store.Store) (*Auth, error) {
	wa, err := webauthn.New(&webauthn.Config{
		RPID: env("RPID", "localhost"),
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

type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]sessionEntry
}

type sessionEntry struct {
	Data    webauthn.SessionData
	UserID  uuid.UUID
	Expires time.Time
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]sessionEntry)}
}

func (s *SessionStore) Put(
	data webauthn.SessionData,
	userID uuid.UUID,
	ttl time.Duration,
) string {
	id := randomID()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = sessionEntry{
		Data:    data,
		UserID:  userID,
		Expires: time.Now().Add(ttl),
	}
	s.gc()
	return id
}

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
