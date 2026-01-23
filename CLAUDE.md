# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Zvednu** is a multi-tenant AI phone assistant SaaS (Czech: "I'll pick it up for you"). When users miss calls, Karen (the AI) answers, captures caller intent, classifies legitimacy (spam/marketing/legitimate), and stores transcripts for review.

## Development Commands

### Backend (Go 1.23)
```bash
cd backend
go test -v ./...              # Run all tests
go test -v ./internal/store   # Run tests for specific package
golangci-lint run             # Lint (uses .golangci.yml config)
```

### Frontend (Node.js 20 / Vite)
```bash
cd frontend
npm install                   # Install dependencies
npm run dev                   # Dev server on :5173
npm run build                 # TypeScript check + Vite build
npm run test                  # Run tests once (vitest)
npm run test:watch            # Watch mode
npm run lint                  # ESLint
npm run lint:fix              # Auto-fix lint issues
npm run format                # Prettier formatting
```

### iOS (Swift 6 / Xcode)
```bash
cd ios
xcodegen generate             # Generate Xcode project from project.yml
open Zvednu.xcodeproj         # Open in Xcode to run locally
xcodebuild test -project Zvednu.xcodeproj -scheme Zvednu \
  -destination 'platform=iOS Simulator,name=iPhone 16 Pro' \
  CODE_SIGNING_ALLOWED=NO
```

To run the iOS app locally:
1. `cd ios && xcodegen generate` - generates the Xcode project from `project.yml`
2. `open Zvednu.xcodeproj` - opens in Xcode
3. Select a simulator and press Run (⌘R)

### Local Development
```bash
docker compose up --build     # Start all services (API :8080, Frontend :5173)
```

## Architecture

### Voice AI Pipeline
Real-time call flow: Twilio (telephony) → Deepgram Nova-3 (STT) → OpenAI GPT-4o-mini (LLM) → ElevenLabs (TTS)

1. Twilio routes inbound call → `POST /telephony/inbound` webhook
2. Backend returns TwiML opening WebSocket to `/media`
3. Audio streams bidirectionally (8kHz μ-law)
4. STT produces streaming transcripts with 800ms silence endpointing
5. LLM generates response, streamed sentence-by-sentence to TTS
6. TTS audio sent back through Twilio media stream
7. On hangup: transcript + screening result saved to Postgres

Key features: barge-in (caller interrupts), filler words ("Jasně...", "Rozumím..."), per-tenant customization.

### Backend Structure (`backend/internal/`)
- `app/` - Config loading, app initialization
- `httpapi/` - HTTP handlers organized by domain (auth, calls, admin, telephony, push)
- `store/` - PostgreSQL operations (pgx/v5), all database models
- `llm/` - OpenAI integration with prompt management
- `stt/` - Deepgram streaming STT
- `tts/` - ElevenLabs streaming TTS
- `eventlog/` - Call event persistence for debugging
- `notifications/` - Discord webhooks, APNs push

### Frontend Structure (`frontend/src/`)
- `api.ts` - Centralized HTTP client with typed API methods
- `AuthContext.tsx` - Global auth state (JWT in localStorage)
- `router.tsx` - React Router with route guards (ProtectedRoute, OnboardingRoute)
- `ui/` - Page components: CallInboxPage, CallDetailPage, SettingsPage, Admin pages
- `ui/landing/` - Public marketing pages

### iOS Structure (`ios/Zvednu/`)
- MVVM architecture with SwiftUI
- `Services/` - Actor-based APIClient, AuthService, CallService
- `ViewModels/` - @MainActor ObservableObject classes
- `Core/Utilities/KeychainManager.swift` - Secure token storage

### Multi-Tenancy Model
- `tenants` - Customer orgs with AI config (system_prompt, voice_id, greeting_text)
- `users` - Phone-authenticated users belonging to a tenant
- `tenant_phone_numbers` - Pool of Twilio numbers assigned to tenants
- `calls` - Call records with tenant isolation
- `call_utterances` - Transcript segments
- `call_screening_results` - AI classification (legitimacy_label, intent)

### Authentication Flow
1. User requests SMS OTP via `/auth/send-code` (Twilio Verify)
2. User verifies code via `/auth/verify-code`, receives JWT
3. JWT contains user_id, tenant_id, phone; validated via `withAuth` middleware
4. Session hash stored in DB for logout/revocation support
5. Admin access: phone must be in `ADMIN_PHONES` env var list

## Database Migrations

Migrations in `backend/migrations/` (001-008). Applied via:
```bash
# Local: manually run SQL files against Postgres
# Production: auto-applied during deploy via docker exec
```

## Key API Endpoints

**Auth**: `/auth/send-code`, `/auth/verify-code`, `/auth/refresh`, `/auth/logout`

**User API** (JWT required):
- `GET /api/me` - Current user + tenant
- `GET /api/calls` - List tenant calls
- `GET /api/calls/{id}` - Call detail with transcript
- `PATCH /api/calls/{id}/resolve` - Mark resolved
- `GET/PATCH /api/tenant` - Tenant settings

**Admin** (admin phone required):
- `/admin/phone-numbers` - Phone pool management
- `/admin/tenants` - Tenant CRUD
- `/admin/calls/{id}/events` - Call event timeline for debugging

## Environment Variables

Backend requires: `DATABASE_URL`, `JWT_SECRET`, `TWILIO_*`, `DEEPGRAM_API_KEY`, `OPENAI_API_KEY`, `ELEVENLABS_API_KEY`

Frontend requires: `VITE_API_BASE_URL`

See `backend/env.example` and `frontend/env.example` for full list.

## AI Debug API

The AI Debug API allows Claude CLI to remotely query call logs and update configuration for debugging.

**Authentication**: Use `X-API-Key` header with the API key from `AI_DEBUG_API_KEY` env var.

**Environment Variables** (set in Claude CLI):
- `ZVEDNU_API_URL` - Base API URL (e.g., `https://api.zvednu.cz`)
- `ZVEDNU_AI_API_KEY` - API key for authentication

**Important**: When using curl, use single quotes around URLs with query parameters to prevent shell glob expansion:
```bash
curl -s -H "X-API-Key: $ZVEDNU_AI_API_KEY" 'https://api.zvednu.cz/ai/calls?limit=5'
```

### Endpoints

**GET /ai/health** - Health check (no authentication required)
- Returns: `{"status": "ok", "api_configured": true}`
- Use this to verify API connectivity before attempting authenticated requests

**GET /ai/calls** - List recent calls
- Query params: `limit` (1-100, default 20), `tenant_id`, `since` (RFC3339 timestamp)
- Returns: calls with screening results (legitimacy, intent, entities)

**GET /ai/calls/{callSid}/events** - Get call details with full event timeline
- Returns: call info, utterances (transcript), and all diagnostic events
- Use Twilio call SID (e.g., `CA9509ef101828137de9e0ff9e85b755e9`)

**GET /ai/tenants/{tenantId}/calls** - List calls for specific tenant
- Query params: `limit` (1-100, default 20)

**GET /ai/stats** - Aggregate statistics
- Query params: `since` (RFC3339, default last 24h)
- Returns: total_calls, event_counts, avg_llm_latency_ms, avg_tts_latency_ms, max_turn_timeout_count, barge_in_count

**GET /ai/config** - List all config values

**PATCH /ai/config/{key}** - Update config value
- Body: `{"value": "5000"}`

### Key Event Types

When analyzing call events, look for these patterns:

| Event | Description |
|-------|-------------|
| `stt_result` | STT transcript chunk. Empty text with `speech_final: true` indicates silence/lost audio |
| `turn_finalized` | Caller's turn complete, ready for LLM |
| `max_turn_timeout` | Forced turn end due to timeout (may indicate cut-off speech) |
| `llm_first_token` | LLM started responding. `latency_ms` shows time-to-first-token |
| `tts_first_chunk` | TTS audio ready. `latency_ms` shows synthesis latency |
| `barge_in` | Caller interrupted the AI response |
| `stt_empty_streak` | Multiple consecutive empty STT results (audio issue) |
| `audio_silence_detected` | Prolonged low audio energy detected |

### Key Config Values

| Key | Default | Description |
|-----|---------|-------------|
| `max_turn_timeout_ms` | 4000 | Max wait for speech_final before forcing turn end |
| `adaptive_turn_enabled` | true | Enable adaptive timeout based on text length |
| `adaptive_text_decay_rate_ms` | 15 | Timeout reduction per character (Czech uses 8) |
| `adaptive_sentence_end_bonus_ms` | 1500 | Extra reduction for complete sentences (Czech uses 500) |
| `stt_debug_enabled` | false | Log raw Deepgram WebSocket messages |
| `robocall_detection_enabled` | true | Auto-detect and hang up on robocalls |

### Debugging Workflow

1. **Get recent calls**: Check for issues in last few calls
   ```bash
   curl -s -H "X-API-Key: $ZVEDNU_AI_API_KEY" 'https://api.zvednu.cz/ai/calls?limit=5'
   ```

2. **Get call events**: Analyze specific call's event timeline
   ```bash
   curl -s -H "X-API-Key: $ZVEDNU_AI_API_KEY" 'https://api.zvednu.cz/ai/calls/CA.../events'
   ```

3. **Check stats**: Look for patterns (high timeout counts, latency spikes)
   ```bash
   curl -s -H "X-API-Key: $ZVEDNU_AI_API_KEY" 'https://api.zvednu.cz/ai/stats?since=2026-01-23T00:00:00Z'
   ```

4. **Tune config**: Adjust timeout settings if needed
   ```bash
   curl -s -X PATCH -H "X-API-Key: $ZVEDNU_AI_API_KEY" -H "Content-Type: application/json" \
     -d '{"value":"5000"}' 'https://api.zvednu.cz/ai/config/max_turn_timeout_ms'
   ```

### Troubleshooting

**"missing API key" error:**

If your API calls return `{"error": "missing API key"}`, the environment variable may not be expanding properly. Verify with:

```bash
# Check if variable is set
echo "Key length: ${#ZVEDNU_AI_API_KEY}"
echo "First 8 chars: ${ZVEDNU_AI_API_KEY:0:8}"
```

If the variable appears empty in curl but is set in your shell:
1. Try exporting it: `export ZVEDNU_AI_API_KEY="your-key"`
2. Use the literal key value instead of the variable
3. Check for newlines/whitespace: `printenv ZVEDNU_AI_API_KEY | xxd | head`

**Verify API is reachable (no auth required):**

```bash
curl -s 'https://api.zvednu.cz/ai/health'
# Returns: {"status":"ok","api_configured":true}
```

**Alternative: Use literal key**

If environment variable expansion fails, use the key directly:
```bash
curl -s -H 'X-API-Key: your-64-char-key-here' 'https://api.zvednu.cz/ai/calls?limit=5'
```

### Example Responses

**GET /ai/calls** returns:
```json
{
  "calls": [
    {
      "provider_call_id": "CA9509ef101828137de9e0ff9e85b755e9",
      "from_number": "+420724794686",
      "to_number": "+420228811386",
      "status": "completed",
      "started_at": "2026-01-23T12:57:46.582748Z",
      "screening": {
        "legitimacy_label": "legitimní",
        "intent_text": "Dotaz na stav webu.",
        "entities_json": {"name": "Tepa", "purpose": "stav webu"}
      }
    }
  ],
  "count": 1
}
```

**GET /ai/stats** returns:
```json
{
  "since": "2026-01-23T00:00:00Z",
  "stats": {
    "total_calls": 19,
    "total_events": 673,
    "avg_llm_latency_ms": 1259.67,
    "avg_tts_latency_ms": 398.16,
    "max_turn_timeout_count": 30,
    "barge_in_count": 32
  }
}
```
