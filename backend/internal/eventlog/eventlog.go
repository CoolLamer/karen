package eventlog

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EventType represents the type of call event
type EventType string

const (
	EventCallStarted     EventType = "call_started"
	EventSTTResult       EventType = "stt_result"
	EventTurnFinalized   EventType = "turn_finalized"
	EventBargeIn         EventType = "barge_in"
	EventFillerSpoken    EventType = "filler_spoken"
	EventFillerSkipped   EventType = "filler_skipped"
	EventLLMStarted      EventType = "llm_started"
	EventLLMCompleted    EventType = "llm_completed"
	EventLLMError        EventType = "llm_error"
	EventGoodbyeDetected EventType = "goodbye_detected"
	EventForwardDetected EventType = "forward_detected"
	EventCallForwarded   EventType = "call_forwarded"
	EventCallHangup      EventType = "call_hangup"
	EventCallEnded       EventType = "call_ended"
)

// Logger provides async event logging to the database
type Logger struct {
	db *pgxpool.Pool
}

// New creates a new event logger
func New(db *pgxpool.Pool) *Logger {
	return &Logger{db: db}
}

// Log writes an event to the database synchronously
func (l *Logger) Log(ctx context.Context, callID string, eventType EventType, data map[string]any) error {
	if l.db == nil || callID == "" {
		return nil // Silently skip if no DB or call ID
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		dataJSON = []byte("{}")
	}

	_, err = l.db.Exec(ctx, `
		INSERT INTO call_events (call_id, event_type, event_data)
		VALUES ($1, $2, $3)
	`, callID, string(eventType), dataJSON)

	return err
}

// LogAsync logs an event without blocking the caller
func (l *Logger) LogAsync(callID string, eventType EventType, data map[string]any) {
	if l.db == nil || callID == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = l.Log(ctx, callID, eventType, data)
	}()
}
