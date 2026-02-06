# Monetization Plan for Zvednu (Karen)

## Executive Summary

Zvednu is a multi-tenant AI phone assistant SaaS with:
- **Web app** (React) - complete with billing UI
- **iOS app** (SwiftUI) - complete with push notifications and billing UI
- **Billing infrastructure** - Stripe integration complete (checkout, portal, webhooks)

### Implementation Status

| Phase | Status |
|-------|--------|
| Phase 1: Usage Tracking | ✅ Complete |
| Phase 2: Trial Enforcement + Time Saved | ✅ Complete |
| Phase 3: Stripe Integration | ✅ Complete (needs Stripe account setup) |
| Phase 4: Analytics & Monitoring | 🔄 Partial |

This plan covers:
1. Pricing tiers and billing model
2. Trial implementation (different for web vs iOS)
3. Usage metering and monitoring
4. Technical implementation roadmap

---

## 1. Pricing Strategy

### Recommended Tiers (from MARKETING_CZ.md)

| Tier | Price | Call Limit | Features | Target |
|------|-------|------------|----------|--------|
| **Trial** | 0 CZK | 20 calls OR 14 days | Basic transcript, SMS notifications | Everyone |
| **Zaklad** | 199 CZK/mo | 50 calls | Full classification, SMS notifications | OSVČ, anti-spam users |
| **Pro** | 499 CZK/mo | Unlimited | VIP forwarding, custom voice, priority support | Professionals |
| **Firma** | Contact sales | Custom | Multiple numbers, team, API, SLA | Businesses |

**Note**: No email - all user communication via SMS (users only have phone numbers).

### Billing Model Decision (Confirmed)

**Flat monthly + hard call limits**

- Simple for users to understand
- Predictable revenue
- Per-call metering would be complex and create anxiety
- **Hard limits**: warn at 80%, **block at 100%** (must upgrade)
- Trial: 20 calls OR 14 days (whichever first)

### Annual Discount
- 20% off annual (2 months free)
- Shows as: "199 CZK/mo" vs "159 CZK/mo (billed annually)"

### iOS App Payment Strategy (Confirmed: Stripe Only)

**Decision**: Stripe via Safari redirect (no StoreKit)

- All payments via Stripe Checkout (Safari opens from iOS app)
- **No Apple commission** (15-30% saved)
- Same billing system for web and iOS
- User taps "Upgrade" → Safari opens Stripe Checkout → returns to app

**Why this is allowed**: Zvednu is a phone answering service (happens outside app). The app is just a dashboard - not consuming digital content in-app. Similar to Dropbox, Slack, VoIP apps.

---

## 2. Trial Implementation

### Trial Parameters (Confirmed)

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| Duration | 14 days OR 20 calls | **Whichever limit hits first** |
| Features | Full Basic tier | Let users experience full value |
| Credit card required | See below | Platform-specific strategy |
| Conversion prompt | Day 10, 12, 14 | Push notification + in-app reminders |

**Platform-specific trial strategy:**
| Platform | Card Required | Rationale |
|----------|---------------|-----------|
| Web | No (frictionless) | Lower friction, higher trial signups |
| iOS (now) | No (frictionless) | Same as web for launch |
| iOS (later) | Yes | Higher conversion rate, less abuse |

### Trial States

```
NEW_USER -> TRIAL_ACTIVE -> TRIAL_EXPIRED -> CHURNED
                    |                |
                    v                v
              PAID_ACTIVE <----- PAID_ACTIVE
```

### Trial Expiration Logic

```sql
-- New fields for tenants table
ALTER TABLE tenants ADD COLUMN trial_ends_at TIMESTAMPTZ;
ALTER TABLE tenants ADD COLUMN calls_used_this_period INT DEFAULT 0;
ALTER TABLE tenants ADD COLUMN period_started_at TIMESTAMPTZ;

-- Trial expires when:
-- 1. created_at + 14 days < NOW(), OR
-- 2. calls_used_this_period >= 20 (for trial plan)
```

### Trial User Experience

1. **Sign up**: No payment required, get Trial plan
2. **Onboarding**: Full onboarding flow (already exists for web + iOS)
3. **Usage**: See remaining calls/days + time saved in dashboard
4. **Day 10**: Push notification (iOS) / SMS (web): "Zbývají ti 4 dny trialu. Karen ti zatím ušetřila X minut."
5. **Day 12**: In-app banner "Trial ending soon, upgrade to keep Karen"
6. **Day 14 / 20 calls**:
   - Can still view past calls in dashboard
   - Push/SMS: "Trial skončil. Upgraduj na zvednu.cz"
   - Karen simply doesn't answer (call rings through/voicemail)

**Notification channels (implemented):**
- iOS: APNs push notifications (via `backend/internal/notifications/apns.go`)
- Web/iOS: SMS via Twilio Programmable Messaging (via `backend/internal/notifications/sms.go`)

### Trial Grace Period and Phone Number Release

After trial expiration, tenants have a configurable grace period (default: 7 days) before their phone number is released back to the pool. This gives users time to upgrade while also ensuring phone numbers are recycled for new users.

**Timeline:**

```
Day 1-9:   Trial active, Karen answers calls
Day 10:    SMS + Push: "Zbývají ti 4 dny trialu. Karen ti ušetřila X minut."
Day 12:    SMS + Push: "Zbývají ti 2 dny trialu. Karen ti vyřídila X hovorů."
Day 14:    Trial expires
           - Karen stops answering calls
           - SMS + Push: "Trial skončil. Upgraduj na zvednu.cz"
           - Grace period starts
Day 15-20: Grace period (user can still upgrade to keep their number)
           - SMS: "Za X dní bude vaše číslo odpojeno. Zrušte přesměrování nebo upgradujte."
Day 21:    Phone number released
           - SMS: "Číslo odpojeno. Prosím zrušte přesměrování hovorů."
           - Number returns to available pool
           - Tenant status set to "churned"
```

**Notifications sent:**

| Event | Channel | Message (Czech) |
|-------|---------|-----------------|
| Day 10 (4 days left) | SMS + Push | "Zbývají ti 4 dny trialu. Karen ti zatím ušetřila X minut. Upgraduj na zvednu.cz" |
| Day 12 (2 days left) | SMS + Push | "Zbývají ti 2 dny trialu. Karen ti vyřídila X hovorů. Upgraduj na zvednu.cz" |
| Day 14 (expired) | SMS + Push | "Trial skončil. Karen nebude přijímat hovory. Upgraduj na zvednu.cz" |
| Grace warning | SMS + Push | "Za X dní bude vaše číslo +420XXX odpojeno. Zrušte přesměrování nebo upgradujte." |
| Number released | SMS only | "Číslo +420XXX odpojeno. Prosím zrušte přesměrování hovorů. Pro obnovení: zvednu.cz" |

**Configuration (via global_config / admin page):**

| Config Key | Default | Description |
|------------|---------|-------------|
| `trial_grace_period_days` | 7 | Days after trial expiration before releasing phone number |
| `sms_sender_number` | (required) | Twilio phone number for sending SMS notifications (E.164 format, e.g., +420123456789) |

Configure via admin page, global_config table, or AI Debug API:
```bash
# Set SMS sender number
curl -X PATCH -H "X-API-Key: $ZVEDNU_AI_API_KEY" \
  -d '{"value":"+420123456789"}' 'https://api.zvednu.cz/ai/config/sms_sender_number'

# Set grace period to 14 days
curl -X PATCH -H "X-API-Key: $ZVEDNU_AI_API_KEY" \
  -d '{"value":"14"}' 'https://api.zvednu.cz/ai/config/trial_grace_period_days'
```

**Note:** The `sms_sender_number` must be configured in global_config for SMS notifications to work. The job fetches this value on each run, so changes take effect immediately without restart.

**Important:** Users must manually cancel call forwarding from their phone settings. The SMS notifications remind them which number to remove from their forwarding rules.

**Database tracking fields (in `tenants` table):**
- `trial_day10_notification_sent_at` - When day 10 notification was sent
- `trial_day12_notification_sent_at` - When day 12 notification was sent
- `trial_day14_notification_sent_at` - When trial expired notification was sent
- `trial_grace_notification_sent_at` - When grace period warning was sent
- `phone_number_released_at` - When phone number was released

**Background job:**
The `TrialLifecycleJob` runs hourly (configurable via `TRIAL_LIFECYCLE_JOB_INTERVAL` env var) and:
1. Sends Day 10/12/14 conversion prompts
2. Sends grace period warnings
3. Releases phone numbers after grace period

---

## 3. Usage Metering & Monitoring

### What to Track

| Metric | Table | Purpose |
|--------|-------|---------|
| Calls per tenant/month | `tenant_usage` | Plan enforcement |
| Minutes per call | `calls.duration_seconds` | Future per-minute billing |
| STT/LLM/TTS costs | `call_costs` | Unit economics monitoring |
| Active users | `users.last_login_at` | Churn prediction |

### New Database Schema

```sql
-- Monthly usage tracking
CREATE TABLE tenant_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,  -- First day of month
    period_end DATE NOT NULL,    -- Last day of month
    calls_count INT DEFAULT 0,
    minutes_used INT DEFAULT 0,  -- Total call duration in minutes
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, period_start)
);

-- Per-call cost tracking (for unit economics)
CREATE TABLE call_costs (
    call_id UUID PRIMARY KEY REFERENCES calls(id) ON DELETE CASCADE,
    twilio_cost_cents INT,       -- Twilio charges
    stt_cost_cents INT,          -- Deepgram charges
    llm_cost_cents INT,          -- OpenAI charges
    tts_cost_cents INT,          -- ElevenLabs charges
    total_cost_cents INT,        -- Sum of all costs
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Add to tenants table
ALTER TABLE tenants ADD COLUMN
    stripe_customer_id TEXT,
    stripe_subscription_id TEXT,
    trial_ends_at TIMESTAMPTZ,
    current_period_start DATE,
    current_period_calls INT DEFAULT 0;
```

### Plan Enforcement Points (Confirmed: Hard Block)

1. **Inbound call webhook** (`/telephony/inbound`):
   - Check tenant status (active/suspended/cancelled)
   - Check plan limits (calls remaining)
   - **If over limit**: Simply don't answer - let call go to voicemail/ring out
     - Alternative: Answer briefly "Omlouvám se, nemohu teď hovořit" and hang up (sounds like busy owner)
     - **Never mention "upgrade" to callers** - that's unprofessional

2. **Post-call processing**:
   - Increment `current_period_calls`
   - Record in `tenant_usage`
   - Check if approaching limit (80%) -> trigger SMS warning to user
   - At 100%: SMS to user "Dosáhli jste limitu, Karen nebude přijímat hovory. Upgradujte na zvednu.cz"

### User Dashboard: "Time Saved" Metric

**Show users how much time Karen saved them** - key retention feature.

Calculation:
- Each call Karen handles = time saved (call duration + assumed callback time)
- Formula: `time_saved = call_duration + 2 minutes` (2 min = average callback overhead)
- Spam calls: `time_saved = call_duration + 5 minutes` (spam takes longer to get rid of)

Display in dashboard:
```
┌─────────────────────────────────────────────┐
│  Karen ti ušetřila                          │
│  ┌─────────────────────────────────────────┐│
│  │   🕐  2h 34min   tento měsíc            ││
│  │   📞  47 hovorů  (12 spam blokováno)    ││
│  └─────────────────────────────────────────┘│
│  Celkem od začátku: 8h 12min               │
└─────────────────────────────────────────────┘
```

Database:
```sql
-- Add to tenant_usage table
ALTER TABLE tenant_usage ADD COLUMN
    time_saved_seconds INT DEFAULT 0,
    spam_calls_blocked INT DEFAULT 0;

-- Or track per-call and aggregate
-- calls.duration_seconds already exists
-- Add calls.time_saved_seconds = duration + (is_spam ? 300 : 120)
```

### Admin Dashboard Metrics

Already partially exists. Add:
- Revenue metrics (MRR, ARR)
- Usage breakdown (calls/tenant, overage tracking)
- Conversion funnel (trial -> paid)
- Churn tracking

---

## 4. Stripe Integration

### Environment Variables

The following environment variables are required for Stripe integration:

```bash
# Stripe API Keys
STRIPE_SECRET_KEY=sk_live_...           # Or sk_test_... for development
STRIPE_WEBHOOK_SECRET=whsec_...         # Webhook signing secret

# Stripe Price IDs (create in Stripe Dashboard)
STRIPE_PRICE_BASIC_MONTHLY=price_...    # Basic plan monthly price ID
STRIPE_PRICE_BASIC_ANNUAL=price_...     # Basic plan annual price ID
STRIPE_PRICE_PRO_MONTHLY=price_...      # Pro plan monthly price ID
STRIPE_PRICE_PRO_ANNUAL=price_...       # Pro plan annual price ID

# Stripe Checkout URLs (optional - defaults to PublicBaseURL)
STRIPE_SUCCESS_URL=https://zvednu.cz/billing/success?session_id={CHECKOUT_SESSION_ID}
STRIPE_CANCEL_URL=https://zvednu.cz/billing/cancel
```

### Stripe Setup

1. **Products** (in Stripe dashboard):
   - `prod_zaklad`: Basic plan, 199 CZK/month
   - `prod_pro`: Pro plan, 499 CZK/month
   - `prod_firma`: Enterprise, custom pricing

2. **Price IDs**:
   - `price_zaklad_monthly`
   - `price_zaklad_annual`
   - `price_pro_monthly`
   - `price_pro_annual`

### Integration Points

| Endpoint | Purpose |
|----------|---------|
| `POST /api/billing/checkout` | Create Stripe Checkout session |
| `POST /api/billing/portal` | Create Stripe Customer Portal session |
| `POST /webhooks/stripe` | Handle Stripe events |
| `GET /api/billing/status` | Get current subscription status |

### Stripe Webhook Events to Handle

| Event | Action |
|-------|--------|
| `checkout.session.completed` | Upgrade tenant plan, set status=active |
| `customer.subscription.updated` | Update plan if changed |
| `customer.subscription.deleted` | Set status=cancelled, plan=trial |
| `invoice.payment_succeeded` | Log payment, reset period_calls |
| `invoice.payment_failed` | Send email, grace period |

### Checkout Flow

```
1. User clicks "Upgrade" in dashboard
2. Frontend calls POST /api/billing/checkout
3. Backend creates Stripe Checkout session
4. User redirected to Stripe Checkout
5. Payment succeeds -> Stripe webhook
6. Backend updates tenant plan/status
7. User redirected to success page
```

---

## 5. Implementation Roadmap

### Phase 1: Usage Tracking (Foundation) ✅ COMPLETE
**Files modified:**
- `backend/migrations/009_billing.sql` - New tables
- `backend/internal/store/store.go` - Add usage methods
- `backend/internal/httpapi/media_ws.go` - Increment usage on call

**Tasks:**
- [x] Add `tenant_usage` table
- [x] Add `call_costs` table
- [x] Add billing fields to `tenants` table
- [x] Implement `IncrementTenantUsage()` store method
- [x] Call it after each successful call
- [x] Add usage to admin tenant list

### Phase 2: Trial Enforcement + Time Saved ✅ COMPLETE
**Files modified:**
- `backend/internal/httpapi/twilio_handlers.go` - Inbound call limit check (TwiML Reject)
- `backend/internal/httpapi/auth_handlers.go` - `/api/billing` endpoint
- `backend/internal/notifications/apns.go` - Trial warning push
- `frontend/src/ui/CallInboxPage.tsx` - Show time saved widget + trial status
- `ios/Zvednu/Views/Inbox/CallInboxView.swift` - Show time saved widget + trial status

**Tasks:**
- [x] Add trial_ends_at to tenant creation (14 days from signup)
- [x] Check trial status on inbound calls (TwiML Reject with "busy" if over limit)
- [x] Calculate time_saved per call (duration + 2min, or +5min for spam)
- [x] Show trial countdown in UI (web + iOS)
- [x] Show "Time saved" widget in dashboard (web + iOS)
- [x] Send push notifications (iOS) at 80% limit and expiration
- [x] Handle expired trial (Karen doesn't answer)

### Phase 3: Stripe Integration ✅ COMPLETE
**Files created/modified:**
- `backend/internal/httpapi/billing_handlers.go` - Stripe endpoints
- `backend/internal/httpapi/router.go` - Added billing routes
- `frontend/src/api.ts` - Checkout and portal API methods
- `frontend/src/ui/SettingsPage.tsx` - Subscription management UI
- `ios/Zvednu/Services/TenantService.swift` - Checkout/portal methods
- `ios/Zvednu/ViewModels/SettingsViewModel.swift` - Billing state
- `ios/Zvednu/Views/Settings/SettingsView.swift` - Subscription section + upgrade sheet

**Tasks:**
- [ ] Set up Stripe account (Czech entity) - **MANUAL STEP**
- [ ] Create products and prices in Stripe Dashboard - **MANUAL STEP**
- [x] Implement checkout endpoint (returns Stripe Checkout URL)
- [x] Implement webhook handler
- [x] Add billing section to Settings page (web)
- [x] Add subscription section to iOS Settings (opens Safari for Stripe Checkout)
- [x] Add `/api/billing` endpoint for status (used by iOS + web)
- [ ] Test full flow (trial -> paid) on both platforms - **TESTING REQUIRED**

### Phase 4: Analytics & Monitoring
**Tasks:**
- [ ] Add revenue metrics to admin dashboard
- [x] Implement usage alerts (80% of limit) - via APNs push
- [ ] Build conversion funnel tracking
- [ ] Set up Stripe revenue reports

---

## 6. Key Files to Modify

### Backend
| File | Changes |
|------|---------|
| `backend/migrations/` | New migration for billing tables |
| `backend/internal/store/store.go` | Usage tracking methods |
| `backend/internal/httpapi/handlers.go` | Usage increment on calls |
| `backend/internal/httpapi/billing_handlers.go` | New file for Stripe |
| `backend/internal/voiceagent/session.go` | Plan limit checks |
| `backend/internal/notifications/apns.go` | Trial warning push notifications |

### Web Frontend
| File | Changes |
|------|---------|
| `frontend/src/ui/SettingsPage.tsx` | Billing section |
| `frontend/src/ui/DashboardLayout.tsx` | Trial status banner |
| `frontend/src/ui/CallInboxPage.tsx` | Time saved widget |

### iOS App
| File | Changes |
|------|---------|
| `ios/Zvednu/Views/Settings/SettingsView.swift` | Add subscription section + trial status |
| `ios/Zvednu/Views/Inbox/CallInboxView.swift` | Time saved widget |
| `ios/Zvednu/Services/BillingService.swift` | New file for billing API |
| `ios/Zvednu/ViewModels/SettingsViewModel.swift` | Add billing state |

---

## 7. Decisions Made

| Question | Decision |
|----------|----------|
| Trial limits | Both: 14 days OR 20 calls (whichever first) |
| Credit card for trial | Frictionless (no card) for web + iOS now; require card for iOS later |
| Overage handling | **Hard block** - Karen doesn't answer, push/SMS to user |
| Firma tier | **Contact sales** (quote-based, companies vary in size) |
| Communication | **Push (iOS) + SMS (web)** - no email |
| Limit exceeded behavior | Karen doesn't answer (no "upgrade" message to callers) |
| User engagement | Show **"Time saved"** metric in dashboard |
| iOS payments | **Stripe via Safari** (avoid Apple's 15-30% commission) |

### Remaining Open Questions

1. **Annual billing**: Offer from start, or add later?
2. **Czech payment methods**: Bank transfer support needed?

---

## 8. Verification Plan

### Testing Trial Flow (Web + iOS)
1. Create new account -> verify trial_ends_at set
2. Make calls -> verify usage increments
3. Verify "Time saved" widget shows correct value
4. At 80% limit -> verify push notification (iOS) / SMS (web) received
5. Hit 20 calls -> verify Karen doesn't answer
6. Wait 14 days -> verify expiration handling

### Testing Payment Flow
**Web:**
1. Click upgrade -> Stripe Checkout opens
2. Complete payment -> plan updates
3. Check webhook logs -> events processed

**iOS:**
1. Tap upgrade in Settings -> Safari opens Stripe Checkout
2. Complete payment -> return to app
3. Verify subscription status updates in iOS app

### Testing Plan Enforcement
1. Downgrade to trial -> verify limits apply
2. Upgrade to Pro -> verify unlimited access
3. Cancel subscription -> verify status changes
4. Test on both web and iOS apps
