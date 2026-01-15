package store

import (
	"context"
	"time"
)

// DevicePushToken represents a push notification token for a device
type DevicePushToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	Platform  string    `json:"platform"` // "ios" or "android"
	CreatedAt time.Time `json:"created_at"`
}

// RegisterPushToken registers or updates a device push token for a user
func (s *Store) RegisterPushToken(ctx context.Context, userID, token, platform string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO device_push_tokens (user_id, token, platform)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, token) DO UPDATE SET
			platform = EXCLUDED.platform,
			created_at = NOW()
	`, userID, token, platform)
	return err
}

// UnregisterPushToken removes a device push token
func (s *Store) UnregisterPushToken(ctx context.Context, token string) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM device_push_tokens WHERE token = $1
	`, token)
	return err
}

// UnregisterUserPushTokens removes all push tokens for a user
func (s *Store) UnregisterUserPushTokens(ctx context.Context, userID string) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM device_push_tokens WHERE user_id = $1
	`, userID)
	return err
}

// GetUserPushTokens returns all push tokens for a user
func (s *Store) GetUserPushTokens(ctx context.Context, userID string) ([]DevicePushToken, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, token, platform, created_at
		FROM device_push_tokens
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []DevicePushToken
	for rows.Next() {
		var t DevicePushToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.Token, &t.Platform, &t.CreatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

// GetTenantPushTokens returns all push tokens for all users in a tenant
func (s *Store) GetTenantPushTokens(ctx context.Context, tenantID string) ([]DevicePushToken, error) {
	rows, err := s.db.Query(ctx, `
		SELECT dpt.id, dpt.user_id, dpt.token, dpt.platform, dpt.created_at
		FROM device_push_tokens dpt
		JOIN users u ON u.id = dpt.user_id
		WHERE u.tenant_id = $1
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []DevicePushToken
	for rows.Next() {
		var t DevicePushToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.Token, &t.Platform, &t.CreatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}
