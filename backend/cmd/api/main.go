package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/dig"

	"github.com/brandon/alcohol-label-verification-app/backend/internal/ai"
	"github.com/brandon/alcohol-label-verification-app/backend/internal/auth"
	"github.com/brandon/alcohol-label-verification-app/backend/internal/config"
	"github.com/brandon/alcohol-label-verification-app/backend/internal/httpapi"
	"github.com/brandon/alcohol-label-verification-app/backend/internal/rules"
	"github.com/brandon/alcohol-label-verification-app/backend/internal/store"
	"github.com/brandon/alcohol-label-verification-app/backend/internal/verification"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	container := dig.New()
	must(container.Provide(config.Load))
	must(container.Provide(func() *slog.Logger { return logger }))
	must(container.Provide(newDatabase))
	must(container.Provide(store.NewRepository))
	must(container.Provide(rules.NewEngine))
	must(container.Provide(newExtractor))
	must(container.Provide(newVerifier))
	must(container.Provide(func(
		extractor ai.Extractor,
		engine *rules.Engine,
		repo *store.Repository,
		cfg config.Config,
	) *verification.Service {
		return verification.NewService(extractor, engine, repo, cfg.BatchConcurrency)
	}))
	must(container.Provide(httpapi.NewHandler))
	must(container.Provide(httpapi.NewServer))

	err := container.Invoke(func(server *httpapi.Server, db *store.DB) error {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		errCh := make(chan error, 1)
		go func() {
			errCh <- server.ListenAndServe()
		}()

		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = server.Shutdown(shutdownCtx)
			_ = db.Close()
			return nil
		case err := <-errCh:
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			return err
		}
	})
	if err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func newDatabase(cfg config.Config) (*store.DB, error) {
	db, err := store.Open(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.Migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func newExtractor(cfg config.Config) ai.Extractor {
	switch cfg.AIProvider {
	case "fake":
		return ai.NewFakeExtractor()
	default:
		return ai.NewGeminiExtractor(cfg.GeminiAPIKey, cfg.GeminiModel, cfg.RequestTimeout)
	}
}

func newVerifier(cfg config.Config) *auth.Verifier {
	return auth.NewVerifier(cfg.NeonAuthJWKSURL, cfg.SkipAuthInDev)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
