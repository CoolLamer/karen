# Deployment & Stack Specification (Coolify + Twilio + Go backend + Vite/React frontend)

## Stack Choices (Locked In)
- **Telephony**: Twilio Programmable Voice (webhooks + Media Streams over WebSocket)
- **Backend**: Go (Golang)
- **Frontend**: Vite + React + TypeScript + UI kit (recommended: Chakra UI, Mantine, or MUI)
- **Database**: PostgreSQL (recommended) for calls/transcripts/screening results
- **Cache/Queue (optional)**: Redis (if you later add async jobs)
- **Deployment**: Coolify (Docker-based deployment with HTTPS + reverse proxy)

---

## Twilio Routing: Can we recognize from which number the call was redirected?
**Yes for the important parts**:
- **`From`**: the caller's phone number (the real person/spammer calling).
- **`To` / `Called`**: the Twilio number that received the call.
- **`ForwardedFrom`**: The original number before forwarding (when carrier provides it).

**About the "original number before forwarding"** (the number your user dialed that then forwarded to Twilio):
- On regular PSTN call forwarding, **it is not reliably available**—many carriers do not pass the original called number end-to-end.
- Sometimes it can be inferred only if you use **SIP trunks** and the carrier provides headers like `Diversion`/`History-Info`, but you should not depend on that for an MVP.
- Twilio provides `ForwardedFrom` when available - use as fallback, not primary routing.

**Reliable solution (recommended for multi-tenant)**:
- Give each destination/user a **unique Twilio number** (~$1/month per number).
- Route by **`To`** (the Twilio number) → map to tenant's prompt/settings in DB.
- Use `ForwardedFrom` as optional fallback for shared-number scenarios.

---

## Deployment Topology on Coolify
### Services
- **`backend` (Go)**:
  - Exposes:
    - `POST /telephony/inbound` (Twilio webhook)
    - `POST /telephony/status` (optional status callbacks)
    - `GET /healthz`
    - `WS /media` (Twilio Media Streams bidirectional WebSocket)
  - Connects to Postgres (and Redis if used).

- **`frontend` (Vite React TS)**:
  - Static build served by a lightweight web server (e.g., Caddy/Nginx) or by a Node adapter.
  - Calls backend via HTTPS REST for listing calls/transcripts/classifications.

- **`postgres`**:
  - Persistent volume enabled in Coolify.

- **(optional) `redis`**:
  - Only if you add background jobs, caching, rate-limit counters, etc.

### Networking / Domains
- `api.yourdomain.com` → backend
- `app.yourdomain.com` → frontend

**Important for Twilio Media Streams**:
- WebSockets must be supported end-to-end (reverse proxy + TLS).
- Use **`wss://api.yourdomain.com/media`** (TLS required for production).

---

## Coolify Setup (Practical Steps)
### 1) Create a Project
- Create a new Coolify project for this app.

### 2) Add Postgres (and Redis if needed)
- Add a Postgres resource.
- Enable persistent storage/volume.
- Note connection info for env vars.

### 3) Deploy Backend (Go)
Recommended deployment style:
- A Dockerized Go service (multi-stage build).
- Expose HTTP + WebSocket on a single port (e.g., `8080`).

In Coolify:
- Create a new service from your Git repo (Dockerfile or Buildpack).
- Set the service domain to `api.yourdomain.com`.
- Ensure “WebSocket support” is enabled (Coolify’s proxy typically supports it; still validate).

### 4) Deploy Frontend (Vite)
Recommended:
- Build Vite in CI/build step.
- Serve `dist/` via Nginx/Caddy container.

In Coolify:
- Create a second service from the same repo (or separate repo) for the frontend.
- Set domain to `app.yourdomain.com`.
- Configure env var for API base URL (e.g., `VITE_API_BASE_URL=https://api.yourdomain.com`).

### 5) HTTPS
- Enable Let’s Encrypt certificates via Coolify for both domains.
- Twilio webhooks should use **HTTPS** endpoints.

---

## Twilio Configuration
### Phone Numbers
For each Twilio number you buy:
- Set **Voice webhook** (A CALL COMES IN) to:
  - `https://api.yourdomain.com/telephony/inbound` (POST)

### Webhook Security
- Validate Twilio signatures on incoming webhooks (recommended).
- Restrict traffic if possible (WAF / allowlist is optional; signature validation is the key).

### Media Streams
On inbound webhook response, instruct Twilio to start a stream to:
- `wss://api.yourdomain.com/media`

Notes:
- Expect 8kHz μ-law frames commonly; decode/normalize as needed for STT.
- Implement barge-in by interrupting TTS playback when inbound audio resumes.

---

## Environment Variables (Backend)
Minimum set:
- **`DATABASE_URL`**: Postgres connection string
- **`TWILIO_AUTH_TOKEN`**: used to validate webhook signatures
- **`TWILIO_ACCOUNT_SID`**: optional (if you call Twilio APIs)
- **`PUBLIC_BASE_URL`**: e.g. `https://api.yourdomain.com` (for constructing stream URLs)
- **`STT_PROVIDER` / `STT_API_KEY`**: streaming speech-to-text
- **`TTS_PROVIDER` / `TTS_API_KEY`**: text-to-speech (if not using prerecorded prompts)
- **`LLM_API_KEY`**: for classification + short dialog prompts
- **`LOG_LEVEL`**

Optional:
- **`REDIS_URL`**
- **`SENTRY_DSN`** (or other error reporting)

---

## Environment Variables (Frontend)
- **`VITE_API_BASE_URL`**: e.g. `https://api.yourdomain.com`

---

## UI Kit Recommendation (Vite + React + TS)
Any of these work well:
- **Mantine**: fast to ship, good defaults, great data tables/forms ecosystem
- **Chakra UI**: simple, accessible components
- **MUI**: very complete, heavier but powerful

For your use case (call list + filters + detail view), Mantine is often the quickest.

---

## Minimal “App” Screens (Frontend)
- **Call Inbox**:
  - filters: label (legit/marketing/spam/unknown), date range, `From` number
  - columns: time, from, label, intent one-liner, confidence
- **Call Detail**:
  - metadata (from/to, duration)
  - classification + rationale (internal)
  - transcript

---

## Operational Checks (Go Live)
- Confirm `POST /telephony/inbound` reachable from Twilio (HTTP 200).
- Confirm `wss://api.../media` upgrades successfully (101 Switching Protocols).
- Measure:
  - STT latency (time to first partial)
  - LLM time-to-first-token
  - TTS time-to-first-byte
- Confirm DB writes on call end (transcript + screening result).

---

## Infrastructure Alternatives (for Scale/Productization)

### Current: Coolify
- Good for MVP and small scale (< 100 tenants)
- Simple Docker-based deployment
- Manual scaling

### Alternative: Railway
- Better DX, built-in PostgreSQL with backups
- Auto-scaling, native WebSocket support
- $5-20/month for this workload
- Good middle ground between simplicity and scale

### Alternative: Render
- Similar to Railway
- Competitive pricing, auto-scaling
- Good for teams

### Alternative: Fly.io
- Edge deployment = lower latency
- Good for globally distributed users
- More complex setup

### Future: Kubernetes (when needed)
- For 1000+ tenants
- Managed K8s (GKE, EKS, DigitalOcean)
- Horizontal pod autoscaling
- Per-tenant isolation if needed

**Recommendation**: Stay on Coolify for initial productization. Add database-level multi-tenancy first. Migrate to Railway or K8s when you need horizontal scaling (likely at 100+ concurrent calls).

See [docs/PRODUCTIZATION.md](docs/PRODUCTIZATION.md) for the full scaling roadmap.


