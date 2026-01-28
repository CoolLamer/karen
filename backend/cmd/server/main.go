package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/lukasbauer/karen/internal/app"
	"github.com/lukasbauer/karen/internal/httpapi"
)

func main() {
	cfg := app.LoadConfigFromEnv()

	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Initialize Sentry for error monitoring
	if cfg.SentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			EnableTracing:    true,
			TracesSampleRate: 0.2, // 20% of requests for performance monitoring
			Environment:      getEnvironment(),
		})
		if err != nil {
			logger.Printf("sentry init failed: %v", err)
		} else {
			logger.Printf("sentry initialized")
			defer sentry.Flush(2 * time.Second)
		}
	}

	a, err := app.New(cfg, logger)
	if err != nil {
		if cfg.SentryDSN != "" {
			sentry.CaptureException(err)
			sentry.Flush(2 * time.Second)
		}
		logger.Fatalf("init app: %v", err)
	}

	calls := httpapi.NewCallRegistry()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           a.Router(calls),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Printf("listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()

	// Phase 1: Start draining — reject new calls immediately
	calls.StartDraining()
	logger.Printf("shutdown: draining started, %d active call(s)", calls.ActiveCount())

	// Phase 2: Wait for active calls to finish (max 10 minutes)
	drainDone := make(chan struct{})
	go func() {
		calls.Wait()
		close(drainDone)
	}()

	select {
	case <-drainDone:
		logger.Printf("shutdown: all calls completed")
	case <-time.After(10 * time.Minute):
		logger.Printf("shutdown: drain timeout reached, %d call(s) still active", calls.ActiveCount())
	}

	// Phase 3: Shut down HTTP server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)

	// Phase 4: Close DB pool
	_ = a.Close()
}

func getEnvironment() string {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}
	return "development"
}
