package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"corpus-tap/internal/capture"
	"corpus-tap/internal/config"
	"corpus-tap/internal/proxy"
	"corpus-tap/internal/store"

	"github.com/google/uuid"
)

func main() {
	cfg := config.Load()
	if err := cfg.Valid(); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	var pg *store.Postgres
	var deploymentID uuid.UUID

	if cfg.DatabaseURL != "" {
		var err error
		pg, err = store.NewPostgres(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("postgres: %v", err)
		}
		defer pg.Close()
		deploymentID, err = pg.EnsureDeployment(ctx, cfg.NewAPIImage, cfg.TapImage)
		if err != nil {
			log.Fatalf("deployment row: %v", err)
		}
		cfg.DeploymentID = deploymentID.String()
		log.Printf("corpus-tap: deployment_id=%s", deploymentID)
	} else {
		log.Print("corpus-tap: CORPUS_TAP_DATABASE_URL unset — metadata-only via local files")
		if cfg.DeploymentID == "" {
			cfg.DeploymentID = "local-dev"
		}
	}

	if cfg.LocalDataDir != "" {
		if err := os.MkdirAll(cfg.LocalDataDir, 0o750); err != nil {
			log.Fatalf("local data dir: %v", err)
		}
	}

	recorder := capture.NewRecorder(cfg, pg, deploymentID)
	srv, err := proxy.New(cfg, recorder)
	if err != nil {
		log.Fatal(err)
	}

	httpSrv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 30 * time.Second,
	}

	go func() {
		log.Printf("corpus-tap: listen %s upstream %s proxy_only=%v local_data=%q",
			cfg.ListenAddr, cfg.Upstream, cfg.ProxyOnly, cfg.LocalDataDir)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}
