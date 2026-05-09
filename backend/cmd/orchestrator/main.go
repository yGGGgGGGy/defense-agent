package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gjy20/defense-agent/backend/internal/agent"
	"github.com/gjy20/defense-agent/backend/internal/api"
	"github.com/gjy20/defense-agent/backend/internal/audit"
	"github.com/gjy20/defense-agent/backend/internal/comm"
	"github.com/gjy20/defense-agent/backend/internal/config"
	"github.com/gjy20/defense-agent/backend/internal/graphiti"
	"github.com/gjy20/defense-agent/backend/internal/memory"
	"github.com/gjy20/defense-agent/backend/internal/orchestrator"
	"github.com/gjy20/defense-agent/backend/internal/sse"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	cfg := config.Load()

	ctx := context.Background()

	// Postgres
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		log.Fatal().Err(err).Msg("postgresql connect failed")
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("postgresql ping failed")
	}
	log.Info().Str("host", cfg.DBHost).Msg("connected to postgresql")

	// NATS
	bus, _ := comm.NewBus(cfg.NatsURL)

	// Neo4j
	graphClient, err := graphiti.NewClient(
		"bolt://localhost:7687", "neo4j", "defense123", "neo4j",
	)
	if err != nil {
		log.Warn().Err(err).Msg("neo4j init warning")
	}
	defer graphClient.Close()

	// SSE Broker
	sseBroker := sse.NewBroker()

	// Core components
	agentRegistry := agent.DefaultRegistry()
	auditPool := audit.NewPool(pool)
	memStore := memory.NewStore(pool)

	orch := orchestrator.New(agentRegistry, auditPool, memStore, bus, graphClient, sseBroker, cfg.MaxInstances)

	// HTTP Server
	srv := api.NewServer(orch, sseBroker)
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      srv,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.ServerPort).Msg("API server starting")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info().Str("signal", sig.String()).Msg("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	orch.Shutdown()
	httpServer.Shutdown(shutdownCtx)
	if bus != nil {
		bus.Close()
	}
	log.Info().Msg("shutdown complete")
}
