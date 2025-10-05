package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/temo927/go-msg-dispatcher/internal/app"
	"github.com/temo927/go-msg-dispatcher/internal/infra/cache"
	"github.com/temo927/go-msg-dispatcher/internal/infra/config"
	"github.com/temo927/go-msg-dispatcher/internal/infra/log"
	"github.com/temo927/go-msg-dispatcher/internal/infra/repository"
	"github.com/temo927/go-msg-dispatcher/internal/infra/webhook"
	httpapi "github.com/temo927/go-msg-dispatcher/internal/transport/http"
)

func main() {
	cfg := config.Load()

	db, err := repository.Connect(cfg.DBDSN)
	if err != nil {
		log.Logger.Error("failed to connect postgres", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Logger.Info("connected to postgres")

	cacheAdapter := cache.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.RedisTTL)

	provider := webhook.NewClient(webhook.Config{
		URL:          cfg.WebhookURL,
		AuthHeader:   cfg.WebhookAuthHeader,
		AuthValue:    cfg.WebhookAuthValue,
		AcceptAny2xx: cfg.AcceptAny2xx,
		Timeout:      5 * time.Second,
	})

	messageRepo := repository.NewMessagesRepo(db)
	sender := app.NewSender(
		messageRepo,
		provider,
		cacheAdapter,
		app.SenderConfig{MaxRetries: cfg.MaxRetries},
	)
	scheduler := app.NewScheduler(messageRepo, sender, cfg.TickInterval, cfg.BatchSize)

	handlers := httpapi.NewHandlers(scheduler, messageRepo)
	router := httpapi.NewRouter(handlers)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Logger.Info("server started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	if err := scheduler.Start(context.Background()); err != nil {
		log.Logger.Error("auto-start scheduler failed", "err", err)
	}

	<-ctx.Done()
	log.Logger.Info("shutdown initiated")

	_ = scheduler.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Logger.Error("graceful shutdown failed", "err", err)
	}

	log.Logger.Info("server stopped cleanly")
}
