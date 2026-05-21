package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"pixeltown/server/internal/store"
)

const sessionTTL = 5 * time.Minute

type registerBeginReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type beginResp struct {
	SessionID string `json:"sessionId"`
	Options   any    `json:"options"`
}

// RegisterBegin starts a WebAuthn registration. Generates a fresh
// user id, stages the user in memory only, and returns the options
// the browser hands to navigator.credentials.create. Nothing is
// written to the DB until RegisterFinish verifies the credential.
func (a *Auth) RegisterBegin(w http.ResponseWriter, r *http.Request) {
	var body registerBeginReq

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	body.Username = strings.TrimSpace(body.Username)
	body.Email = strings.TrimSpace(body.Email)

	if body.Username == "" || len(body.Username) > 32 {
		http.Error(w, "invalid username", http.StatusBadRequest)
		return
	}

	if body.Email == "" || !strings.Contains(body.Email, "@") {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}

	user := &store.User{
		ID:       uuid.New(),
		Username: body.Username,
		Email:    body.Email,
	}

	wu := &webauthnUser{User: user}

	options, session, err := a.WebAuthn.BeginRegistration(wu)

	if err != nil {
		log.Printf("register begin: %v", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	sid := a.Sessions.Put(*session, user, sessionTTL)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(beginResp{
		SessionID: sid,
		Options:   options,
	})
}

// RegisterFinish verifies the credential the browser produced and
// persists the user + passkey atomically. The session id arrives
// in the X-Pixeltown-Session header; the request body is the raw
// WebAuthn credential payload.
func (a *Auth) RegisterFinish(w http.ResponseWriter, r *http.Request) {
	sid := r.Header.Get("X-Pixeltown-Session")

	if sid == "" {
		http.Error(w, "missing session", http.StatusBadRequest)
		return
	}

	entry, ok := a.Sessions.Take(sid)

	if !ok {
		http.Error(w, "session expired", http.StatusUnauthorized)
		return
	}

	wu := &webauthnUser{User: entry.User}

	cred, err := a.WebAuthn.FinishRegistration(wu, entry.Data, r)

	if err != nil {
		log.Printf("register finish: verify: %v", err)
		http.Error(w, "verification failed", http.StatusBadRequest)
		return
	}

	transports := make([]string, len(cred.Transport))
	for i, t := range cred.Transport {
		transports[i] = string(t)
	}

	p := &store.Passkey{
		UserID:         entry.User.ID,
		CredentialID:   cred.ID,
		PublicKey:      cred.PublicKey,
		SignCount:      cred.Authenticator.SignCount,
		Transports:     transports,
		BackupEligible: cred.Flags.BackupEligible,
		BackupState:    cred.Flags.BackupState,
	}

	err = a.Store.CreateUserWithPasskey(r.Context(), entry.User, p)

	if errors.Is(err, store.ErrUserExists) {
		http.Error(w, "user exists", http.StatusConflict)
		return
	}

	if err != nil {
		log.Printf("register finish: persist: %v", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user": map[string]any{
			"id":       entry.User.ID,
			"username": entry.User.Username,
		},
	})
}

type loginBeginReq struct {
	Email string `json:"email"`
}

// LoginBegin starts a WebAuthn login. Looks up the user by email
// and returns the assertion options for navigator.credentials.get.
// "User not found" and "user has no passkeys" both surface as a
// generic 401 so callers can't enumerate accounts.
func (a *Auth) LoginBegin(w http.ResponseWriter, r *http.Request) {
	var body loginBeginReq

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(body.Email)

	if email == "" {
		http.Error(w, "missing email", http.StatusBadRequest)
		return
	}

	user, err := a.Store.FindUserByEmail(r.Context(), email)

	if err != nil {
		log.Printf("login begin: find user: %v", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	passkeys, err := a.Store.FindPasskeysForUser(r.Context(), user.ID)

	if err != nil {
		log.Printf("login begin: find passkeys: %v", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	if len(passkeys) == 0 {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	wu := &webauthnUser{
		User:  user,
		Creds: passkeysToCredentials(passkeys),
	}

	options, session, err := a.WebAuthn.BeginLogin(wu)

	if err != nil {
		log.Printf("login begin: %v", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	sid := a.Sessions.Put(*session, user, sessionTTL)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(beginResp{
		SessionID: sid,
		Options:   options,
	})
}

// LoginFinish verifies the assertion the browser produced, bumps
// the matched passkey's sign_count + last_used_at, and returns the
// authenticated user.
func (a *Auth) LoginFinish(w http.ResponseWriter, r *http.Request) {
	sid := r.Header.Get("X-Pixeltown-Session")

	if sid == "" {
		http.Error(w, "missing session", http.StatusBadRequest)
		return
	}

	entry, ok := a.Sessions.Take(sid)

	if !ok {
		http.Error(w, "session expired", http.StatusUnauthorized)
		return
	}

	passkeys, err := a.Store.FindPasskeysForUser(r.Context(), entry.User.ID)

	if err != nil {
		log.Printf("login finish: find passkeys: %v", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	wu := &webauthnUser{
		User:  entry.User,
		Creds: passkeysToCredentials(passkeys),
	}

	cred, err := a.WebAuthn.FinishLogin(wu, entry.Data, r)

	if err != nil {
		log.Printf("login finish: verify: %v", err)
		http.Error(w, "verification failed", http.StatusUnauthorized)
		return
	}

	// CloneWarning fires when the authenticator's sign_count is at
	// or below what we have stored — strong signal of a cloned
	// credential. Reject the login.
	if cred.Authenticator.CloneWarning {
		log.Printf("login finish: clone warning for user %s", entry.User.ID)
		http.Error(w, "credential rejected", http.StatusUnauthorized)
		return
	}

	pk, err := a.Store.FindPasskeyByCredentialID(r.Context(), cred.ID)

	if err != nil {
		log.Printf("login finish: find passkey: %v", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	if pk == nil {
		log.Printf("login finish: passkey vanished")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	err = a.Store.UpdatePasskeyAfterLogin(
		r.Context(), pk.ID,
		cred.Authenticator.SignCount, cred.Flags.BackupState,
	)

	if err != nil {
		log.Printf("login finish: update passkey: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user": map[string]any{
			"id":       entry.User.ID,
			"username": entry.User.Username,
		},
	})
}
