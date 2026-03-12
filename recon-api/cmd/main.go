package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/umerthow/reconcile/recon-api/internal/config"
	"github.com/umerthow/reconcile/recon-api/internal/handlers"
	"github.com/umerthow/reconcile/recon-api/internal/middleware"
	"github.com/umerthow/reconcile/recon-api/internal/services"
	"github.com/umerthow/reconcile/recon-api/internal/utils"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	if cfg == nil {
		log.Fatal().Msg("Failed to load configuration")
	}

	// Initialize logger
	utils.InitLogger(cfg.Logger.Level)
	log.Info().Msg("Starting recon-api service")

	// Initialize services
	db, err := services.NewDatabaseService(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	redis, err := services.NewRedisService(&cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Redis")
	}
	defer redis.Close()

	kafka, err := services.NewKafkaService(&cfg.Kafka)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Kafka")
	}
	defer kafka.Close()

	log.Info().Msg("All services initialized successfully")

	// Initialize handlers
	webhookHandler := handlers.NewWebhookHandler(db, redis, kafka)
	reconHandler := handlers.NewReconciliationHandler(db)
	discrepancyHandler := handlers.NewDiscrepancyHandler(db)
	healthHandler := handlers.NewHealthHandler(db, redis)

	// Setup router
	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(middleware.Recovery())

	// Public routes
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// Webhook routes (no auth for provider webhooks)
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/xendit", webhookHandler.HandleXenditWebhook)
		webhooks.POST("/custody", webhookHandler.HandleCustodyWebhook)
	}

	// Protected API routes
	api := router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		// Reconciliation endpoints
		api.GET("/reconciliations", reconHandler.ListReconciliationRuns)
		api.POST("/reconciliations", reconHandler.TriggerReconciliation)

		// Discrepancy endpoints
		api.GET("/reconciliations/:run_id/discrepancies", discrepancyHandler.ListDiscrepancies)
		api.POST("/discrepancies/:discrepancy_id/resolve", discrepancyHandler.ResolveDiscrepancy)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("HTTP server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Print mock token for testing (POC only)
	if cfg.IsDevelopment() {
		token, _ := middleware.GenerateMockToken(&cfg.JWT, "dev@example.com")
		log.Info().Str("mock_token", token).Msg("Mock JWT token generated for testing")
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown with 10s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited gracefully")
}
