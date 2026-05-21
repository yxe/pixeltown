package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"pixeltown/server/internal/db"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "4000"
	}

	d, err := db.Open(context.Background())

	if err != nil {
		log.Fatalf("db: %v", err)
	}

	defer d.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pixeltown server up")
	})

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
