# Karen - Productization Strategy

## Vision

Transform Karen from a single-user call screener into a **multi-tenant SaaS** where:
- Each user gets their own phone assistant
- Users can customize their prompt/persona
- Each user gets a dedicated phone number (or uses call forwarding detection)

## Architecture Evolution

### Current State (MVP)
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Twilio    │────▶│   Backend   │────▶│  PostgreSQL │
│  (1 number) │     │  (1 prompt) │     │  (no tenant)│
└─────────────┘     └─────────────┘     └─────────────┘
```

### Target State (Multi-tenant)
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Twilio    │────▶│   Backend   │────▶│  PostgreSQL │
│ (N numbers) │     │ (N prompts) │     │ (tenants)   │
└─────────────┘     │   + Tenant  │     └─────────────┘
                    │   Routing   │
                    └─────────────┘
```

## Phone Number Routing Options

### Option A: Dedicated Number per Tenant (Recommended)

**How it works:**
1. Each tenant provisions a Twilio phone number (~$1/month)
2. When call comes in, `To` field identifies the Twilio number
3. Backend looks up tenant by phone number → applies their prompt

**Pros:**
- Simple, reliable routing
- No carrier dependency
- Clear billing per tenant
- Users can give out their "assistant" number directly

**Cons:**
- Each tenant needs a Twilio number
- Small additional cost

**Implementation:**
```sql
-- Lookup tenant by the Twilio number that received the call
SELECT t.* FROM tenants t
JOIN tenant_phone_numbers pn ON pn.tenant_id = t.id
WHERE pn.twilio_number = $1; -- $1 = "To" from webhook
```

### Option B: Call Forwarding Detection

**How it works:**
1. Single Twilio number for all tenants
2. Users forward their personal number to the Twilio number
3. Twilio passes `ForwardedFrom` header (when available)
4. Backend looks up tenant by original number

**Twilio headers available:**
- `ForwardedFrom` - Original number before forwarding (when carrier provides it)
- `From` - Caller's number
- `To` - Twilio number that received call

**Reality check:**
- **PSTN forwarding**: Carriers often DON'T pass the original number reliably
- **SIP trunks**: Headers like `Diversion` / `History-Info` are more reliable
- **Conclusion**: Not reliable for general PSTN use

**When this works:**
- If using SIP trunking with carriers that preserve headers
- Enterprise VoIP systems with proper header propagation

### Option C: Hybrid Approach

1. Dedicated Twilio number for paid tenants (reliable)
2. Shared number + forwarding detection for trials (best effort)
3. IVR fallback: "Please enter your 4-digit tenant code"

## Multi-Tenant Database Schema

### New Tables

```sql
-- Tenants (customers/organizations)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,

    -- AI Configuration
    system_prompt TEXT NOT NULL,           -- Custom agent prompt
    greeting_text TEXT,                     -- Custom greeting
    voice_id TEXT,                          -- ElevenLabs voice ID
    language TEXT DEFAULT 'cs',             -- STT/TTS language

    -- Behavior settings
    vip_names TEXT[],                       -- Names to forward immediately
    marketing_email TEXT,                   -- Email for marketing redirects
    forward_number TEXT,                    -- Number to forward urgent calls

    -- Subscription
    plan TEXT DEFAULT 'trial',              -- trial, basic, pro
    status TEXT DEFAULT 'active',           -- active, suspended, cancelled

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Users (authenticated via phone)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    phone TEXT UNIQUE NOT NULL,            -- E.164 format: +420777123456
    phone_verified BOOLEAN DEFAULT false,
    name TEXT,
    role TEXT DEFAULT 'owner',             -- owner, admin, member
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Phone numbers assigned to tenants (for incoming calls)
CREATE TABLE tenant_phone_numbers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    twilio_number TEXT UNIQUE NOT NULL,    -- E.164 format: +1234567890
    twilio_sid TEXT,                        -- Twilio Phone Number SID

    -- For forwarding detection fallback
    forwarding_source TEXT,                 -- User's original number

    is_primary BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Extend calls table
ALTER TABLE calls ADD COLUMN tenant_id UUID REFERENCES tenants(id);
CREATE INDEX idx_calls_tenant ON calls(tenant_id);
```

### Routing Logic

```go
func (h *Handler) InboundCall(w http.ResponseWriter, r *http.Request) {
    toNumber := r.FormValue("To")
    forwardedFrom := r.FormValue("ForwardedFrom")

    // Try primary lookup by Twilio number
    tenant, err := h.store.GetTenantByTwilioNumber(ctx, toNumber)
    if err != nil && forwardedFrom != "" {
        // Fallback: try forwarding source lookup
        tenant, err = h.store.GetTenantByForwardingSource(ctx, forwardedFrom)
    }
    if err != nil {
        // Unknown caller - use default tenant or reject
        tenant = h.defaultTenant
    }

    // Create call record with tenant context
    call := &Call{
        TenantID: tenant.ID,
        // ... other fields
    }

    // Use tenant's custom configuration
    session := &CallSession{
        SystemPrompt:  tenant.SystemPrompt,
        GreetingText:  tenant.GreetingText,
        VoiceID:       tenant.VoiceID,
        VIPNames:      tenant.VIPNames,
        // ...
    }
}
```

## Infrastructure Options

### Option 1: Stay on Coolify (Recommended for Start)

**Keep current setup, add multi-tenancy:**
- Same deployment model
- Database-level multi-tenancy
- Single instance serves all tenants
- Scale vertically first

**Limits:**
- ~100 concurrent calls before needing horizontal scaling
- Good enough for initial product

### Option 2: Migrate to Railway/Render

**Why consider:**
- Better developer experience
- Built-in PostgreSQL with auto-backups
- Auto-scaling when needed
- Simpler CI/CD

**Railway specifics:**
- Native PostgreSQL addon
- WebSocket support
- $5-20/month for this workload

### Option 3: Kubernetes (Future Scale)

**When to consider:**
- 1000+ tenants
- Need per-tenant isolation
- Complex compliance requirements

**Stack:**
- Managed K8s (GKE, EKS, DigitalOcean)
- PostgreSQL operator or managed DB
- Horizontal pod autoscaling

## API Extensions for Multi-Tenancy

### Tenant Management API

```
POST   /api/tenants                 # Create tenant
GET    /api/tenants/:id             # Get tenant
PATCH  /api/tenants/:id             # Update tenant
DELETE /api/tenants/:id             # Delete tenant

POST   /api/tenants/:id/phone-numbers   # Provision phone number
DELETE /api/tenants/:id/phone-numbers/:phone # Release number

GET    /api/tenants/:id/calls       # List tenant's calls
GET    /api/tenants/:id/analytics   # Usage stats
```

### Authentication (Twilio Verify - Phone OTP)

**Stack:** Twilio Verify for SMS OTP + JWT sessions

**Flow:**
```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Frontend   │     │   Backend    │     │Twilio Verify │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                    │
       │ 1. Enter phone     │                    │
       │───────────────────▶│                    │
       │                    │ 2. Start verify    │
       │                    │───────────────────▶│
       │                    │                    │
       │                    │◀───────────────────│
       │ 3. Show OTP input  │    SMS sent        │
       │◀───────────────────│                    │
       │                    │                    │
       │ 4. Enter OTP       │                    │
       │───────────────────▶│ 5. Check verify    │
       │                    │───────────────────▶│
       │                    │                    │
       │                    │◀───────────────────│
       │ 6. JWT token       │    approved        │
       │◀───────────────────│                    │
```

**API Endpoints:**
```
POST /auth/send-code     { phone: "+420777123456" }
                         → { success: true }

POST /auth/verify-code   { phone: "+420777123456", code: "123456" }
                         → { token: "jwt...", user: {...} }

POST /auth/refresh       { token: "jwt..." }
                         → { token: "new-jwt..." }

POST /auth/logout        (invalidate token)
```

**Database:**
```sql
-- Users table (linked to tenants)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    phone TEXT UNIQUE NOT NULL,           -- E.164 format
    phone_verified BOOLEAN DEFAULT false,

    -- Profile
    name TEXT,

    -- Session management
    last_login_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- For token invalidation (optional - for logout/revoke)
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,             -- SHA256 of JWT
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Go Implementation:**
```go
// Send verification code
func (h *AuthHandler) SendCode(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Phone string `json:"phone"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    // Validate phone format (E.164)
    if !isValidE164(req.Phone) {
        http.Error(w, "Invalid phone format", 400)
        return
    }

    // Start Twilio Verify
    _, err := h.twilioClient.Verify.Verifications.Create(
        h.verifyServiceSID,
        &verify.CreateVerificationParams{
            To:      req.Phone,
            Channel: "sms",
        },
    )
    if err != nil {
        http.Error(w, "Failed to send code", 500)
        return
    }

    json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Verify code and issue JWT
func (h *AuthHandler) VerifyCode(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Phone string `json:"phone"`
        Code  string `json:"code"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    // Check with Twilio Verify
    check, err := h.twilioClient.Verify.VerificationChecks.Create(
        h.verifyServiceSID,
        &verify.CreateVerificationCheckParams{
            To:   req.Phone,
            Code: req.Code,
        },
    )
    if err != nil || check.Status != "approved" {
        http.Error(w, "Invalid code", 401)
        return
    }

    // Find or create user
    user, err := h.store.FindOrCreateUser(ctx, req.Phone)
    if err != nil {
        http.Error(w, "Database error", 500)
        return
    }

    // Generate JWT
    token := h.generateJWT(user)

    json.NewEncoder(w).Encode(map[string]any{
        "token": token,
        "user":  user,
    })
}
```

**JWT Structure:**
```json
{
  "sub": "user-uuid",
  "tenant_id": "tenant-uuid",
  "phone": "+420777123456",
  "exp": 1234567890,
  "iat": 1234567890
}
```

**Environment Variables:**
```
TWILIO_ACCOUNT_SID=ACxxxxx
TWILIO_AUTH_TOKEN=xxxxx
TWILIO_VERIFY_SERVICE_SID=VAxxxxx
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h
```

**Cost:** ~$0.05 per SMS verification

**Future expansion:**
- Add Google OAuth when needed (verify ID token)
- Add Apple Sign-In when needed (verify ID token)
- Add email magic link as alternative

## Phased Implementation

### Phase 1: Database Multi-Tenancy + Auth
1. Add `tenants` table
2. Add `users` table
3. Add `tenant_phone_numbers` table
4. Add `tenant_id` to `calls` table
5. Implement Twilio Verify phone auth
6. Implement JWT session management
7. Implement routing by `To` number
8. Use tenant's prompt instead of hardcoded

### Phase 2: Web App Self-Service
1. Landing page (explain value proposition)
2. Phone login UI (send code → verify)
3. Onboarding wizard (personalize → get number → setup forwarding → test)
4. Call inbox dashboard
5. Call detail with transcript
6. Settings (profile, assistant config)
7. Prompt customization UI

See [UX.md](UX.md) for detailed screen designs and user flows.

### Phase 3: Billing & Scale
1. Stripe integration
2. Usage metering (minutes, calls)
3. Plan limits enforcement
4. Horizontal scaling if needed

### Phase 4: Mobile Apps
1. React Native or native development
2. Push notifications (APNs + FCM)
3. Call forwarding shortcuts (one-tap enable/disable)
4. Home screen widget
5. Siri/Google Assistant integration ("Hey Siri, enable Karen")

See [UX.md](UX.md) for mobile-specific considerations.

## Cost Structure (Estimates)

### Per Tenant Costs
- Twilio phone number: $1-2/month
- Twilio voice: ~$0.015/min (both legs)
- Deepgram STT: ~$0.0059/min
- OpenAI GPT-4o-mini: ~$0.15-0.60 per 1M tokens
- ElevenLabs TTS: ~$0.18/1K chars (or $5-22/month plans)

### Infrastructure (Shared)
- Coolify/Railway: $5-20/month
- PostgreSQL: Included or $7-15/month
- Total fixed: ~$20-50/month for small scale

### Unit Economics
- Avg call: 2 min = ~$0.06-0.10 total cost
- 100 calls/month per tenant = ~$6-10 cost
- Can price at $20-50/month per tenant for healthy margins

## Open Questions

1. **Billing model**: Per call, per minute, or flat monthly?
2. **Number provisioning**: Self-service or manual?
3. **Voice selection**: Let users choose from ElevenLabs voices?
4. **Recording storage**: Offer call recordings? (storage costs, privacy)
5. **Integrations**: Webhooks? Slack? Email summaries?
