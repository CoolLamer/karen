package app

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lukasbauer/karen/internal/eventlog"
	"github.com/lukasbauer/karen/internal/httpapi"
	"github.com/lukasbauer/karen/internal/store"
)

type App struct {
	cfg        Config
	logger     *log.Logger
	db         *pgxpool.Pool
	store      *store.Store
	eventLog   *eventlog.Logger
	httpClient *http.Client // Shared HTTP client with connection pooling for TTS
}

func New(cfg Config, logger *log.Logger) (*App, error) {
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}

	s := store.New(db)
	el := eventlog.New(db)

	// Migrations are applied externally by the CI deploy job (docker exec psql).
	// No automatic migration runner at startup.

	// Shared HTTP client with connection pooling for TTS.
	// Keeps TCP connections alive to reduce latency for repeated TTS calls to ElevenLabs.
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10, // ElevenLabs is single host
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	return &App{
		cfg:        cfg,
		logger:     logger,
		db:         db,
		store:      s,
		eventLog:   el,
		httpClient: httpClient,
	}, nil
}

func (a *App) Router(calls *httpapi.CallRegistry) http.Handler {
	routerCfg := httpapi.RouterConfig{
		PublicBaseURL:         a.cfg.PublicBaseURL,
		TwilioAuthToken:       a.cfg.TwilioAuthTok,
		TwilioAccountSID:      a.cfg.TwilioAccountSID,
		TwilioVerifyServiceID: a.cfg.TwilioVerifyServiceID,
		DeepgramAPIKey:        a.cfg.DeepgramAPIKey,
		OpenAIAPIKey:          a.cfg.OpenAIAPIKey,
		ElevenLabsAPIKey:      a.cfg.ElevenLabsAPIKey,
		STTEndpointingMs:      a.cfg.STTEndpointingMs,
		STTUtteranceEndMs:     a.cfg.STTUtteranceEndMs,
		GreetingText:          a.cfg.GreetingText,
		TTSVoiceID:            a.cfg.TTSVoiceID,
		TTSStability:          a.cfg.TTSStability,
		TTSSimilarity:         a.cfg.TTSSimilarity,
		TTSHTTPClient:         a.httpClient,
		JWTSecret:             a.cfg.JWTSecret,
		JWTExpiry:             a.cfg.JWTExpiry,
		AdminPhones:           a.cfg.AdminPhones,
		DiscordWebhookURL:     a.cfg.DiscordWebhookURL,
		AIDebugAPIKey:         a.cfg.AIDebugAPIKey,
	}
	return httpapi.NewRouter(routerCfg, a.logger, a.store, a.eventLog, calls)
}

func (a *App) Close() error {
	if a.db != nil {
		a.db.Close()
	}
	return nil
}
