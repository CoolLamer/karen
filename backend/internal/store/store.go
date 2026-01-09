package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

type Call struct {
	Provider       string    `json:"provider"`
	ProviderCallID string    `json:"provider_call_id"`
	FromNumber     string    `json:"from_number"`
	ToNumber       string    `json:"to_number"`
	Status         string    `json:"status"`
	StartedAt      time.Time `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at,omitempty"`
}

type ScreeningResult struct {
	LegitimacyLabel      string          `json:"legitimacy_label"`
	LegitimacyConfidence float64         `json:"legitimacy_confidence"`
	IntentCategory       string          `json:"intent_category"`
	IntentText           string          `json:"intent_text"`
	EntitiesJSON         json.RawMessage `json:"entities_json"`
	CreatedAt            time.Time       `json:"created_at"`
}

type CallListItem struct {
	Call
	Screening *ScreeningResult `json:"screening,omitempty"`
}

type Utterance struct {
	Speaker       string     `json:"speaker"`
	Text          string     `json:"text"`
	Sequence      int        `json:"sequence"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	EndedAt       *time.Time `json:"ended_at,omitempty"`
	STTConfidence *float64   `json:"stt_confidence,omitempty"`
	Interrupted   bool       `json:"interrupted"`
}

type CallDetail struct {
	Call
	Screening  *ScreeningResult `json:"screening,omitempty"`
	Utterances []Utterance      `json:"utterances"`
}

func (s *Store) UpsertCall(ctx context.Context, c Call) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO calls (id, provider, provider_call_id, from_number, to_number, status, started_at)
		VALUES (gen_random_uuid(), $1,$2,$3,$4,$5,$6)
		ON CONFLICT (provider, provider_call_id) DO UPDATE SET
			from_number = EXCLUDED.from_number,
			to_number = EXCLUDED.to_number,
			status = EXCLUDED.status
	`, c.Provider, c.ProviderCallID, c.FromNumber, c.ToNumber, c.Status, c.StartedAt)
	return err
}

func (s *Store) UpdateCallStatus(ctx context.Context, providerCallID string, status string, at time.Time) error {
	var endedAt *time.Time
	if status == "completed" || status == "canceled" || status == "failed" || status == "busy" || status == "no-answer" {
		endedAt = &at
	}
	_, err := s.db.Exec(ctx, `
		UPDATE calls
		SET status = $1,
		    ended_at = COALESCE($2, ended_at)
		WHERE provider='twilio' AND provider_call_id=$3
	`, status, endedAt, providerCallID)
	return err
}

func (s *Store) ListCalls(ctx context.Context, limit int) ([]CallListItem, error) {
	rows, err := s.db.Query(ctx, `
		SELECT c.provider, c.provider_call_id, c.from_number, c.to_number, c.status, c.started_at, c.ended_at,
		       r.legitimacy_label, r.legitimacy_confidence, r.intent_category, r.intent_text, r.entities_json, r.created_at
		FROM calls c
		LEFT JOIN call_screening_results r ON r.call_id = c.id
		ORDER BY c.started_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CallListItem
	for rows.Next() {
		var item CallListItem
		var legitimacyLabel *string
		var legitimacyConfidence *float64
		var intentCategory *string
		var intentText *string
		var entities []byte
		var screeningCreatedAt *time.Time

		err := rows.Scan(
			&item.Provider, &item.ProviderCallID, &item.FromNumber, &item.ToNumber, &item.Status, &item.StartedAt, &item.EndedAt,
			&legitimacyLabel, &legitimacyConfidence, &intentCategory, &intentText, &entities, &screeningCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if screeningCreatedAt != nil && legitimacyLabel != nil && legitimacyConfidence != nil && intentCategory != nil && intentText != nil {
			sr := ScreeningResult{
				LegitimacyLabel:      *legitimacyLabel,
				LegitimacyConfidence: *legitimacyConfidence,
				IntentCategory:       *intentCategory,
				IntentText:           *intentText,
				CreatedAt:            *screeningCreatedAt,
			}
			if len(entities) > 0 {
				sr.EntitiesJSON = json.RawMessage(entities)
			} else {
				sr.EntitiesJSON = json.RawMessage(`{}`)
			}
			item.Screening = &sr
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// GetCallID retrieves the internal call ID for a provider call ID.
func (s *Store) GetCallID(ctx context.Context, providerCallID string) (string, error) {
	var callID string
	err := s.db.QueryRow(ctx, `
		SELECT id FROM calls WHERE provider='twilio' AND provider_call_id=$1
	`, providerCallID).Scan(&callID)
	return callID, err
}

// InsertUtterance inserts a new utterance for a call.
func (s *Store) InsertUtterance(ctx context.Context, callID string, u Utterance) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO call_utterances (id, call_id, speaker, text, sequence, started_at, ended_at, stt_confidence, interrupted)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8)
	`, callID, u.Speaker, u.Text, u.Sequence, u.StartedAt, u.EndedAt, u.STTConfidence, u.Interrupted)
	return err
}

// InsertScreeningResult inserts a screening result for a call.
func (s *Store) InsertScreeningResult(ctx context.Context, callID string, sr ScreeningResult) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO call_screening_results (call_id, legitimacy_label, legitimacy_confidence, intent_category, intent_text, entities_json, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (call_id) DO UPDATE SET
			legitimacy_label = EXCLUDED.legitimacy_label,
			legitimacy_confidence = EXCLUDED.legitimacy_confidence,
			intent_category = EXCLUDED.intent_category,
			intent_text = EXCLUDED.intent_text,
			entities_json = EXCLUDED.entities_json,
			created_at = EXCLUDED.created_at
	`, callID, sr.LegitimacyLabel, sr.LegitimacyConfidence, sr.IntentCategory, sr.IntentText, sr.EntitiesJSON, sr.CreatedAt)
	return err
}

func (s *Store) GetCallDetail(ctx context.Context, providerCallID string) (CallDetail, error) {
	var out CallDetail

	var callID string
	err := s.db.QueryRow(ctx, `
		SELECT id, provider, provider_call_id, from_number, to_number, status, started_at, ended_at
		FROM calls
		WHERE provider='twilio' AND provider_call_id=$1
	`, providerCallID).Scan(&callID, &out.Provider, &out.ProviderCallID, &out.FromNumber, &out.ToNumber, &out.Status, &out.StartedAt, &out.EndedAt)
	if err != nil {
		return CallDetail{}, err
	}

	// Screening result (optional)
	{
		var sr ScreeningResult
		var entities []byte
		err := s.db.QueryRow(ctx, `
			SELECT legitimacy_label, legitimacy_confidence, intent_category, intent_text, entities_json, created_at
			FROM call_screening_results
			WHERE call_id=$1
		`, callID).Scan(&sr.LegitimacyLabel, &sr.LegitimacyConfidence, &sr.IntentCategory, &sr.IntentText, &entities, &sr.CreatedAt)
		if err == nil {
			sr.EntitiesJSON = json.RawMessage(entities)
			out.Screening = &sr
		}
	}

	// Utterances (optional)
	rows, err := s.db.Query(ctx, `
		SELECT speaker, text, sequence, started_at, ended_at, stt_confidence, interrupted
		FROM call_utterances
		WHERE call_id=$1
		ORDER BY sequence ASC
	`, callID)
	if err != nil {
		return out, nil
	}
	defer rows.Close()

	for rows.Next() {
		var u Utterance
		if err := rows.Scan(&u.Speaker, &u.Text, &u.Sequence, &u.StartedAt, &u.EndedAt, &u.STTConfidence, &u.Interrupted); err != nil {
			return out, nil
		}
		out.Utterances = append(out.Utterances, u)
	}

	return out, nil
}


