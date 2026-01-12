package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lukasbauer/karen/internal/eventlog"
	"github.com/lukasbauer/karen/internal/httpapi"
	"github.com/lukasbauer/karen/internal/store"
)

type App struct {
	cfg      Config
	logger   *log.Logger
	db       *pgxpool.Pool
	store    *store.Store
	eventLog *eventlog.Logger
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

	// MVP: no automatic migrations to keep startup simple in Coolify.
	// Run migrations externally (psql) or extend later with a migration runner.

	return &App{
		cfg:      cfg,
		logger:   logger,
		db:       db,
		store:    s,
		eventLog: el,
	}, nil
}

func (a *App) Router() http.Handler {
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
		JWTSecret:             a.cfg.JWTSecret,
		JWTExpiry:             a.cfg.JWTExpiry,
		AdminPhones:           a.cfg.AdminPhones,
		DiscordWebhookURL:     a.cfg.DiscordWebhookURL,
	}
	return httpapi.NewRouter(routerCfg, a.logger, a.store, a.eventLog)
}

func (a *App) Close() error {
	if a.db != nil {
		a.db.Close()
	}
	return nil
}


