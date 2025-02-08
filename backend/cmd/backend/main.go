package main

import (
	containershandler "app-pinger/backend/internal/api/handlers/containers"
	"app-pinger/backend/internal/api/handlers/verifier"
	"app-pinger/backend/internal/api/utilapi"
	"app-pinger/backend/internal/config"
	"app-pinger/backend/internal/usecase"
	repo "app-pinger/backend/internal/usecase/repo/postgres"
	"app-pinger/pkg/loger"
	queue "app-pinger/pkg/queue"
	"context"
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.ConfigLoad()

	virifierCfg := config.LoadVerifierConfiger()

	log := loger.SetupLogger(cfg.LogLevel)

	log.Info("starting backend-server")

	db, err := sql.Open("postgres", cfg.StoragePath)
	if err != nil {
		log.Error("failed to connect db: ", slog.Any("error", err))
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Error("failed to check db connection", slog.Any("error", err))
	} else {
		log.Info("successful connect to db")
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})

	m, err := migrate.NewWithDatabaseInstance(
		"file:///migrations",
		"postgres", driver)
	if err != nil {
		log.Error("failed to create migrations", slog.Any("error", err))
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error("failed to use migrations", slog.Any("error", err))
	}

	rabbitMQ, err := queue.NewConnection(cfg.RabbitMQPath, cfg.RabbitMQ.Queue)
	if err != nil {
		log.Error("failed to create rabbitMQ connection", slog.Any("error", err))
	}
	defer rabbitMQ.Close()

	containers := repo.NewContainerRepo(db)

	containerUseCase := usecase.NewBackendService(containers)

	containerHandler := containershandler.NewContainersHandler(containerUseCase, *rabbitMQ)
	verifierHandler := verifier.NewVerifier(virifierCfg.Keys, virifierCfg.RateLimit, virifierCfg.RateTime)

	router := utilapi.NewRouter(log)

	// RabbitMQ обработчик
	go func() {
		containerHandler.ProcessQueue(log)
	}()

	router.Handle("/container/getall", verifierHandler.Verify, containerHandler.GetAll)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		srv.ListenAndServe()
	}()

	log.Info("backend-server started")
	log.Debug("server settings", slog.Any("Address", cfg.Addr), slog.Any("ReadTimeout", cfg.Timeout),
		slog.Any("WriteTimeout", cfg.Timeout), slog.Any("IdleTimeout", cfg.IdleTimeout))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", slog.Any("error", err))

		return
	}

	log.Info("server stopped")
}
