package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/config"
	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/database"
	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/handler"
	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/server"
	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/store"
)

const shutdownTimeout = 10 * time.Second

func main() {
	cfg := config.Load()

	db, err := database.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	dataStore := store.New(db)
	workspaceHandler := handler.NewWorkspaceHandler(dataStore)
	serviceHandler := handler.NewServiceHandler(dataStore)
	healthEventHandler := handler.NewHealthEventHandler(dataStore)
	businessJourneyHandler := handler.NewBusinessJourneyHandler(dataStore)
	impactAnalysisHandler := handler.NewImpactAnalysisHandler(dataStore)

	gin.SetMode(cfg.GinMode)

	router, err := server.New(
		db,
		workspaceHandler,
		serviceHandler,
		healthEventHandler,
		businessJourneyHandler,
		impactAnalysisHandler,
	)
	if err != nil {
		log.Fatalf("failed to configure server: %v", err)
	}

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	shutdownContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		log.Printf("OpsBoard API listening on port %s", cfg.Port)

		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	<-shutdownContext.Done()

	log.Println("shutting down OpsBoard API")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("OpsBoard API stopped")
}
