package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// Tenant represents a customer/organization
type Tenant struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	SystemPrompt   string     `json:"system_prompt"`
	GreetingText   *string    `json:"greeting_text,omitempty"`
	VoiceID        *string    `json:"voice_id,omitempty"`
	Language       string     `json:"language"`
	VIPNames       []string   `json:"vip_names"`
	MarketingEmail *string    `json:"marketing_email,omitempty"`
	ForwardNumber  *string    `json:"forward_number,omitempty"`
	Plan           string     `json:"plan"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
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

// ============================================================================
// Tenant operations
// ============================================================================

// GetTenantByTwilioNumber looks up a tenant by the Twilio phone number that received the call.
func (s *Store) GetTenantByTwilioNumber(ctx context.Context, twilioNumber string) (*Tenant, error) {
	var t Tenant
	err := s.db.QueryRow(ctx, `
		SELECT t.id, t.name, t.system_prompt, t.greeting_text, t.voice_id, t.language,
		       t.vip_names, t.marketing_email, t.forward_number, t.plan, t.status,
		       t.created_at, t.updated_at
		FROM tenants t
		JOIN tenant_phone_numbers pn ON pn.tenant_id = t.id
		WHERE pn.twilio_number = $1 AND t.status = 'active'
	`, twilioNumber).Scan(
		&t.ID, &t.Name, &t.SystemPrompt, &t.GreetingText, &t.VoiceID, &t.Language,
		&t.VIPNames, &t.MarketingEmail, &t.ForwardNumber, &t.Plan, &t.Status,
		&t.CreatedAt, &t.UpdatedAt,
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
		       t.vip_names, t.marketing_email, t.forward_number, t.plan, t.status,
		       t.created_at, t.updated_at
		FROM tenants t
		JOIN tenant_phone_numbers pn ON pn.tenant_id = t.id
		WHERE pn.forwarding_source = $1 AND t.status = 'active'
	`, forwardingSource).Scan(
		&t.ID, &t.Name, &t.SystemPrompt, &t.GreetingText, &t.VoiceID, &t.Language,
		&t.VIPNames, &t.MarketingEmail, &t.ForwardNumber, &t.Plan, &t.Status,
		&t.CreatedAt, &t.UpdatedAt,
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
		       vip_names, marketing_email, forward_number, plan, status,
		       created_at, updated_at
		FROM tenants
		WHERE id = $1
	`, id).Scan(
		&t.ID, &t.Name, &t.SystemPrompt, &t.GreetingText, &t.VoiceID, &t.Language,
		&t.VIPNames, &t.MarketingEmail, &t.ForwardNumber, &t.Plan, &t.Status,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateTenant creates a new tenant and returns it.
func (s *Store) CreateTenant(ctx context.Context, name, systemPrompt string) (*Tenant, error) {
	var t Tenant
	err := s.db.QueryRow(ctx, `
		INSERT INTO tenants (name, system_prompt)
		VALUES ($1, $2)
		RETURNING id, name, system_prompt, greeting_text, voice_id, language,
		          vip_names, marketing_email, forward_number, plan, status,
		          created_at, updated_at
	`, name, systemPrompt).Scan(
		&t.ID, &t.Name, &t.SystemPrompt, &t.GreetingText, &t.VoiceID, &t.Language,
		&t.VIPNames, &t.MarketingEmail, &t.ForwardNumber, &t.Plan, &t.Status,
		&t.CreatedAt, &t.UpdatedAt,
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
		    forward_number = COALESCE($8, forward_number)
		WHERE id = $1
	`, id, updates["name"], updates["system_prompt"], updates["greeting_text"],
		updates["voice_id"], updates["vip_names"], updates["marketing_email"],
		updates["forward_number"])
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

// UpdateUserName updates a user's name.
func (s *Store) UpdateUserName(ctx context.Context, userID, name string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET name = $2 WHERE id = $1
	`, userID, name)
	return err
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
		SELECT c.provider, c.provider_call_id, c.from_number, c.to_number, c.status, c.started_at, c.ended_at,
		       r.legitimacy_label, r.legitimacy_confidence, r.intent_category, r.intent_text, r.entities_json, r.created_at
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


