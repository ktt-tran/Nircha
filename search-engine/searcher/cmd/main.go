package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"search-engine/searcher/config"
	fuzzysearch "search-engine/searcher/fuzzy-search"
	"search-engine/searcher/handlers"
	"search-engine/searcher/repository"
	"search-engine/searcher/server"
	"search-engine/searcher/services"

	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	repo := repository.NewRedisRepository()

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), cfg.StartupTimeout)
	defer cancelStartup()

	// Fail fast with a clear message if Redis isn't reachable, rather
	// than surfacing as a confusing downstream error the first time a
	// request comes in.
	if err := repo.Ping(startupCtx); err != nil {
		log.Fatalf("startup: cannot reach Redis at %s: %v", cfg.RedisAddress, err)
	}

	fuzzySearcher, err := fuzzysearch.NewFuzzySearcher(startupCtx, repo, cfg.FuzzyTolerance)
	if err != nil {
		log.Fatalf("startup: building fuzzy-search dictionary: %v", err)
	}

	searchService := services.NewSearchService(repo, fuzzySearcher)
	handler := handlers.NewHandler(searchService, fuzzySearcher, repo, cfg)
	srv := server.NewServer(handler, cfg)

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("server: listening on %s", srv.Addr())
		serverErr <- srv.Start()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			log.Fatalf("server: %v", err)
		}
	case sig := <-shutdown:
		log.Printf("server: received %s, shutting down gracefully", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("server: graceful shutdown failed: %v", err)
		}
		if err := repo.Close(); err != nil {
			log.Printf("server: closing redis connection: %v", err)
		}
	}
}
