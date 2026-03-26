package main

import (
	"log"
)

func main() {
	cfg := mustLoadConfig()
	store := newStore()
	analytics := newAnalyticsEngine(store)
	srv := newServer(cfg, store, analytics)

	log.Printf("go-url-shortener v%s starting on :%s", cfg.Version, cfg.Port)
	if err := srv.run(); err != nil {
		log.Fatal(err)
	}
}
// v0-0
// v5-1
// v9-0
// v13-1
// v20-0
// v26-0
// v31-1
