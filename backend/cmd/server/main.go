package main

import (
	"log"

	"github.com/660710627/my-research/internal/db"
	"github.com/660710627/my-research/internal/handler"
	"github.com/660710627/my-research/internal/repo"
	"github.com/660710627/my-research/internal/service"
)

func main() {
	database, err := db.Open("library.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	healthRepository := repo.NewHealthRepository(database)
	healthService := service.NewHealthService(healthRepository)
	router := handler.NewRouter(handler.Dependencies{Health: healthService})

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
