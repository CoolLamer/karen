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
