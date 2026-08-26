package main

import (
	"example.com/reproducible-build-farm/internal/application"
	"example.com/reproducible-build-farm/internal/cache"
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/internal/infrastructure"
	"example.com/reproducible-build-farm/internal/repository"
	"example.com/reproducible-build-farm/internal/transport"
	"log"
	"net/http"
	"os"
)

func main() {
	store := repository.NewMemory()
	_ = store.SaveWorker(nil, domain.Worker{ID: "sim-1", Platform: "linux/amd64", Version: "1", Capacity: domain.ResourceBudget{CPU: 8, MemoryMB: 8192}})
	app := application.New(store, cache.NewMemory(1000), infrastructure.NewSimulatedExecutor())
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("build farm listening on %s", addr)
	if err := http.ListenAndServe(addr, transport.New(app).Mux); err != nil {
		log.Fatal(err)
	}
}
