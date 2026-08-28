package main

import (
	"context"
	"log"
	"path/filepath"

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
	contractStager := service.NewLocalContractStager(filepath.Join("storage", "contracts"))
	multipartResearchService := service.NewMultipartResearchService(researchRepository, contractStager)
	multipartResearchHandler := handler.NewMultipartResearchHandler(multipartResearchService)
	researchListService := service.NewResearchListService(researchRepository)
	researchDeleteService := service.NewResearchDeleteService(researchRepository)
	researchStatusService := service.NewResearchStatusService(researchRepository)
	researchProcessService := service.NewResearchProcessService(researchRepository)
	router := handler.NewRouter(handler.Dependencies{
		Health: healthService, MultipartResearch: multipartResearchHandler, ResearchList: researchListService, ResearchDelete: researchDeleteService, ResearchStatus: researchStatusService, ResearchProcess: researchProcessService,
	})

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
