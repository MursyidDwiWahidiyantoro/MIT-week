// Command server is AMBA's HTTP entrypoint.
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

	"github.com/amba/amba/backend/internal/config"
	"github.com/amba/amba/backend/internal/httpapi"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		// Boot errors name the variable, never its value.
		slog.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	handler, metrics := httpapi.NewRouterWithMetrics(cfg)
	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      cfg.HardTimeout + 15*time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// The metrics listener is separate and internal-only: it is never mounted
	// on the public router, and it stays off unless METRICS_ADDR is set.
	var metricsSrv *http.Server
	if cfg.MetricsAddr != "" {
		metricsSrv = &http.Server{
			Addr:              cfg.MetricsAddr,
			Handler:           metrics,
			ReadHeaderTimeout: 3 * time.Second,
		}
		go func() {
			slog.Info("metrics listening", "addr", cfg.MetricsAddr)
			if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("metrics listener failed", "err", err)
			}
		}()
	}

	go func() {
		slog.Info("listening",
			"addr", cfg.Addr(),
			"env", cfg.Environment,
			"llm", llmMode(cfg),
			"model", cfg.RouterModel,
			"retrieval", cfg.Retrieval && !cfg.LLMMock,
			"reasoning", cfg.Reasoning,
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown incomplete", "err", err)
	}
	if metricsSrv != nil {
		_ = metricsSrv.Shutdown(shutdownCtx)
	}
	slog.Info("stopped")
}

func llmMode(cfg *config.Config) string {
	if cfg.LLMMock {
		return "mock"
	}
	return "live"
}
