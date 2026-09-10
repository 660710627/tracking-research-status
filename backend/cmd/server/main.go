package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/660710627/my-research/internal/db"
	"github.com/660710627/my-research/internal/handler"
	"github.com/660710627/my-research/internal/repo"
	"github.com/660710627/my-research/internal/service"
)

const defaultDatabasePath = "library.db"

func main() {
	databasePath := os.Getenv("RESEARCH_DB_PATH")
	if databasePath == "" {
		databasePath = defaultDatabasePath
	}

	database, err := db.Open(databasePath)
	if err != nil {
		log.Fatalf("initialize database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	healthRepository := repo.NewHealthRepository(database)
	healthService := service.NewHealthService(healthRepository)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler.NewRouter(healthService),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("server listening on http://127.0.0.1:8080")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve HTTP: %v", err)
	}
}
