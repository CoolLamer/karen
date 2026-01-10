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
	"slices"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
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
		captureError(req, err, "failed to send verification code via Twilio")
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
		captureError(req, err, "Twilio verification check failed")
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
		sentry.CaptureException(err)
		http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
		return
	}

	// Generate JWT
	token, expiresAt, err := r.generateJWT(user)
	if err != nil {
		r.logger.Printf("auth: failed to generate JWT: %v", err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to create session"}`, http.StatusInternalServerError)
		return
	}

	// Store session for logout/revocation
	tokenHash := hashToken(token)
	if err := r.store.CreateSession(req.Context(), user.ID, tokenHash, expiresAt); err != nil {
		r.logger.Printf("auth: failed to store session: %v", err)
		sentry.CaptureException(err)
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

	// Check if user is admin
	isAdmin := slices.Contains(r.cfg.AdminPhones, user.Phone)

	writeJSON(w, http.StatusOK, map[string]any{
		"user":     user,
		"tenant":   tenant,
		"is_admin": isAdmin,
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

	// Get current tenant to check for name change
	currentTenant, err := r.store.GetTenantByID(req.Context(), *authUser.TenantID)
	if err != nil {
		http.Error(w, `{"error": "tenant not found"}`, http.StatusNotFound)
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

	// Check if we need to regenerate system prompt
	// (when name, vip_names, or marketing_email changes and system_prompt is not explicitly set)
	if _, hasExplicitPrompt := updates["system_prompt"]; !hasExplicitPrompt {
		needsRegeneration := false
		newName := currentTenant.Name
		newVIPNames := currentTenant.VIPNames
		newMarketingEmail := currentTenant.MarketingEmail

		// Check for name change
		if newNameVal, ok := updates["name"]; ok {
			if nameStr, ok := newNameVal.(string); ok && nameStr != "" && nameStr != currentTenant.Name {
				newName = nameStr
				// Only regenerate if current prompt appears to be auto-generated (contains old name)
				if strings.Contains(currentTenant.SystemPrompt, currentTenant.Name) {
					needsRegeneration = true
				}
			}
		}

		// Check for VIP names change
		if vipVal, ok := updates["vip_names"]; ok {
			if vipArr, ok := vipVal.([]interface{}); ok {
				newVIPNames = make([]string, 0, len(vipArr))
				for _, v := range vipArr {
					if s, ok := v.(string); ok && s != "" {
						newVIPNames = append(newVIPNames, s)
					}
				}
				needsRegeneration = true
			}
		}

		// Check for marketing email change
		if meVal, ok := updates["marketing_email"]; ok {
			if meStr, ok := meVal.(string); ok {
				newMarketingEmail = &meStr
				needsRegeneration = true
			}
		}

		// Regenerate prompt if needed
		if needsRegeneration {
			newPrompt := generateSystemPromptWithVIPs(newName, newVIPNames, newMarketingEmail)
			updates["system_prompt"] = newPrompt
			r.logger.Printf("auth: auto-regenerated system prompt for tenant %s", *authUser.TenantID)
		}
	}

	// Apply updates
	if err := r.store.UpdateTenant(req.Context(), *authUser.TenantID, updates); err != nil {
		r.logger.Printf("auth: failed to update tenant %s: %v", *authUser.TenantID, err)
		sentry.CaptureException(err)
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
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to create tenant"}`, http.StatusInternalServerError)
		return
	}

	// Update user name
	if err := r.store.UpdateUserName(req.Context(), authUser.ID, body.Name); err != nil {
		r.logger.Printf("auth: failed to update user name: %v", err)
		sentry.CaptureException(err)
	}

	// Assign user to tenant
	if err := r.store.AssignUserToTenant(req.Context(), authUser.ID, tenant.ID); err != nil {
		r.logger.Printf("auth: failed to assign user to tenant: %v", err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to assign tenant"}`, http.StatusInternalServerError)
		return
	}

	// Auto-assign an available phone number from the pool (if tenant doesn't already have one)
	var phoneNumber *store.TenantPhoneNumber
	existingNumbers, err := r.store.GetTenantPhoneNumbers(req.Context(), tenant.ID)
	if err != nil {
		r.logger.Printf("auth: failed to get existing phone numbers: %v", err)
		sentry.CaptureException(err)
	}

	if len(existingNumbers) > 0 {
		// Tenant already has a phone number (e.g., from a previous onboarding attempt)
		phoneNumber = &existingNumbers[0]
		r.logger.Printf("auth: tenant %s already has phone number %s", tenant.ID, phoneNumber.TwilioNumber)
	} else {
		// Claim a new phone number from the pool
		phoneNumber, err = r.store.ClaimAvailablePhoneNumber(req.Context(), tenant.ID)
		if err != nil {
			r.logger.Printf("auth: failed to claim phone number: %v", err)
			sentry.CaptureException(err)
			// Continue without phone number - not a fatal error
		} else if phoneNumber != nil {
			r.logger.Printf("auth: assigned phone number %s to tenant %s", phoneNumber.TwilioNumber, tenant.ID)
		} else {
			r.logger.Printf("auth: no available phone numbers for tenant %s", tenant.ID)
		}
	}

	// Get updated user
	user, _ := r.store.GetUserByID(req.Context(), authUser.ID)

	// Generate new JWT with tenant ID
	token, expiresAt, err := r.generateJWT(user)
	if err != nil {
		r.logger.Printf("auth: failed to generate JWT: %v", err)
		sentry.CaptureException(err)
	}

	// Store new session
	if token != "" {
		tokenHash := hashToken(token)
		_ = r.store.CreateSession(req.Context(), user.ID, tokenHash, expiresAt)
	}

	r.logger.Printf("auth: onboarding complete for user %s, tenant %s", user.ID, tenant.ID)

	// Include phone number in response if assigned
	response := map[string]any{
		"tenant":     tenant,
		"user":       user,
		"token":      token,
		"expires_at": expiresAt.Format(time.RFC3339),
	}
	if phoneNumber != nil {
		response["phone_number"] = phoneNumber
	}

	writeJSON(w, http.StatusOK, response)
}

// generateDefaultSystemPrompt creates a default prompt for a new tenant
func generateDefaultSystemPrompt(name string) string {
	return generateSystemPromptWithVIPs(name, nil, nil)
}

// generateSystemPromptWithVIPs creates a system prompt with VIP names and marketing email support
func generateSystemPromptWithVIPs(name string, vipNames []string, marketingEmail *string) string {
	basePrompt := fmt.Sprintf(`Jsi Karen, přátelská telefonní asistentka uživatele %s. %s teď nemá čas a ty přijímáš hovory za něj.

JIŽ JSI ŘEKLA ÚVODNÍ POZDRAV.

TVŮJ ÚKOL:
1. Zjisti co volající potřebuje od %s
2. Zjisti jméno volajícího
3. Rozluč se zdvořile

Pro zpětný kontakt automaticky použijeme číslo, ze kterého volají - netřeba se ptát.

PRAVIDLA:
- Mluv česky, přátelsky a stručně (1-2 věty)
- Neptej se na více věcí najednou
- Buď trpělivá, někteří lidé potřebují čas na odpověď
- NIKDY neříkej že hovor je "podezřelý" - prostě sbírej informace
- Když máš účel a jméno, rozluč se: "Děkuji, předám %s vzkaz. Na shledanou!"`, name, name, name, name)

	// Add VIP forwarding rules if VIP names configured
	if len(vipNames) > 0 {
		vipSection := "\n\nKRIZOVÉ SITUACE - OKAMŽITĚ PŘEPOJIT:\n"
		vipSection += fmt.Sprintf("- Pokud volající zmíní NEBEZPEČÍ nebo NOUZI týkající se blízkých %s (rodina, přátelé) → řekni: \"[PŘEPOJIT] Rozumím, přepojuji vás přímo.\"\n", name)
		for _, vip := range vipNames {
			if vip != "" {
				vipSection += fmt.Sprintf("- Pokud se volající představí jako \"%s\" → řekni: \"[PŘEPOJIT] Přepojuji tě.\"\n", vip)
			}
		}
		basePrompt += vipSection
	}

	// Add marketing email handling if configured
	if marketingEmail != nil && *marketingEmail != "" {
		basePrompt += fmt.Sprintf("\n\nMARKETING:\n- U marketingu a nabídek: řekni že %s nemá zájem, ale pokud chtějí, mohou nabídku poslat na email %s", name, *marketingEmail)
	} else {
		basePrompt += fmt.Sprintf("\n\nMARKETING:\n- U marketingu a nabídek: zdvořile odmítni a řekni že %s nemá zájem", name)
	}

	return basePrompt
}
