package main

import (
	"context"
	"hyperliquid-copy-trading/config"
	"hyperliquid-copy-trading/internal/database"
	"hyperliquid-copy-trading/internal/engine"
	"hyperliquid-copy-trading/internal/handlers"
	"hyperliquid-copy-trading/internal/websocket"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration
	cfg := config.Load()
	log.Info().Msg("Starting Hyperliquid Copy Trading Backend")

	// Initialize database
	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Initialize WebSocket manager
	wsManager := websocket.NewManager(cfg)

	// Initialize copy trading engine
	copyEngine := engine.NewCopyEngine(cfg, db, wsManager)

	// Initialize API handlers
	apiHandler := handlers.NewAPIHandler(copyEngine, db)

	// Setup routes
	router := mux.NewRouter()
	
	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/followers", apiHandler.CreateFollower).Methods("POST")
	api.HandleFunc("/followers", apiHandler.GetFollowers).Methods("GET")
	api.HandleFunc("/followers/{id}", apiHandler.UpdateFollower).Methods("PUT")
	api.HandleFunc("/followers/{id}", apiHandler.DeleteFollower).Methods("DELETE")
	api.HandleFunc("/leaders", apiHandler.GetLeaders).Methods("GET")
	api.HandleFunc("/leaders/{address}/performance", apiHandler.GetLeaderPerformance).Methods("GET")
	api.HandleFunc("/trades", apiHandler.GetTrades).Methods("GET")
	api.HandleFunc("/positions", apiHandler.GetPositions).Methods("GET")
	api.HandleFunc("/analytics/{follower_id}/pnl", apiHandler.GetPnLAnalytics).Methods("GET")
	
	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/")))

	// Start copy engine
	go copyEngine.Start()

	// Start HTTP server
	srv := &http.Server{
		Addr:         ":8000",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("Starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	copyEngine.Stop()
	wsManager.Close()
	srv.Shutdown(ctx)

	log.Info().Msg("Server shutdown complete")
}
