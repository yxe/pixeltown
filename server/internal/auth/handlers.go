package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const sessionTTL = 5 * time.Minute

type registerBeginReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type registerBeginResp struct {
	SessionID string `json:"sessionId"`
	Options any `json:"options"`
}

func (a *Auth) RegisterBegin(w http.ResponseWriter, r *http.Request) {
	var body registerBeginReq

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	_ = body
	_ = log.Printf
}

type registerFinishReq struct {
	SessionID  string          `json:"sessionId"`
	Credential json.RawMessage `json:"credential"`
}

func (a *Auth) RegisterFinish(w http.ResponseWriter, r *http.Request) {
	var body registerFinishReq

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	_ = body
}
