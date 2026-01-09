package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lukasbauer/karen/internal/store"
)

// Context key for user data
type contextKey string

const userContextKey contextKey = "user"

// JWTClaims represents the claims in the JWT token
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID   string  `json:"user_id"`
	TenantID *string `json:"tenant_id,omitempty"`
	Phone    string  `json:"phone"`
}

// AuthUser represents the authenticated user in request context
type AuthUser struct {
	ID       string
	TenantID *string
	Phone    string
}

// E.164 phone number validation (international format)
var e164Regex = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

func isValidE164(phone string) bool {
	return e164Regex.MatchString(phone)
}

// hashToken creates a SHA256 hash of the token for storage
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// withAuth is middleware that requires valid JWT authentication
func (r *Router) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Get token from Authorization header
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error": "invalid authorization format"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Parse and validate JWT
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(r.cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"error": "invalid token"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			http.Error(w, `{"error": "invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		// Check if session is valid (not revoked)
		tokenHash := hashToken(tokenString)
		valid, err := r.store.IsSessionValid(req.Context(), tokenHash)
		if err != nil || !valid {
			http.Error(w, `{"error": "session expired or revoked"}`, http.StatusUnauthorized)
			return
		}

		// Add user to context
		user := &AuthUser{
			ID:       claims.UserID,
			TenantID: claims.TenantID,
			Phone:    claims.Phone,
		}
		ctx := context.WithValue(req.Context(), userContextKey, user)
		next.ServeHTTP(w, req.WithContext(ctx))
	}
}

// getAuthUser extracts the authenticated user from context
func getAuthUser(ctx context.Context) *AuthUser {
	user, _ := ctx.Value(userContextKey).(*AuthUser)
	return user
}

// generateJWT creates a new JWT token for a user
func (r *Router) generateJWT(user *store.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(r.cfg.JWTExpiry)

	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:   user.ID,
		TenantID: user.TenantID,
		Phone:    user.Phone,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(r.cfg.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// handleSendCode initiates phone verification via Twilio Verify
func (r *Router) handleSendCode(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Phone string `json:"phone"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate phone format
	if !isValidE164(body.Phone) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid phone format, use E.164 (e.g., +420777123456)",
		})
		return
	}

	// Check if Twilio Verify is configured
	if r.cfg.TwilioAccountSID == "" || r.cfg.TwilioVerifyServiceID == "" {
		r.logger.Printf("auth: Twilio Verify not configured")
		http.Error(w, `{"error": "SMS verification not configured"}`, http.StatusServiceUnavailable)
		return
	}

	// Call Twilio Verify API to send SMS
	err := r.sendTwilioVerifyCode(req.Context(), body.Phone)
	if err != nil {
		r.logger.Printf("auth: failed to send verification code to %s: %v", body.Phone, err)
		http.Error(w, `{"error": "failed to send verification code"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// handleVerifyCode verifies the OTP code and issues JWT
func (r *Router) handleVerifyCode(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate inputs
	if !isValidE164(body.Phone) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid phone format",
		})
		return
	}

	if len(body.Code) != 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "code must be 6 digits",
		})
		return
	}

	// Check if Twilio Verify is configured
	if r.cfg.TwilioAccountSID == "" || r.cfg.TwilioVerifyServiceID == "" {
		r.logger.Printf("auth: Twilio Verify not configured")
		http.Error(w, `{"error": "SMS verification not configured"}`, http.StatusServiceUnavailable)
		return
	}

	// Verify code with Twilio
	valid, err := r.verifyTwilioCode(req.Context(), body.Phone, body.Code)
	if err != nil {
		r.logger.Printf("auth: verification check failed for %s: %v", body.Phone, err)
		http.Error(w, `{"error": "verification failed"}`, http.StatusInternalServerError)
		return
	}

	if !valid {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "invalid or expired code",
		})
		return
	}

	// Find or create user
	user, isNew, err := r.store.FindOrCreateUser(req.Context(), body.Phone)
	if err != nil {
		r.logger.Printf("auth: failed to find/create user for %s: %v", body.Phone, err)
		http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
		return
	}

	// Generate JWT
	token, expiresAt, err := r.generateJWT(user)
	if err != nil {
		r.logger.Printf("auth: failed to generate JWT: %v", err)
		http.Error(w, `{"error": "failed to create session"}`, http.StatusInternalServerError)
		return
	}

	// Store session for logout/revocation
	tokenHash := hashToken(token)
	if err := r.store.CreateSession(req.Context(), user.ID, tokenHash, expiresAt); err != nil {
		r.logger.Printf("auth: failed to store session: %v", err)
		http.Error(w, `{"error": "failed to create session"}`, http.StatusInternalServerError)
		return
	}

	r.logger.Printf("auth: user %s logged in (new: %v)", body.Phone, isNew)

	writeJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"expires_at": expiresAt.Format(time.RFC3339),
		"user":       user,
		"is_new":     isNew,
	})
}

// handleRefreshToken issues a new JWT token
func (r *Router) handleRefreshToken(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Parse existing token (allow expired tokens for refresh)
	parser := jwt.NewParser(jwt.WithExpirationRequired())
	token, err := parser.ParseWithClaims(body.Token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.cfg.JWTSecret), nil
	})

	// Allow expired tokens (we're refreshing) but reject other errors
	if err != nil {
		if !errors.Is(err, jwt.ErrTokenExpired) {
			http.Error(w, `{"error": "invalid token"}`, http.StatusUnauthorized)
			return
		}
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		http.Error(w, `{"error": "invalid token claims"}`, http.StatusUnauthorized)
		return
	}

	// Check if old session is still valid (not revoked)
	oldTokenHash := hashToken(body.Token)
	valid, err := r.store.IsSessionValid(req.Context(), oldTokenHash)
	if err != nil || !valid {
		http.Error(w, `{"error": "session revoked"}`, http.StatusUnauthorized)
		return
	}

	// Get fresh user data
	user, err := r.store.GetUserByID(req.Context(), claims.UserID)
	if err != nil {
		http.Error(w, `{"error": "user not found"}`, http.StatusUnauthorized)
		return
	}

	// Generate new token
	newToken, expiresAt, err := r.generateJWT(user)
	if err != nil {
		http.Error(w, `{"error": "failed to create session"}`, http.StatusInternalServerError)
		return
	}

	// Revoke old session and create new one
	_ = r.store.RevokeSession(req.Context(), oldTokenHash)
	newTokenHash := hashToken(newToken)
	_ = r.store.CreateSession(req.Context(), user.ID, newTokenHash, expiresAt)

	writeJSON(w, http.StatusOK, map[string]any{
		"token":      newToken,
		"expires_at": expiresAt.Format(time.RFC3339),
		"user":       user,
	})
}

// handleLogout revokes the current session
func (r *Router) handleLogout(w http.ResponseWriter, req *http.Request) {
	// Get token from header
	authHeader := req.Header.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 {
		tokenHash := hashToken(parts[1])
		_ = r.store.RevokeSession(req.Context(), tokenHash)
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// handleGetMe returns the current user's data
func (r *Router) handleGetMe(w http.ResponseWriter, req *http.Request) {
	authUser := getAuthUser(req.Context())
	if authUser == nil {
		http.Error(w, `{"error": "not authenticated"}`, http.StatusUnauthorized)
		return
	}

	user, err := r.store.GetUserByID(req.Context(), authUser.ID)
	if err != nil {
		http.Error(w, `{"error": "user not found"}`, http.StatusNotFound)
		return
	}

	// Get tenant info if assigned
	var tenant *store.Tenant
	if user.TenantID != nil {
		tenant, _ = r.store.GetTenantByID(req.Context(), *user.TenantID)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":   user,
		"tenant": tenant,
	})
}

// handleGetTenant returns the current user's tenant
func (r *Router) handleGetTenant(w http.ResponseWriter, req *http.Request) {
	authUser := getAuthUser(req.Context())
	if authUser == nil || authUser.TenantID == nil {
		http.Error(w, `{"error": "no tenant assigned"}`, http.StatusNotFound)
		return
	}

	tenant, err := r.store.GetTenantByID(req.Context(), *authUser.TenantID)
	if err != nil {
		http.Error(w, `{"error": "tenant not found"}`, http.StatusNotFound)
		return
	}

	// Get phone numbers
	numbers, _ := r.store.GetTenantPhoneNumbers(req.Context(), tenant.ID)

	writeJSON(w, http.StatusOK, map[string]any{
		"tenant":        tenant,
		"phone_numbers": numbers,
	})
}

// allowedTenantUpdateFields are the fields users can update on their tenant
var allowedTenantUpdateFields = map[string]bool{
	"name":            true,
	"system_prompt":   true,
	"greeting_text":   true,
	"voice_id":        true,
	"vip_names":       true,
	"marketing_email": true,
	"forward_number":  true,
}

// handleUpdateTenant updates the current user's tenant settings
func (r *Router) handleUpdateTenant(w http.ResponseWriter, req *http.Request) {
	authUser := getAuthUser(req.Context())
	if authUser == nil || authUser.TenantID == nil {
		http.Error(w, `{"error": "no tenant assigned"}`, http.StatusNotFound)
		return
	}

	var rawUpdates map[string]any
	if err := json.NewDecoder(req.Body).Decode(&rawUpdates); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Filter to only allowed fields
	updates := make(map[string]any)
	for key, value := range rawUpdates {
		if allowedTenantUpdateFields[key] {
			updates[key] = value
		}
	}

	if len(updates) == 0 {
		http.Error(w, `{"error": "no valid fields to update"}`, http.StatusBadRequest)
		return
	}

	// Apply updates
	if err := r.store.UpdateTenant(req.Context(), *authUser.TenantID, updates); err != nil {
		r.logger.Printf("auth: failed to update tenant %s: %v", *authUser.TenantID, err)
		http.Error(w, `{"error": "failed to update tenant"}`, http.StatusInternalServerError)
		return
	}

	// Return updated tenant
	tenant, err := r.store.GetTenantByID(req.Context(), *authUser.TenantID)
	if err != nil {
		http.Error(w, `{"error": "tenant not found"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, tenant)
}

// handleCompleteOnboarding creates a tenant for a new user
func (r *Router) handleCompleteOnboarding(w http.ResponseWriter, req *http.Request) {
	authUser := getAuthUser(req.Context())
	if authUser == nil {
		http.Error(w, `{"error": "not authenticated"}`, http.StatusUnauthorized)
		return
	}

	// Check if already has tenant
	if authUser.TenantID != nil {
		http.Error(w, `{"error": "already onboarded"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		Name         string `json:"name"`          // User's name
		SystemPrompt string `json:"system_prompt"` // Custom prompt (optional)
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if body.Name == "" {
		http.Error(w, `{"error": "name is required"}`, http.StatusBadRequest)
		return
	}

	// Use default system prompt if not provided
	systemPrompt := body.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = generateDefaultSystemPrompt(body.Name)
	}

	// Create tenant
	tenant, err := r.store.CreateTenant(req.Context(), body.Name, systemPrompt)
	if err != nil {
		r.logger.Printf("auth: failed to create tenant: %v", err)
		http.Error(w, `{"error": "failed to create tenant"}`, http.StatusInternalServerError)
		return
	}

	// Update user name
	if err := r.store.UpdateUserName(req.Context(), authUser.ID, body.Name); err != nil {
		r.logger.Printf("auth: failed to update user name: %v", err)
	}

	// Assign user to tenant
	if err := r.store.AssignUserToTenant(req.Context(), authUser.ID, tenant.ID); err != nil {
		r.logger.Printf("auth: failed to assign user to tenant: %v", err)
		http.Error(w, `{"error": "failed to assign tenant"}`, http.StatusInternalServerError)
		return
	}

	// Get updated user
	user, _ := r.store.GetUserByID(req.Context(), authUser.ID)

	// Generate new JWT with tenant ID
	token, expiresAt, err := r.generateJWT(user)
	if err != nil {
		r.logger.Printf("auth: failed to generate JWT: %v", err)
	}

	// Store new session
	if token != "" {
		tokenHash := hashToken(token)
		_ = r.store.CreateSession(req.Context(), user.ID, tokenHash, expiresAt)
	}

	r.logger.Printf("auth: onboarding complete for user %s, tenant %s", user.ID, tenant.ID)

	writeJSON(w, http.StatusOK, map[string]any{
		"tenant":     tenant,
		"user":       user,
		"token":      token,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
}

// generateDefaultSystemPrompt creates a default prompt for a new tenant
func generateDefaultSystemPrompt(name string) string {
	return fmt.Sprintf(`Jsi Karen, přátelská telefonní asistentka uživatele %s. %s teď nemá čas a ty přijímáš hovory za něj.

JIŽ JSI ŘEKLA ÚVODNÍ POZDRAV.

TVŮJ ÚKOL:
1. Zjisti co volající potřebuje od %s
2. Zjisti jméno volajícího
3. Zjisti telefonní číslo pro zpětný kontakt
4. Rozluč se zdvořile

PRAVIDLA:
- Mluv česky, přátelsky a stručně (1-2 věty)
- Neptej se na více věcí najednou
- Buď trpělivá, někteří lidé potřebují čas na odpověď
- NIKDY neříkej že hovor je "podezřelý" - prostě sbírej informace
- Při dotazu na telefon: zeptej se jestli můžeme použít číslo ze kterého volají, nebo jestli chtějí dát jiné
- České telefonní číslo má 9 číslic (např. 777 123 456). Pokud dostaneš méně než 9 číslic, zeptej se na zbytek!
- Když máš všechny informace (účel, jméno, telefon), rozluč se: "Děkuji, předám %s vzkaz. Na shledanou!"`, name, name, name, name)
}
