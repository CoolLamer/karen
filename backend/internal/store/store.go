package store

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lukasbauer/karen/internal/costs"
)

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// stringOrDefault returns the string value or a default if nil
func stringOrDefault(s *string, def string) string {
	if s == nil {
		return def
	}
	return *s
}

// Tenant represents a customer/organization
type Tenant struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	SystemPrompt     string    `json:"system_prompt"`
	GreetingText     *string   `json:"greeting_text,omitempty"`
	VoiceID          *string   `json:"voice_id,omitempty"`
	Language         string    `json:"language"`
	VIPNames         []string  `json:"vip_names"`
	MarketingEmail   *string   `json:"marketing_email,omitempty"`
	ForwardNumber    *string   `json:"forward_number,omitempty"`
	MaxTurnTimeoutMs *int      `json:"max_turn_timeout_ms,omitempty"` // Hard timeout for speech_final in ms
	Plan             string    `json:"plan"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	// Billing fields
	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty"`
	CurrentPeriodCalls int        `json:"current_period_calls"`
}

// User represents an authenticated user
type User struct {
	ID            string     `json:"id"`
	TenantID      *string    `json:"tenant_id,omitempty"`
	Phone         string     `json:"phone"`
	PhoneVerified bool       `json:"phone_verified"`
	Name          *string    `json:"name,omitempty"`
	Role          string     `json:"role"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// TenantPhoneNumber represents a phone number assigned to a tenant
type TenantPhoneNumber struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	TwilioNumber     string    `json:"twilio_number"`
	TwilioSID        *string   `json:"twilio_sid,omitempty"`
	ForwardingSource *string   `json:"forwarding_source,omitempty"`
	IsPrimary        bool      `json:"is_primary"`
	CreatedAt        time.Time `json:"created_at"`
}

// UserSession represents a JWT session for logout/invalidation
type UserSession struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type Call struct {
	ID             string     `json:"id,omitempty"`
	TenantID       *string    `json:"tenant_id,omitempty"`
	Provider       string     `json:"provider"`
	ProviderCallID string     `json:"provider_call_id"`
	FromNumber     string     `json:"from_number"`
	ToNumber       string     `json:"to_number"`
	Status         string     `json:"status"`
	StartedAt      time.Time  `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at,omitempty"`
	EndedBy        *string    `json:"ended_by,omitempty"`
	FirstViewedAt  *time.Time `json:"first_viewed_at,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy     *string    `json:"resolved_by,omitempty"`
}

type ScreeningResult struct {
	LegitimacyLabel      string          `json:"legitimacy_label"`
	LegitimacyConfidence float64         `json:"legitimacy_confidence"`
	LeadLabel            string          `json:"lead_label"`
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

func (s *Store) UpdateCallEndedBy(ctx context.Context, providerCallID string, endedBy string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE calls
		SET ended_by = $1
		WHERE provider='twilio' AND provider_call_id=$2
	`, endedBy, providerCallID)
	return err
}

func (s *Store) ListCalls(ctx context.Context, limit int) ([]CallListItem, error) {
	rows, err := s.db.Query(ctx, `
		SELECT c.provider, c.provider_call_id, c.from_number, c.to_number, c.status, c.started_at, c.ended_at, c.ended_by,
		       r.legitimacy_label, r.legitimacy_confidence, r.lead_label, r.intent_category, r.intent_text, r.entities_json, r.created_at
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
		var leadLabel *string
		var intentCategory *string
		var intentText *string
		var entities []byte
		var screeningCreatedAt *time.Time

		err := rows.Scan(
			&item.Provider, &item.ProviderCallID, &item.FromNumber, &item.ToNumber, &item.Status, &item.StartedAt, &item.EndedAt, &item.EndedBy,
			&legitimacyLabel, &legitimacyConfidence, &leadLabel, &intentCategory, &intentText, &entities, &screeningCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if screeningCreatedAt != nil && legitimacyLabel != nil && legitimacyConfidence != nil && intentCategory != nil && intentText != nil {
			sr := ScreeningResult{
				LegitimacyLabel:      *legitimacyLabel,
				LegitimacyConfidence: *legitimacyConfidence,
				LeadLabel:            stringOrDefault(leadLabel, "nezjisteno"),
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
		INSERT INTO call_screening_results (call_id, legitimacy_label, legitimacy_confidence, lead_label, intent_category, intent_text, entities_json, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (call_id) DO UPDATE SET
			legitimacy_label = EXCLUDED.legitimacy_label,
			legitimacy_confidence = EXCLUDED.legitimacy_confidence,
			lead_label = EXCLUDED.lead_label,
			intent_category = EXCLUDED.intent_category,
			intent_text = EXCLUDED.intent_text,
			entities_json = EXCLUDED.entities_json,
			created_at = EXCLUDED.created_at
	`, callID, sr.LegitimacyLabel, sr.LegitimacyConfidence, sr.LeadLabel, sr.IntentCategory, sr.IntentText, sr.EntitiesJSON, sr.CreatedAt)
	return err
}

func (s *Store) GetCallDetail(ctx context.Context, providerCallID string) (CallDetail, error) {
	out, _, err := s.GetCallDetailWithTenantCheck(ctx, providerCallID)
	return out, err
}

// GetCallDetailWithTenantCheck retrieves call detail and also returns the tenant_id for verification.
func (s *Store) GetCallDetailWithTenantCheck(ctx context.Context, providerCallID string) (CallDetail, *string, error) {
	var out CallDetail
	var tenantID *string

	var callID string
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, provider, provider_call_id, from_number, to_number, status, started_at, ended_at, ended_by,
		       first_viewed_at, resolved_at, resolved_by
		FROM calls
		WHERE provider='twilio' AND provider_call_id=$1
	`, providerCallID).Scan(&callID, &tenantID, &out.Provider, &out.ProviderCallID, &out.FromNumber, &out.ToNumber, &out.Status, &out.StartedAt, &out.EndedAt, &out.EndedBy,
		&out.FirstViewedAt, &out.ResolvedAt, &out.ResolvedBy)
	if err != nil {
		return CallDetail{}, nil, err
	}
	out.ID = callID
	out.TenantID = tenantID

	// Screening result (optional)
	{
		var sr ScreeningResult
		var entities []byte
		err := s.db.QueryRow(ctx, `
			SELECT legitimacy_label, legitimacy_confidence, lead_label, intent_category, intent_text, entities_json, created_at
			FROM call_screening_results
			WHERE call_id=$1
		`, callID).Scan(&sr.LegitimacyLabel, &sr.LegitimacyConfidence, &sr.LeadLabel, &sr.IntentCategory, &sr.IntentText, &entities, &sr.CreatedAt)
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
		return out, tenantID, nil
	}
	defer rows.Close()

	for rows.Next() {
		var u Utterance
		if err := rows.Scan(&u.Speaker, &u.Text, &u.Sequence, &u.StartedAt, &u.EndedAt, &u.STTConfidence, &u.Interrupted); err != nil {
			return out, tenantID, nil
		}
		out.Utterances = append(out.Utterances, u)
	}

	return out, tenantID, nil
}

// ============================================================================
// Tenant operations
// ============================================================================

// GetTenantByTwilioNumber looks up a tenant by the Twilio phone number that received the call.
func (s *Store) GetTenantByTwilioNumber(ctx context.Context, twilioNumber string) (*Tenant, error) {
	var t Tenant
	err := s.db.QueryRow(ctx, `
		SELECT t.id, t.name, t.system_prompt, t.greeting_text, t.voice_id, t.language,
		       t.vip_names, t.marketing_email, t.forward_number, t.max_turn_timeout_ms,
		       t.plan, t.status, t.created_at, t.updated_at,
		       t.trial_ends_at, COALESCE(t.current_period_calls, 0)
		FROM tenants t
		JOIN tenant_phone_numbers pn ON pn.tenant_id = t.id
		WHERE pn.twilio_number = $1 AND t.status = 'active'
	`, twilioNumber).Scan(
		&t.ID, &t.Name, &t.SystemPrompt, &t.GreetingText, &t.VoiceID, &t.Language,
		&t.VIPNames, &t.MarketingEmail, &t.ForwardNumber, &t.MaxTurnTimeoutMs,
		&t.Plan, &t.Status, &t.CreatedAt, &t.UpdatedAt,
		&t.TrialEndsAt, &t.CurrentPeriodCalls,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetTenantByForwardingSource looks up a tenant by the forwarding source number (fallback).
func (s *Store) GetTenantByForwardingSource(ctx context.Context, forwardingSource string) (*Tenant, error) {
	var t Tenant
	err := s.db.QueryRow(ctx, `
		SELECT t.id, t.name, t.system_prompt, t.greeting_text, t.voice_id, t.language,
		       t.vip_names, t.marketing_email, t.forward_number, t.max_turn_timeout_ms,
		       t.plan, t.status, t.created_at, t.updated_at,
		       t.trial_ends_at, COALESCE(t.current_period_calls, 0)
		FROM tenants t
		JOIN tenant_phone_numbers pn ON pn.tenant_id = t.id
		WHERE pn.forwarding_source = $1 AND t.status = 'active'
	`, forwardingSource).Scan(
		&t.ID, &t.Name, &t.SystemPrompt, &t.GreetingText, &t.VoiceID, &t.Language,
		&t.VIPNames, &t.MarketingEmail, &t.ForwardNumber, &t.MaxTurnTimeoutMs,
		&t.Plan, &t.Status, &t.CreatedAt, &t.UpdatedAt,
		&t.TrialEndsAt, &t.CurrentPeriodCalls,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetTenantByID retrieves a tenant by ID.
func (s *Store) GetTenantByID(ctx context.Context, id string) (*Tenant, error) {
	var t Tenant
	err := s.db.QueryRow(ctx, `
		SELECT id, name, system_prompt, greeting_text, voice_id, language,
		       vip_names, marketing_email, forward_number, max_turn_timeout_ms,
		       plan, status, created_at, updated_at,
		       trial_ends_at, COALESCE(current_period_calls, 0)
		FROM tenants
		WHERE id = $1
	`, id).Scan(
		&t.ID, &t.Name, &t.SystemPrompt, &t.GreetingText, &t.VoiceID, &t.Language,
		&t.VIPNames, &t.MarketingEmail, &t.ForwardNumber, &t.MaxTurnTimeoutMs,
		&t.Plan, &t.Status, &t.CreatedAt, &t.UpdatedAt,
		&t.TrialEndsAt, &t.CurrentPeriodCalls,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateTenant creates a new tenant and returns it.
// New tenants are created on the trial plan with trial_ends_at set to 14 days from now.
func (s *Store) CreateTenant(ctx context.Context, name, systemPrompt, greetingText string) (*Tenant, error) {
	var t Tenant
	trialEndsAt := time.Now().AddDate(0, 0, 14) // 14 day trial
	err := s.db.QueryRow(ctx, `
		INSERT INTO tenants (name, system_prompt, greeting_text, trial_ends_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, system_prompt, greeting_text, voice_id, language,
		          vip_names, marketing_email, forward_number, max_turn_timeout_ms,
		          plan, status, created_at, updated_at, trial_ends_at, COALESCE(current_period_calls, 0)
	`, name, systemPrompt, greetingText, trialEndsAt).Scan(
		&t.ID, &t.Name, &t.SystemPrompt, &t.GreetingText, &t.VoiceID, &t.Language,
		&t.VIPNames, &t.MarketingEmail, &t.ForwardNumber, &t.MaxTurnTimeoutMs,
		&t.Plan, &t.Status, &t.CreatedAt, &t.UpdatedAt, &t.TrialEndsAt, &t.CurrentPeriodCalls,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateTenant updates a tenant's settings.
func (s *Store) UpdateTenant(ctx context.Context, id string, updates map[string]any) error {
	// Build dynamic UPDATE query
	// Note: In production, use a proper query builder
	_, err := s.db.Exec(ctx, `
		UPDATE tenants
		SET name = COALESCE($2, name),
		    system_prompt = COALESCE($3, system_prompt),
		    greeting_text = COALESCE($4, greeting_text),
		    voice_id = COALESCE($5, voice_id),
		    vip_names = COALESCE($6, vip_names),
		    marketing_email = COALESCE($7, marketing_email),
		    forward_number = COALESCE($8, forward_number),
		    max_turn_timeout_ms = COALESCE($9, max_turn_timeout_ms)
		WHERE id = $1
	`, id, updates["name"], updates["system_prompt"], updates["greeting_text"],
		updates["voice_id"], updates["vip_names"], updates["marketing_email"],
		updates["forward_number"], updates["max_turn_timeout_ms"])
	return err
}

// ============================================================================
// User operations
// ============================================================================

// GetUserByPhone retrieves a user by phone number.
func (s *Store) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	var u User
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, phone, phone_verified, name, role, last_login_at, created_at, updated_at
		FROM users
		WHERE phone = $1
	`, phone).Scan(
		&u.ID, &u.TenantID, &u.Phone, &u.PhoneVerified, &u.Name, &u.Role,
		&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByID retrieves a user by ID.
func (s *Store) GetUserByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, phone, phone_verified, name, role, last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&u.ID, &u.TenantID, &u.Phone, &u.PhoneVerified, &u.Name, &u.Role,
		&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindOrCreateUser finds a user by phone or creates a new one.
// If the user exists, marks phone as verified and updates last_login_at.
func (s *Store) FindOrCreateUser(ctx context.Context, phone string) (*User, bool, error) {
	var u User
	var isNew bool

	// Try to find existing user
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, phone, phone_verified, name, role, last_login_at, created_at, updated_at
		FROM users
		WHERE phone = $1
	`, phone).Scan(
		&u.ID, &u.TenantID, &u.Phone, &u.PhoneVerified, &u.Name, &u.Role,
		&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		// Create new user
		err = s.db.QueryRow(ctx, `
			INSERT INTO users (phone, phone_verified, last_login_at)
			VALUES ($1, true, NOW())
			RETURNING id, tenant_id, phone, phone_verified, name, role, last_login_at, created_at, updated_at
		`, phone).Scan(
			&u.ID, &u.TenantID, &u.Phone, &u.PhoneVerified, &u.Name, &u.Role,
			&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, false, err
		}
		isNew = true
	} else if err != nil {
		return nil, false, err
	} else {
		// Update existing user
		_, err = s.db.Exec(ctx, `
			UPDATE users
			SET phone_verified = true, last_login_at = NOW()
			WHERE id = $1
		`, u.ID)
		if err != nil {
			return nil, false, err
		}
		u.PhoneVerified = true
	}

	return &u, isNew, nil
}

// AssignUserToTenant assigns a user to a tenant.
func (s *Store) AssignUserToTenant(ctx context.Context, userID, tenantID string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET tenant_id = $2 WHERE id = $1
	`, userID, tenantID)
	return err
}

// ClearUserTenant removes the tenant association from a user.
// Used to fix orphaned tenant_id references when the tenant no longer exists.
func (s *Store) ClearUserTenant(ctx context.Context, userID string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET tenant_id = NULL WHERE id = $1
	`, userID)
	return err
}

// UpdateUserName updates a user's name.
func (s *Store) UpdateUserName(ctx context.Context, userID, name string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET name = $2 WHERE id = $1
	`, userID, name)
	return err
}

// GetTenantOwnerPhone retrieves the phone number of the tenant owner (for call forwarding).
func (s *Store) GetTenantOwnerPhone(ctx context.Context, tenantID string) (string, error) {
	var phone string
	err := s.db.QueryRow(ctx, `
		SELECT phone FROM users WHERE tenant_id = $1 AND role = 'owner' LIMIT 1
	`, tenantID).Scan(&phone)
	return phone, err
}

// ResetUserOnboarding performs a "smart reset" of a user's onboarding status.
// It clears the user's tenant_id and name (allowing re-onboarding) and releases
// any phone numbers associated with that tenant back to the pool.
// The tenant itself is preserved for call history purposes.
// Returns the user's previous tenant_id (if any) for logging purposes.
func (s *Store) ResetUserOnboarding(ctx context.Context, userID string) (*string, error) {
	// First get the user to find their tenant_id
	var previousTenantID *string
	err := s.db.QueryRow(ctx, `
		SELECT tenant_id FROM users WHERE id = $1
	`, userID).Scan(&previousTenantID)
	if err != nil {
		return nil, err
	}

	// If user has a tenant, release any phone numbers for that tenant back to pool
	if previousTenantID != nil {
		_, err = s.db.Exec(ctx, `
			UPDATE tenant_phone_numbers
			SET tenant_id = NULL, is_primary = false
			WHERE tenant_id = $1
		`, *previousTenantID)
		if err != nil {
			return previousTenantID, err
		}
	}

	// Clear the user's tenant_id and name (this makes them "not onboarded")
	_, err = s.db.Exec(ctx, `
		UPDATE users SET tenant_id = NULL, name = NULL WHERE id = $1
	`, userID)
	if err != nil {
		return previousTenantID, err
	}

	return previousTenantID, nil
}

// ============================================================================
// Phone number operations
// ============================================================================

// AssignPhoneNumberToTenant assigns a Twilio phone number to a tenant.
func (s *Store) AssignPhoneNumberToTenant(ctx context.Context, tenantID, twilioNumber, twilioSID string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO tenant_phone_numbers (tenant_id, twilio_number, twilio_sid)
		VALUES ($1, $2, $3)
		ON CONFLICT (twilio_number) DO UPDATE SET tenant_id = $1, twilio_sid = $3
	`, tenantID, twilioNumber, twilioSID)
	return err
}

// GetTenantPhoneNumbers retrieves all phone numbers for a tenant.
func (s *Store) GetTenantPhoneNumbers(ctx context.Context, tenantID string) ([]TenantPhoneNumber, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, tenant_id, twilio_number, twilio_sid, forwarding_source, is_primary, created_at
		FROM tenant_phone_numbers
		WHERE tenant_id = $1
		ORDER BY is_primary DESC, created_at ASC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var numbers []TenantPhoneNumber
	for rows.Next() {
		var pn TenantPhoneNumber
		if err := rows.Scan(&pn.ID, &pn.TenantID, &pn.TwilioNumber, &pn.TwilioSID,
			&pn.ForwardingSource, &pn.IsPrimary, &pn.CreatedAt); err != nil {
			return nil, err
		}
		numbers = append(numbers, pn)
	}
	return numbers, rows.Err()
}

// ClaimAvailablePhoneNumber atomically claims an available phone number from the pool
// and assigns it to the given tenant. Returns the claimed phone number or nil if none available.
func (s *Store) ClaimAvailablePhoneNumber(ctx context.Context, tenantID string) (*TenantPhoneNumber, error) {
	var pn TenantPhoneNumber
	err := s.db.QueryRow(ctx, `
		UPDATE tenant_phone_numbers
		SET tenant_id = $1, is_primary = true
		WHERE id = (
			SELECT id FROM tenant_phone_numbers
			WHERE tenant_id IS NULL
			ORDER BY created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, tenant_id, twilio_number, twilio_sid, forwarding_source, is_primary, created_at
	`, tenantID).Scan(&pn.ID, &pn.TenantID, &pn.TwilioNumber, &pn.TwilioSID,
		&pn.ForwardingSource, &pn.IsPrimary, &pn.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil // No available numbers
	}
	if err != nil {
		return nil, err
	}
	return &pn, nil
}

// ============================================================================
// Session operations
// ============================================================================

// CreateSession creates a new user session.
func (s *Store) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO user_sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

// RevokeSession revokes a session by token hash.
func (s *Store) RevokeSession(ctx context.Context, tokenHash string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE user_sessions SET revoked_at = NOW() WHERE token_hash = $1
	`, tokenHash)
	return err
}

// IsSessionValid checks if a session is valid (not revoked and not expired).
func (s *Store) IsSessionValid(ctx context.Context, tokenHash string) (bool, error) {
	var valid bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM user_sessions
			WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
		)
	`, tokenHash).Scan(&valid)
	return valid, err
}

// ============================================================================
// Call operations (tenant-aware)
// ============================================================================

// UpsertCallWithTenant creates or updates a call record with tenant ID.
func (s *Store) UpsertCallWithTenant(ctx context.Context, c Call) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO calls (id, tenant_id, provider, provider_call_id, from_number, to_number, status, started_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (provider, provider_call_id) DO UPDATE SET
			tenant_id = COALESCE(EXCLUDED.tenant_id, calls.tenant_id),
			from_number = EXCLUDED.from_number,
			to_number = EXCLUDED.to_number,
			status = EXCLUDED.status
	`, c.TenantID, c.Provider, c.ProviderCallID, c.FromNumber, c.ToNumber, c.Status, c.StartedAt)
	return err
}

// ListCallsByTenant lists calls for a specific tenant.
func (s *Store) ListCallsByTenant(ctx context.Context, tenantID string, limit int) ([]CallListItem, error) {
	rows, err := s.db.Query(ctx, `
		SELECT c.provider, c.provider_call_id, c.from_number, c.to_number, c.status, c.started_at, c.ended_at, c.ended_by,
		       c.first_viewed_at, c.resolved_at, c.resolved_by,
		       r.legitimacy_label, r.legitimacy_confidence, r.lead_label, r.intent_category, r.intent_text, r.entities_json, r.created_at
		FROM calls c
		LEFT JOIN call_screening_results r ON r.call_id = c.id
		WHERE c.tenant_id = $1
		ORDER BY c.started_at DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCallListItems(rows)
}

// scanCallListItems is a helper to scan call list rows.
func scanCallListItems(rows pgx.Rows) ([]CallListItem, error) {
	out := []CallListItem{}
	for rows.Next() {
		var item CallListItem
		var legitimacyLabel *string
		var legitimacyConfidence *float64
		var leadLabel *string
		var intentCategory *string
		var intentText *string
		var entities []byte
		var screeningCreatedAt *time.Time

		err := rows.Scan(
			&item.Provider, &item.ProviderCallID, &item.FromNumber, &item.ToNumber, &item.Status, &item.StartedAt, &item.EndedAt, &item.EndedBy,
			&item.FirstViewedAt, &item.ResolvedAt, &item.ResolvedBy,
			&legitimacyLabel, &legitimacyConfidence, &leadLabel, &intentCategory, &intentText, &entities, &screeningCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if screeningCreatedAt != nil && legitimacyLabel != nil && legitimacyConfidence != nil && intentCategory != nil && intentText != nil {
			sr := ScreeningResult{
				LegitimacyLabel:      *legitimacyLabel,
				LegitimacyConfidence: *legitimacyConfidence,
				LeadLabel:            stringOrDefault(leadLabel, "nezjisteno"),
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

// ============================================================================
// Admin operations
// ============================================================================

// AdminPhoneNumber includes tenant info for admin view.
type AdminPhoneNumber struct {
	ID               string    `json:"id"`
	TwilioNumber     string    `json:"twilio_number"`
	TwilioSID        *string   `json:"twilio_sid,omitempty"`
	ForwardingSource *string   `json:"forwarding_source,omitempty"`
	IsPrimary        bool      `json:"is_primary"`
	TenantID         *string   `json:"tenant_id,omitempty"`
	TenantName       *string   `json:"tenant_name,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// ListAllPhoneNumbers returns all phone numbers with tenant info (for admin view).
func (s *Store) ListAllPhoneNumbers(ctx context.Context) ([]AdminPhoneNumber, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			pn.id, pn.twilio_number, pn.twilio_sid, pn.forwarding_source,
			pn.is_primary, pn.tenant_id, t.name, pn.created_at
		FROM tenant_phone_numbers pn
		LEFT JOIN tenants t ON t.id = pn.tenant_id
		ORDER BY pn.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var numbers []AdminPhoneNumber
	for rows.Next() {
		var pn AdminPhoneNumber
		if err := rows.Scan(&pn.ID, &pn.TwilioNumber, &pn.TwilioSID, &pn.ForwardingSource,
			&pn.IsPrimary, &pn.TenantID, &pn.TenantName, &pn.CreatedAt); err != nil {
			return nil, err
		}
		numbers = append(numbers, pn)
	}
	return numbers, rows.Err()
}

// AddPhoneNumberToPool adds a new phone number to the available pool.
func (s *Store) AddPhoneNumberToPool(ctx context.Context, twilioNumber string, twilioSID *string) (*TenantPhoneNumber, error) {
	var pn TenantPhoneNumber
	err := s.db.QueryRow(ctx, `
		INSERT INTO tenant_phone_numbers (twilio_number, twilio_sid)
		VALUES ($1, $2)
		RETURNING id, COALESCE(tenant_id::text, ''), twilio_number, twilio_sid, forwarding_source, is_primary, created_at
	`, twilioNumber, twilioSID).Scan(&pn.ID, &pn.TenantID, &pn.TwilioNumber, &pn.TwilioSID,
		&pn.ForwardingSource, &pn.IsPrimary, &pn.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &pn, nil
}

// DeletePhoneNumber removes a phone number from the system.
func (s *Store) DeletePhoneNumber(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM tenant_phone_numbers WHERE id = $1`, id)
	return err
}

// DeleteTenant removes a tenant and all associated data from the system.
// This deletes the tenant along with all calls, users, and sessions.
// Phone numbers are unassigned (not deleted) so they return to the available pool.
func (s *Store) DeleteTenant(ctx context.Context, id string) error {
	// Use a transaction to ensure atomicity
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Delete all calls for this tenant first (not cascade-deleted)
	_, err = tx.Exec(ctx, `DELETE FROM calls WHERE tenant_id = $1`, id)
	if err != nil {
		return err
	}

	// Unassign phone numbers (set tenant_id to NULL) so they return to the pool
	// This must be done before deleting the tenant to avoid ON DELETE CASCADE
	_, err = tx.Exec(ctx, `UPDATE tenant_phone_numbers SET tenant_id = NULL WHERE tenant_id = $1`, id)
	if err != nil {
		return err
	}

	// Delete the tenant (cascades to users and sessions)
	result, err := tx.Exec(ctx, `DELETE FROM tenants WHERE id = $1`, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return tx.Commit(ctx)
}

// UpdatePhoneNumber updates a phone number's assignment and metadata.
// Pass nil for tenantID to unassign.
func (s *Store) UpdatePhoneNumber(ctx context.Context, id string, tenantID *string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE tenant_phone_numbers
		SET tenant_id = $2
		WHERE id = $1
	`, id, tenantID)
	return err
}

// AdminTenant is a simplified tenant view for admin lists.
type AdminTenant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AdminTenantDetail is a full tenant view for admin dashboard with counts.
type AdminTenantDetail struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	SystemPrompt     string    `json:"system_prompt"`
	GreetingText     *string   `json:"greeting_text,omitempty"`
	VoiceID          *string   `json:"voice_id,omitempty"`
	Language         string    `json:"language"`
	VIPNames         []string  `json:"vip_names"`
	MarketingEmail   *string   `json:"marketing_email,omitempty"`
	ForwardNumber    *string   `json:"forward_number,omitempty"`
	MaxTurnTimeoutMs *int      `json:"max_turn_timeout_ms,omitempty"`
	Plan             string    `json:"plan"`
	Status           string    `json:"status"`
	UserCount        int       `json:"user_count"`
	CallCount        int       `json:"call_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	// Billing fields
	StripeCustomerID     *string    `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty"`
	TrialEndsAt          *time.Time `json:"trial_ends_at,omitempty"`
	CurrentPeriodStart   *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodCalls   int        `json:"current_period_calls"`
	TimeSavedSeconds     int        `json:"time_saved_seconds"`
	SpamCallsBlocked     int        `json:"spam_calls_blocked"`
	// Admin-only fields
	AdminNotes *string `json:"admin_notes,omitempty"`
}

// AdminUser is a user view for admin dashboard.
type AdminUser struct {
	ID            string     `json:"id"`
	Phone         string     `json:"phone"`
	PhoneVerified bool       `json:"phone_verified"`
	Name          *string    `json:"name,omitempty"`
	Role          string     `json:"role"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ListAllTenants returns all tenants (for admin dropdowns).
func (s *Store) ListAllTenants(ctx context.Context) ([]AdminTenant, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name FROM tenants ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tenants := []AdminTenant{}
	for rows.Next() {
		var t AdminTenant
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

// ============================================================================
// Call event operations
// ============================================================================

// CallEvent represents a logged event for a call
type CallEvent struct {
	ID        string          `json:"id"`
	CallID    string          `json:"call_id"`
	EventType string          `json:"event_type"`
	EventData json.RawMessage `json:"event_data"`
	CreatedAt time.Time       `json:"created_at"`
}

// ListCallEvents retrieves events for a specific call
func (s *Store) ListCallEvents(ctx context.Context, callID string, limit int) ([]CallEvent, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, call_id, event_type, event_data, created_at
		FROM call_events
		WHERE call_id = $1
		ORDER BY created_at ASC
		LIMIT $2
	`, callID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []CallEvent
	for rows.Next() {
		var e CallEvent
		var eventData []byte
		if err := rows.Scan(&e.ID, &e.CallID, &e.EventType, &eventData, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.EventData = json.RawMessage(eventData)
		events = append(events, e)
	}
	return events, rows.Err()
}

// ============================================================================
// Admin users dashboard operations
// ============================================================================

// ListAllTenantsWithDetails returns all tenants with full details and counts for admin dashboard.
func (s *Store) ListAllTenantsWithDetails(ctx context.Context) ([]AdminTenantDetail, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			t.id, t.name, t.system_prompt, t.greeting_text, t.voice_id, t.language,
			t.vip_names, t.marketing_email, t.forward_number, t.max_turn_timeout_ms,
			t.plan, t.status, t.created_at, t.updated_at,
			COALESCE((SELECT COUNT(*) FROM users u WHERE u.tenant_id = t.id), 0) as user_count,
			COALESCE((SELECT COUNT(*) FROM calls c WHERE c.tenant_id = t.id), 0) as call_count,
			t.stripe_customer_id,
			t.stripe_subscription_id,
			t.trial_ends_at,
			t.current_period_start,
			COALESCE(t.current_period_calls, 0) as current_period_calls,
			COALESCE((SELECT SUM(time_saved_seconds) FROM tenant_usage tu WHERE tu.tenant_id = t.id), 0) as time_saved_seconds,
			COALESCE((SELECT SUM(spam_calls_blocked) FROM tenant_usage tu WHERE tu.tenant_id = t.id), 0) as spam_calls_blocked,
			t.admin_notes
		FROM tenants t
		ORDER BY t.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []AdminTenantDetail
	for rows.Next() {
		var t AdminTenantDetail
		if err := rows.Scan(
			&t.ID, &t.Name, &t.SystemPrompt, &t.GreetingText, &t.VoiceID, &t.Language,
			&t.VIPNames, &t.MarketingEmail, &t.ForwardNumber, &t.MaxTurnTimeoutMs,
			&t.Plan, &t.Status, &t.CreatedAt, &t.UpdatedAt, &t.UserCount, &t.CallCount,
			&t.StripeCustomerID, &t.StripeSubscriptionID,
			&t.TrialEndsAt, &t.CurrentPeriodStart, &t.CurrentPeriodCalls,
			&t.TimeSavedSeconds, &t.SpamCallsBlocked, &t.AdminNotes,
		); err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

// ListUsersByTenant returns all users for a specific tenant.
func (s *Store) ListUsersByTenant(ctx context.Context, tenantID string) ([]AdminUser, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, phone, phone_verified, name, role, last_login_at, created_at
		FROM users
		WHERE tenant_id = $1
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []AdminUser
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.Phone, &u.PhoneVerified, &u.Name, &u.Role, &u.LastLoginAt, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// ListCallsByTenantWithDetails returns calls with utterances for a tenant.
func (s *Store) ListCallsByTenantWithDetails(ctx context.Context, tenantID string, limit int) ([]CallDetail, error) {
	// First get the calls
	rows, err := s.db.Query(ctx, `
		SELECT c.id, c.tenant_id, c.provider, c.provider_call_id, c.from_number, c.to_number,
		       c.status, c.started_at, c.ended_at, c.ended_by,
		       r.legitimacy_label, r.legitimacy_confidence, r.lead_label, r.intent_category, r.intent_text, r.entities_json, r.created_at
		FROM calls c
		LEFT JOIN call_screening_results r ON r.call_id = c.id
		WHERE c.tenant_id = $1
		ORDER BY c.started_at DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calls []CallDetail
	var callIDs []string

	for rows.Next() {
		var cd CallDetail
		var callID string
		var legitimacyLabel *string
		var legitimacyConfidence *float64
		var leadLabel *string
		var intentCategory *string
		var intentText *string
		var entities []byte
		var screeningCreatedAt *time.Time

		err := rows.Scan(
			&callID, &cd.TenantID, &cd.Provider, &cd.ProviderCallID, &cd.FromNumber, &cd.ToNumber,
			&cd.Status, &cd.StartedAt, &cd.EndedAt, &cd.EndedBy,
			&legitimacyLabel, &legitimacyConfidence, &leadLabel, &intentCategory, &intentText, &entities, &screeningCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if screeningCreatedAt != nil && legitimacyLabel != nil && legitimacyConfidence != nil && intentCategory != nil && intentText != nil {
			sr := ScreeningResult{
				LegitimacyLabel:      *legitimacyLabel,
				LegitimacyConfidence: *legitimacyConfidence,
				LeadLabel:            stringOrDefault(leadLabel, "nezjisteno"),
				IntentCategory:       *intentCategory,
				IntentText:           *intentText,
				CreatedAt:            *screeningCreatedAt,
			}
			if len(entities) > 0 {
				sr.EntitiesJSON = json.RawMessage(entities)
			} else {
				sr.EntitiesJSON = json.RawMessage(`{}`)
			}
			cd.Screening = &sr
		}

		cd.ID = callID
		calls = append(calls, cd)
		callIDs = append(callIDs, callID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Build map after slice is complete to avoid stale pointers from slice reallocation
	callMap := make(map[string]*CallDetail, len(calls))
	for i := range calls {
		callMap[calls[i].ID] = &calls[i]
	}

	// Fetch utterances for all calls
	if len(callIDs) > 0 {
		utteranceRows, err := s.db.Query(ctx, `
			SELECT call_id, speaker, text, sequence, started_at, ended_at, stt_confidence, interrupted
			FROM call_utterances
			WHERE call_id = ANY($1)
			ORDER BY call_id, sequence ASC
		`, callIDs)
		if err != nil {
			// Log error but return calls without utterances rather than failing entirely
			return calls, nil
		}
		defer utteranceRows.Close()

		for utteranceRows.Next() {
			var callID string
			var u Utterance
			if err := utteranceRows.Scan(&callID, &u.Speaker, &u.Text, &u.Sequence, &u.StartedAt, &u.EndedAt, &u.STTConfidence, &u.Interrupted); err != nil {
				continue
			}
			if cd, ok := callMap[callID]; ok {
				cd.Utterances = append(cd.Utterances, u)
			}
		}
	}

	return calls, nil
}

// UpdateTenantPlanStatus updates only the plan and status fields of a tenant.
// Returns the number of rows affected (0 if tenant not found).
func (s *Store) UpdateTenantPlanStatus(ctx context.Context, tenantID, plan, status string) (int64, error) {
	result, err := s.db.Exec(ctx, `
		UPDATE tenants
		SET plan = $2, status = $3, updated_at = NOW()
		WHERE id = $1
	`, tenantID, plan, status)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// AdminUpdateTenantBilling updates billing-related fields and admin_notes for admin operations.
// Supports: trial_ends_at, current_period_calls, current_period_start, admin_notes
func (s *Store) AdminUpdateTenantBilling(ctx context.Context, tenantID string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE tenants SET updated_at = NOW()"
	args := []any{tenantID}
	argNum := 2

	if v, ok := updates["trial_ends_at"]; ok {
		query += fmt.Sprintf(", trial_ends_at = $%d", argNum)
		args = append(args, v)
		argNum++
	}
	if v, ok := updates["current_period_calls"]; ok {
		query += fmt.Sprintf(", current_period_calls = $%d", argNum)
		args = append(args, v)
		argNum++
	}
	if v, ok := updates["current_period_start"]; ok {
		query += fmt.Sprintf(", current_period_start = $%d", argNum)
		args = append(args, v)
		argNum++
	}
	if v, ok := updates["admin_notes"]; ok {
		query += fmt.Sprintf(", admin_notes = $%d", argNum)
		args = append(args, v)
	}

	query += " WHERE id = $1"
	_, err := s.db.Exec(ctx, query, args...)
	return err
}

// ============================================================================
// Call resolution tracking
// ============================================================================

// MarkCallViewed sets first_viewed_at if it's currently NULL.
// Returns true if the call was marked as viewed, false if already viewed.
func (s *Store) MarkCallViewed(ctx context.Context, providerCallID string) (bool, error) {
	result, err := s.db.Exec(ctx, `
		UPDATE calls
		SET first_viewed_at = NOW()
		WHERE provider = 'twilio' AND provider_call_id = $1 AND first_viewed_at IS NULL
	`, providerCallID)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

// MarkCallResolved sets resolved_at and resolved_by.
func (s *Store) MarkCallResolved(ctx context.Context, providerCallID string, userID string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE calls
		SET resolved_at = NOW(), resolved_by = $2
		WHERE provider = 'twilio' AND provider_call_id = $1
	`, providerCallID, userID)
	return err
}

// MarkCallUnresolved clears resolved_at and resolved_by.
func (s *Store) MarkCallUnresolved(ctx context.Context, providerCallID string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE calls
		SET resolved_at = NULL, resolved_by = NULL
		WHERE provider = 'twilio' AND provider_call_id = $1
	`, providerCallID)
	return err
}

// CountUnresolvedCalls returns the count of unresolved calls for a tenant.
// Unresolved means resolved_at IS NULL.
func (s *Store) CountUnresolvedCalls(ctx context.Context, tenantID string) (int, error) {
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM calls
		WHERE tenant_id = $1 AND resolved_at IS NULL
	`, tenantID).Scan(&count)
	return count, err
}

// GetCallTenantID retrieves the tenant_id for a call by provider_call_id.
func (s *Store) GetCallTenantID(ctx context.Context, providerCallID string) (*string, error) {
	var tenantID *string
	err := s.db.QueryRow(ctx, `
		SELECT tenant_id FROM calls
		WHERE provider = 'twilio' AND provider_call_id = $1
	`, providerCallID).Scan(&tenantID)
	return tenantID, err
}

// ============================================================================
// Usage tracking operations
// ============================================================================

// TenantUsage represents monthly usage for a tenant
type TenantUsage struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	CallsCount       int       `json:"calls_count"`
	MinutesUsed      int       `json:"minutes_used"`
	TimeSavedSeconds int       `json:"time_saved_seconds"`
	SpamCallsBlocked int       `json:"spam_calls_blocked"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TenantBillingInfo contains billing-related fields for a tenant
type TenantBillingInfo struct {
	TenantID             string     `json:"tenant_id"`
	Plan                 string     `json:"plan"`
	Status               string     `json:"status"`
	StripeCustomerID     *string    `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty"`
	TrialEndsAt          *time.Time `json:"trial_ends_at,omitempty"`
	CurrentPeriodStart   *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodCalls   int        `json:"current_period_calls"`
}

// IncrementTenantUsage increments usage counters for the current billing period.
// Creates a new usage record if one doesn't exist for the current month.
// isSpam indicates if the call was spam/marketing (for spam_calls_blocked counter).
// callDurationSeconds is the duration of the call for time_saved calculation.
func (s *Store) IncrementTenantUsage(ctx context.Context, tenantID string, callDurationSeconds int, isSpam bool) error {
	// Calculate time saved: call duration + overhead (5 min for spam, 2 min for others)
	overheadSeconds := 120 // 2 minutes
	if isSpam {
		overheadSeconds = 300 // 5 minutes
	}
	timeSaved := callDurationSeconds + overheadSeconds

	// Get current month boundaries
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, -1) // Last day of month

	spamIncrement := 0
	if isSpam {
		spamIncrement = 1
	}

	// Upsert the usage record for current month
	_, err := s.db.Exec(ctx, `
		INSERT INTO tenant_usage (tenant_id, period_start, period_end, calls_count, minutes_used, time_saved_seconds, spam_calls_blocked)
		VALUES ($1, $2, $3, 1, $4, $5, $6)
		ON CONFLICT (tenant_id, period_start) DO UPDATE SET
			calls_count = tenant_usage.calls_count + 1,
			minutes_used = tenant_usage.minutes_used + $4,
			time_saved_seconds = tenant_usage.time_saved_seconds + $5,
			spam_calls_blocked = tenant_usage.spam_calls_blocked + $6,
			updated_at = NOW()
	`, tenantID, periodStart, periodEnd, (callDurationSeconds+59)/60, timeSaved, spamIncrement)

	if err != nil {
		return err
	}

	// Also increment current_period_calls on tenant for quick limit checks
	_, err = s.db.Exec(ctx, `
		UPDATE tenants
		SET current_period_calls = current_period_calls + 1,
		    current_period_start = COALESCE(current_period_start, $2)
		WHERE id = $1
	`, tenantID, periodStart)

	return err
}

// GetTenantUsage retrieves usage for a specific period.
func (s *Store) GetTenantUsage(ctx context.Context, tenantID string, periodStart time.Time) (*TenantUsage, error) {
	var u TenantUsage
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, period_start, period_end, calls_count, minutes_used,
		       time_saved_seconds, spam_calls_blocked, created_at, updated_at
		FROM tenant_usage
		WHERE tenant_id = $1 AND period_start = $2
	`, tenantID, periodStart).Scan(
		&u.ID, &u.TenantID, &u.PeriodStart, &u.PeriodEnd, &u.CallsCount, &u.MinutesUsed,
		&u.TimeSavedSeconds, &u.SpamCallsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetTenantCurrentUsage retrieves usage for the current month.
func (s *Store) GetTenantCurrentUsage(ctx context.Context, tenantID string) (*TenantUsage, error) {
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	return s.GetTenantUsage(ctx, tenantID, periodStart)
}

// GetTenantTotalTimeSaved retrieves the total time saved across all periods.
func (s *Store) GetTenantTotalTimeSaved(ctx context.Context, tenantID string) (int, error) {
	var total int
	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(time_saved_seconds), 0)
		FROM tenant_usage
		WHERE tenant_id = $1
	`, tenantID).Scan(&total)
	return total, err
}

// GetTenantBillingInfo retrieves billing-related information for a tenant.
func (s *Store) GetTenantBillingInfo(ctx context.Context, tenantID string) (*TenantBillingInfo, error) {
	var b TenantBillingInfo
	err := s.db.QueryRow(ctx, `
		SELECT id, plan, status, stripe_customer_id, stripe_subscription_id,
		       trial_ends_at, current_period_start, COALESCE(current_period_calls, 0)
		FROM tenants
		WHERE id = $1
	`, tenantID).Scan(
		&b.TenantID, &b.Plan, &b.Status, &b.StripeCustomerID, &b.StripeSubscriptionID,
		&b.TrialEndsAt, &b.CurrentPeriodStart, &b.CurrentPeriodCalls,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// ResetTenantPeriodCalls resets the current period calls counter (called at billing cycle).
func (s *Store) ResetTenantPeriodCalls(ctx context.Context, tenantID string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE tenants
		SET current_period_calls = 0, current_period_start = $2
		WHERE id = $1
	`, tenantID, time.Now())
	return err
}

// GetPlanCallLimit returns the call limit for a given plan.
// Returns -1 for unlimited plans.
func GetPlanCallLimit(plan string) int {
	switch plan {
	case "trial":
		return 20
	case "basic":
		return 50
	case "pro":
		return -1 // unlimited
	default:
		return 0 // unknown plan, no calls allowed
	}
}

// TenantCallStatus represents the result of checking if a tenant can receive calls.
type TenantCallStatus struct {
	CanReceive     bool   `json:"can_receive"`
	Reason         string `json:"reason,omitempty"` // "ok", "trial_expired", "limit_exceeded"
	CallsUsed      int    `json:"calls_used"`       // Current period calls
	CallsLimit     int    `json:"calls_limit"`      // Plan limit (-1 = unlimited)
	TrialDaysLeft  int    `json:"trial_days_left,omitempty"`
	TrialCallsLeft int    `json:"trial_calls_left,omitempty"`
}

// CanTenantReceiveCalls checks if a tenant can receive calls based on their plan and usage.
// Returns detailed status including remaining calls/days for trials.
func CanTenantReceiveCalls(tenant *Tenant) TenantCallStatus {
	now := time.Now()
	limit := GetPlanCallLimit(tenant.Plan)

	status := TenantCallStatus{
		CallsUsed:  tenant.CurrentPeriodCalls,
		CallsLimit: limit,
	}

	// Check for cancelled or suspended status
	if tenant.Status == "cancelled" {
		status.CanReceive = false
		status.Reason = "subscription_cancelled"
		return status
	}
	if tenant.Status == "suspended" {
		status.CanReceive = false
		status.Reason = "subscription_suspended"
		return status
	}

	// Check for trial plan
	if tenant.Plan == "trial" {
		// Check trial expiration by date
		if tenant.TrialEndsAt != nil && now.After(*tenant.TrialEndsAt) {
			status.CanReceive = false
			status.Reason = "trial_expired"
			status.TrialDaysLeft = 0
			status.TrialCallsLeft = 0
			return status
		}

		// Calculate days left (use ceiling so partial days count as 1)
		if tenant.TrialEndsAt != nil {
			hoursLeft := tenant.TrialEndsAt.Sub(now).Hours()
			daysLeft := int(math.Ceil(hoursLeft / 24))
			if daysLeft < 0 {
				daysLeft = 0
			}
			status.TrialDaysLeft = daysLeft
		}

		// Check trial expiration by call count
		if tenant.CurrentPeriodCalls >= limit {
			status.CanReceive = false
			status.Reason = "limit_exceeded"
			status.TrialCallsLeft = 0
			return status
		}

		status.TrialCallsLeft = limit - tenant.CurrentPeriodCalls
		status.CanReceive = true
		status.Reason = "ok"
		return status
	}

	// For paid plans with limits (basic)
	if limit > 0 && tenant.CurrentPeriodCalls >= limit {
		status.CanReceive = false
		status.Reason = "limit_exceeded"
		return status
	}

	// Unlimited or within limits
	status.CanReceive = true
	status.Reason = "ok"
	return status
}

// UpdateTenantBilling updates billing-related fields on a tenant.
func (s *Store) UpdateTenantBilling(ctx context.Context, tenantID string, updates map[string]any) error {
	// Build dynamic UPDATE query for billing fields
	query := "UPDATE tenants SET updated_at = NOW()"
	args := []any{tenantID}
	argNum := 2

	if v, ok := updates["stripe_customer_id"]; ok {
		query += fmt.Sprintf(", stripe_customer_id = $%d", argNum)
		args = append(args, v)
		argNum++
	}
	if v, ok := updates["stripe_subscription_id"]; ok {
		query += fmt.Sprintf(", stripe_subscription_id = $%d", argNum)
		args = append(args, v)
		argNum++
	}
	if v, ok := updates["plan"]; ok {
		query += fmt.Sprintf(", plan = $%d", argNum)
		args = append(args, v)
		argNum++
	}
	if v, ok := updates["status"]; ok {
		query += fmt.Sprintf(", status = $%d", argNum)
		args = append(args, v)
		argNum++
	}
	if v, ok := updates["trial_ends_at"]; ok {
		query += fmt.Sprintf(", trial_ends_at = $%d", argNum)
		args = append(args, v)
		argNum++
	}
	if v, ok := updates["current_period_start"]; ok {
		query += fmt.Sprintf(", current_period_start = $%d", argNum)
		args = append(args, v)
		argNum++
	}
	if v, ok := updates["current_period_calls"]; ok {
		query += fmt.Sprintf(", current_period_calls = $%d", argNum)
		args = append(args, v)
		// Note: argNum not incremented since this is the last field
	}

	query += " WHERE id = $1"

	_, err := s.db.Exec(ctx, query, args...)
	return err
}

// GetTenantIDByStripeCustomer finds a tenant by their Stripe customer ID.
func (s *Store) GetTenantIDByStripeCustomer(ctx context.Context, stripeCustomerID string) (string, error) {
	var tenantID string
	err := s.db.QueryRow(ctx, `
		SELECT id FROM tenants WHERE stripe_customer_id = $1
	`, stripeCustomerID).Scan(&tenantID)
	if err != nil {
		return "", err
	}
	return tenantID, nil
}

// CallCostMetrics contains the raw metrics for a call used to calculate costs.
type CallCostMetrics struct {
	CallDurationSeconds int
	STTDurationSeconds  int
	LLMInputTokens      int
	LLMOutputTokens     int
	TTSCharacters       int
}

// CallCosts contains the calculated costs for a call in cents.
type CallCosts struct {
	CallID              string    `json:"call_id"`
	TwilioCostCents     int       `json:"twilio_cost_cents"`
	STTCostCents        int       `json:"stt_cost_cents"`
	LLMCostCents        int       `json:"llm_cost_cents"`
	TTSCostCents        int       `json:"tts_cost_cents"`
	TotalCostCents      int       `json:"total_cost_cents"`
	CallDurationSeconds int       `json:"call_duration_seconds"`
	STTDurationSeconds  int       `json:"stt_duration_seconds"`
	LLMInputTokens      int       `json:"llm_input_tokens"`
	LLMOutputTokens     int       `json:"llm_output_tokens"`
	TTSCharacters       int       `json:"tts_characters"`
	CreatedAt           time.Time `json:"created_at"`
}

// TenantCostSummary contains aggregated cost data for a tenant over a period.
type TenantCostSummary struct {
	TenantID             string `json:"tenant_id"`
	Period               string `json:"period"` // YYYY-MM format
	CallCount            int    `json:"call_count"`
	TotalDurationSeconds int    `json:"total_duration_seconds"`
	TwilioCostCents      int    `json:"twilio_cost_cents"`
	STTCostCents         int    `json:"stt_cost_cents"`
	LLMCostCents         int    `json:"llm_cost_cents"`
	TTSCostCents         int    `json:"tts_cost_cents"`
	TotalAPICostCents    int    `json:"total_api_cost_cents"`
	PhoneNumberCount     int    `json:"phone_number_count"`
	PhoneRentalCents     int    `json:"phone_rental_cents"`
	TotalCostCents       int    `json:"total_cost_cents"`
	// Raw metrics for debugging
	TotalSTTSeconds      int `json:"total_stt_seconds"`
	TotalLLMInputTokens  int `json:"total_llm_input_tokens"`
	TotalLLMOutputTokens int `json:"total_llm_output_tokens"`
	TotalTTSCharacters   int `json:"total_tts_characters"`
}

// RecordCallCosts saves the cost metrics for a call.
func (s *Store) RecordCallCosts(ctx context.Context, callID string, metrics CallCostMetrics, costs CallCosts) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO call_costs (
			call_id, twilio_cost_cents, stt_cost_cents, llm_cost_cents, tts_cost_cents,
			total_cost_cents, call_duration_seconds, stt_duration_seconds,
			llm_input_tokens, llm_output_tokens, tts_characters
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (call_id) DO UPDATE SET
			twilio_cost_cents = $2, stt_cost_cents = $3, llm_cost_cents = $4,
			tts_cost_cents = $5, total_cost_cents = $6, call_duration_seconds = $7,
			stt_duration_seconds = $8, llm_input_tokens = $9, llm_output_tokens = $10,
			tts_characters = $11
	`, callID, costs.TwilioCostCents, costs.STTCostCents, costs.LLMCostCents,
		costs.TTSCostCents, costs.TotalCostCents, metrics.CallDurationSeconds,
		metrics.STTDurationSeconds, metrics.LLMInputTokens, metrics.LLMOutputTokens,
		metrics.TTSCharacters)
	return err
}

// GetCallCosts retrieves the costs for a specific call.
func (s *Store) GetCallCosts(ctx context.Context, callID string) (*CallCosts, error) {
	var c CallCosts
	err := s.db.QueryRow(ctx, `
		SELECT call_id, twilio_cost_cents, stt_cost_cents, llm_cost_cents, tts_cost_cents,
		       total_cost_cents, COALESCE(call_duration_seconds, 0), COALESCE(stt_duration_seconds, 0),
		       COALESCE(llm_input_tokens, 0), COALESCE(llm_output_tokens, 0),
		       COALESCE(tts_characters, 0), created_at
		FROM call_costs WHERE call_id = $1
	`, callID).Scan(
		&c.CallID, &c.TwilioCostCents, &c.STTCostCents, &c.LLMCostCents, &c.TTSCostCents,
		&c.TotalCostCents, &c.CallDurationSeconds, &c.STTDurationSeconds,
		&c.LLMInputTokens, &c.LLMOutputTokens, &c.TTSCharacters, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetTenantCostSummary retrieves aggregated costs for a tenant for a specific month.
// period should be in YYYY-MM format.
func (s *Store) GetTenantCostSummary(ctx context.Context, tenantID string, period string) (*TenantCostSummary, error) {
	summary := &TenantCostSummary{
		TenantID: tenantID,
		Period:   period,
	}

	// Parse period to get date range (uses index on created_at instead of TO_CHAR)
	periodStart, err := time.Parse("2006-01", period)
	if err != nil {
		return nil, fmt.Errorf("invalid period format: %w", err)
	}
	periodEnd := periodStart.AddDate(0, 1, 0) // First day of next month

	// Get aggregated API costs from call_costs joined with calls
	err = s.db.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(cc.call_duration_seconds), 0),
			COALESCE(SUM(cc.twilio_cost_cents), 0),
			COALESCE(SUM(cc.stt_cost_cents), 0),
			COALESCE(SUM(cc.llm_cost_cents), 0),
			COALESCE(SUM(cc.tts_cost_cents), 0),
			COALESCE(SUM(cc.total_cost_cents), 0),
			COALESCE(SUM(cc.stt_duration_seconds), 0),
			COALESCE(SUM(cc.llm_input_tokens), 0),
			COALESCE(SUM(cc.llm_output_tokens), 0),
			COALESCE(SUM(cc.tts_characters), 0)
		FROM call_costs cc
		JOIN calls c ON c.id = cc.call_id
		WHERE c.tenant_id = $1
		  AND cc.created_at >= $2
		  AND cc.created_at < $3
	`, tenantID, periodStart, periodEnd).Scan(
		&summary.CallCount,
		&summary.TotalDurationSeconds,
		&summary.TwilioCostCents,
		&summary.STTCostCents,
		&summary.LLMCostCents,
		&summary.TTSCostCents,
		&summary.TotalAPICostCents,
		&summary.TotalSTTSeconds,
		&summary.TotalLLMInputTokens,
		&summary.TotalLLMOutputTokens,
		&summary.TotalTTSCharacters,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	// Get phone number count for this tenant
	var phoneCount int
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM tenant_phone_numbers WHERE tenant_id = $1
	`, tenantID).Scan(&phoneCount)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	summary.PhoneNumberCount = phoneCount

	// Calculate phone rental using centralized pricing constant
	summary.PhoneRentalCents = costs.CalculateMonthlyPhoneRentalCost(phoneCount)

	// Calculate total
	summary.TotalCostCents = summary.TotalAPICostCents + summary.PhoneRentalCents

	return summary, nil
}

// GetTenantPhoneNumberCount returns the count of phone numbers assigned to a tenant.
func (s *Store) GetTenantPhoneNumberCount(ctx context.Context, tenantID string) (int, error) {
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM tenant_phone_numbers WHERE tenant_id = $1
	`, tenantID).Scan(&count)
	return count, err
}
