package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/brandon/alcohol-label-verification-app/backend/internal/auth"
	"github.com/brandon/alcohol-label-verification-app/backend/internal/config"
)

type Server struct {
	cfg        config.Config
	handler    *Handler
	verifier   *auth.Verifier
	logger     *slog.Logger
	httpServer *http.Server
}

func NewServer(
	cfg config.Config,
	handler *Handler,
	verifier *auth.Verifier,
	logger *slog.Logger,
) *Server {
	mux := http.NewServeMux()
	s := &Server{cfg: cfg, handler: handler, verifier: verifier, logger: logger}

	cors := CORSMiddleware(cfg.AllowedOrigins, cfg.SkipAuthInDev)
	authMiddleware := AuthMiddleware(verifier)

	wrapProtected := func(fn http.HandlerFunc) http.Handler {
		return authMiddleware(http.HandlerFunc(fn))
	}

	mux.Handle("GET /healthz", http.HandlerFunc(s.handler.Health))
	mux.Handle("GET /api/v1/me", wrapProtected(s.handler.Me))
	mux.Handle("POST /api/v1/verifications", wrapProtected(s.handler.CreateVerification))
	mux.Handle("POST /api/v1/batches", wrapProtected(s.handler.CreateBatch))
	mux.Handle("GET /api/v1/verifications", wrapProtected(s.handler.ListVerifications))
	mux.Handle("GET /api/v1/verifications/{id}", wrapProtected(s.handler.GetVerification))
	mux.Handle("GET /api/v1/batches/{id}", wrapProtected(s.handler.GetBatch))

	s.httpServer = &http.Server{
		Addr:              cfg.Addr,
		Handler:           cors(LoggingMiddleware(logger)(mux)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.RequestTimeout,
		WriteTimeout:      cfg.RequestTimeout,
	}

	return s
}

func (s *Server) ListenAndServe() error {
	s.logger.Info("starting api server", "addr", s.cfg.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
