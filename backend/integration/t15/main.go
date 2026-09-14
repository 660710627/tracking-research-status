// Manual E2E runner using the real handlers/services and an isolated database.
package main

import (
	"context"
	"github.com/660710627/my-research/internal/db"
	"github.com/660710627/my-research/internal/handler"
	"github.com/660710627/my-research/internal/repo"
	"github.com/660710627/my-research/internal/service"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) > 1 {
		if err := fixtureCommand(os.Args[1:]); err != nil {
			log.Fatal(err)
		}
		return
	}
	root, err := os.MkdirTemp("", "research-t15-")
	if err != nil {
		log.Fatal(err)
	}
	database, err := db.Open(filepath.Join(root, "test.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	if err = db.Migrate(context.Background(), database); err != nil {
		log.Fatal(err)
	}
	store, err := service.NewFileContractStore(filepath.Join(root, "files"))
	if err != nil {
		log.Fatal(err)
	}
	repository := repo.NewResearchRepository(database)
	log.Printf("T15 isolated data: %s", root)
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", handler.NewRouter(
		service.NewHealthService(repo.NewHealthRepository(database)),
		handler.WithResearchCreator(service.NewResearchService(repository, store)),
		handler.WithResearchLister(service.NewResearchListService(repository)))))
}
