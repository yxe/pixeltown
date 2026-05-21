package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"pixeltown/server/internal/auth"
	"pixeltown/server/internal/db"
	"pixeltown/server/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	port := os.Getenv("PORT")

	if port == "" {
		port = "4000"
	}

	d, err := db.Open(context.Background())

	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	defer d.Close()

	s := store.New(d.Pool)
	authH, err := auth.New(s)

	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "pixeltown server up")
	})

	http.HandleFunc("/api/auth/register/begin", authH.RegisterBegin)
	http.HandleFunc("/api/auth/register/finish", authH.RegisterFinish)

	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		var n int
		err := d.Pool.QueryRow(
			r.Context(),
			"SELECT count(*) FROM users",
		).Scan(&n)

		if err != nil {
			log.Printf("health: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    true,
			"users": n,
		})
	})

	log.Printf("listening on :%s", port)
	return http.ListenAndServe(":"+port, nil)
}
