package main

import (
	"context"
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
	if err := db.Initialize(context.Background(), database); err != nil {
		log.Fatal(err)
	}

	healthRepository := repo.NewHealthRepository(database)
	healthService := service.NewHealthService(healthRepository)
	researchRepository := repo.NewResearchRepository(database)
	researchService := service.NewResearchService(researchRepository)
	researchListService := service.NewResearchListService(researchRepository)
	researchUpdateService := service.NewResearchUpdateService(researchRepository)
	researchDeleteService := service.NewResearchDeleteService(researchRepository)
	router := handler.NewRouter(handler.Dependencies{
		Health: healthService, Researches: researchService, ResearchList: researchListService,
		ResearchUpdate: researchUpdateService, ResearchDelete: researchDeleteService,
	})

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
