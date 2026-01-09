# Architecture: Low‑Latency Call Screener (Redirected Calls → Intent Capture + Legit/Spam Classification → DB/App)

## Goals
- **Fast responses**: minimize time-to-first-word and avoid long “dead air”.
- **High understanding**: robust speech recognition, clarification for critical fields, and concise spoken dialog.
- **Intent collection**: capture **why the caller is calling** + essential details (e.g., name/company/callback).
- **Legitimacy classification**: label calls as **legitimate**, **marketing**, or **spam** (configurable taxonomy).
- **Persistence + visibility**: store **transcript** + **structured intent + classification** and show it in an app/dashboard.
- **Operationally safe**: privacy, auditing, monitoring, and replayability.

## Non‑Goals (for the first iteration)
- Performing real “assistant actions” (booking, cancellations, payments, account changes).
- Deep knowledge base Q&A (RAG) unless later needed.
- Custom acoustic models; rely on strong vendor STT first.
- Complex omnichannel (chat/email) — focus on phone calls.

---

## High‑Level System Overview
Calls are **redirected** (forwarded) to a telephony provider which streams audio to your backend. The backend runs a **real‑time conversational loop**:

1. **Telephony Ingress** receives the call + streams audio.
2. **Speech-to-Text (STT)** transcribes audio in near real time (streaming partials + finals).
3. **Conversation Orchestrator** maintains session state and runs a short “screener” script (ask purpose, clarify, confirm).
4. **Legitimacy Classifier** labels the call (legit/marketing/spam) and extracts structured intent/entities.
5. **Text-to-Speech (TTS)** streams synthesized speech back to the caller (barge‑in supported) or uses pre-recorded prompts.
6. **Persistence Pipeline** stores transcript + classification + extracted intent into DB for the app/dashboard.

**Key design choice for latency**: use **streaming everywhere** (telephony audio, STT partials, LLM streaming tokens, TTS streaming audio) and keep a **single session worker per call** with in-memory state and durable event log.

---

## Core Components

### 1) Telephony Provider (Redirected Calls)
**Purpose**: receive forwarded PSTN/SIP calls and provide a programmable interface.

- Recommended: **Twilio Programmable Voice** (or Vonage/Plivo).
- Features needed:
  - Webhook on call start (call SID, from/to, etc.)
  - **Bidirectional media streaming** (WebSocket) for real-time audio
  - DTMF events (optional)
  - Call recordings (optional, policy-driven)

**Redirect/Forward setup**:
- User’s number forwards to a Twilio number.
- Twilio hits your `POST /telephony/inbound` webhook.
- You respond with instructions to start a media stream to your `wss://.../media`.

### 2) Media Gateway (WebSocket)
**Purpose**: terminate telephony media stream, normalize audio, and publish frames/events internally.

- Stateless service.
- Responsibilities:
  - Validate signature/auth (provider-specific).
  - Decode provider audio (often 8kHz μ-law) → normalize to STT-required format.
  - Forward audio frames to the Session Worker (or message bus).
  - Receive TTS audio and push back to telephony stream.
  - Support **barge-in**: if caller speaks while TTS plays, stop TTS playback and prioritize incoming audio.

### 3) Streaming STT Service
**Purpose**: convert audio → text quickly and accurately.

Options:
- Vendor streaming STT (e.g., Deepgram, Google, Azure, AssemblyAI).
- Must support:
  - streaming partial results (low latency),
  - endpointing/VAD controls,
  - diarization optional (caller vs agent separation can be inferred by channel events).

**Understanding strategy**:
- Use **domain hints** (custom vocabulary, phrases).
- Prefer STT with strong punctuation + numeric normalization.
- Use confidence scores to trigger clarification.

### 4) Conversation Orchestrator (Session Worker)
**Purpose**: the “brain” of each call, but scoped to **screening** (collect intent + classify).

Implementation pattern:
- One logical worker per call (sticky routing).
- Maintains:
  - system prompt + per-customer prompt template
  - conversation history (condensed + recent window)
  - turn-taking state (who is speaking, is TTS playing)
  - tool results + grounding context

Key sub-modules:
- **Turn Manager**:
  - consumes STT partial/final segments
  - determines “end of user turn” (endpointing + heuristics)
  - handles barge-in (interrupts agent output)
- **Prompt Manager**:
  - merges base screener prompt + configured “specified prompt”
  - injects policies (PII, allowed questions, escalation)
  - controls response style (very short, spoken)
- **Screener Script Engine**:
  - asks a small fixed set of questions (purpose + optional details)
  - performs confirmations for critical fields (name/number/company)
  - stops once intent is captured (avoid long conversations)
- **Classifier/Extractor**:
  - produces structured JSON: intent + entities + legitimacy label + confidence + rationale

### 5) LLM Layer (Realtime)
**Purpose**: generate short spoken prompts and produce a **structured intent + legitimacy classification**.

For latency + natural conversation:
- Use a model/API that supports **streaming tokens**.
- Keep responses short and spoken-friendly (avoid paragraphs).

Understanding improvements:
- **Clarification policy**:
  - if low STT confidence or ambiguity → ask one focused clarifying question.
  - confirm critical entities (name/company/callback number).
- **Structured extraction** (always):
  - intent category, free-text reason, entities, urgency, requested next step.
- **Legitimacy classification** (always):
  - legit vs marketing vs spam (+ confidence + short rationale).

### 6) Streaming TTS Service
**Purpose**: convert agent text → speech quickly and naturally.

Requirements:
- streaming audio output
- stable voice, controllable prosody
- ability to stop playback immediately (barge-in)

Option (often simpler/faster for MVP):
- Use **pre-recorded prompts** for the first sentence (“Hi, who is calling and what is this about?”) and only use TTS for dynamic confirmations.

Latency tactics:
- Start TTS as soon as first tokens arrive (“incremental speech”).
- Use short sentences and early acknowledgements (“Okay—got it. One sec.”) while tools run.

### 7) Data Storage (Transcript + Summary)
**Purpose**: durable artifacts for screening + audit:
- transcript
- extracted intent/entities
- legitimacy label (legit/marketing/spam)

Recommended: **PostgreSQL** as system-of-record.

Store:
- call metadata (from/to, timestamps, status, recording refs)
- transcript as an ordered list of utterances (speaker, text, timestamps, confidence)
- raw events (optional but great for debugging: STT partials)
- screening result (structured JSON + short human-readable summary)

Optional:
- Object storage for recordings (S3-compatible) with retention policies.
- Vector DB for knowledge base (pgvector or dedicated service).

### 8) Async Workers (Post-Call)
**Purpose**: do heavier work without affecting call latency.

Examples:
- final “clean” summary + structured extraction (if not done live)
- spam/marketing enrichment (number reputation, known spam lists)
- sentiment / QA scoring (optional)
- analytics aggregation
- redact PII for downstream sharing

Queue: Redis (BullMQ), SQS, or RabbitMQ.

---

## End-to-End Call Flow (Low Latency)

### Inbound Call (Forwarded)
1. Caller dials → call forwarded to Telephony Provider.
2. Provider calls `POST /telephony/inbound` with call metadata.
3. Backend returns instructions to open media stream to `wss://api.yourapp.com/media`.
4. Provider streams audio frames to Media Gateway.
5. Media Gateway forwards frames to STT stream.
6. STT emits partial transcripts continuously.
7. Turn Manager detects end-of-turn → sends text to LLM (screener dialog).
8. LLM produces:
   - the next short spoken prompt (streaming),
   - incremental structured extraction updates (optional).
9. TTS streams audio → Media Gateway streams to provider.
10. On hangup:
   - finalize transcript
   - finalize screening result (intent + legitimacy)
   - persist to DB

### Barge-in / Interrupts
- If caller speaks while agent is speaking:
  - stop TTS stream immediately,
  - mark current agent utterance as interrupted,
  - resume STT prioritization.

---

## “Specified Prompt” Handling
You’ll support per-customer or per-number prompts:

- `PromptTemplate`:
  - identity/role (“You are a helpful phone assistant for …”)
  - goals (**collect reason for call** and **assess legitimacy**)
  - required fields (e.g., name, company, callback, topic)
  - constraints (privacy, allowed questions, escalation rules)
  - style (concise, friendly, confirm important details)
  - completion criteria (when to stop asking questions)

Best practice:
- Keep the system prompt stable.
- Put customer prompt into a separate “instructions” block.
- Add a small “speech style guide” (short, conversational, avoid jargon).

---

## Database Schema (Suggested)

### `calls`
- `id` (uuid, pk)
- `provider` (text)
- `provider_call_id` (text, unique)
- `from_number` (text)
- `to_number` (text)
- `started_at`, `ended_at` (timestamptz)
- `status` (text: in_progress/completed/failed)
- `prompt_id` (uuid, fk)
- `recording_url` (text, nullable)

### `call_utterances`
- `id` (uuid, pk)
- `call_id` (uuid, fk)
- `speaker` (text: caller/agent/system)
- `text` (text)
- `started_at`, `ended_at` (timestamptz, nullable)
- `stt_confidence` (float, nullable)
- `sequence` (int) — stable ordering
- `interrupted` (bool)

### `call_summaries`
- `call_id` (uuid, pk/fk)
- `summary_text` (text)
- `summary_json` (jsonb) — intent + entities + legitimacy label + confidence + rationale
- `created_at` (timestamptz)

Recommended addition:

### `call_screening_results`
- `call_id` (uuid, pk/fk)
- `legitimacy_label` (text: legitimate/marketing/spam/unknown)
- `legitimacy_confidence` (float)
- `intent_category` (text) — configurable taxonomy
- `intent_text` (text) — one-sentence “why they called”
- `entities_json` (jsonb) — name/company/callback/etc
- `needs_follow_up` (bool)
- `created_at` (timestamptz)

### `prompts`
- `id` (uuid, pk)
- `name` (text)
- `template` (text)
- `created_at` (timestamptz)

Optional:
- `call_events` (jsonb append-only) for debugging/replay.

---

## Latency & “Fast Response” Tactics (Most Important)

### Streaming + Early Acknowledgement
- Stream STT partials; don’t wait for full sentences.
- Use **early backchannels** (short confirmations) while tools load.
- Stream LLM tokens; begin TTS once you have a clause, not a paragraph.

### Turn Detection (Avoid Awkward Pauses)
- Combine:
  - STT endpointing (silence duration),
  - punctuation/intent heuristics,
  - max turn length guardrails.

### Keep the “Hot Path” Minimal
On the live call path, do only:
- STT streaming
- short LLM generation
- TTS streaming
- append events to a fast store

Push expensive work to async workers.

### Session Locality
- Sticky route each call to one worker to avoid cross-node chatter.
- Keep recent conversation in memory; periodically snapshot to DB/event log.

### Tooling & RAG Carefully
- Keep external lookups off the hot path for MVP.
- If you later add lookups (caller reputation/CRM), cache aggressively and time-box calls.

---

## “High Ability to Understand” Tactics

### Clarification Policy
If any of the following:
- STT confidence < threshold,
- conflicting entities,
- user request ambiguous,

…then ask **one** targeted question (not multiple).

### Entity Confirmation
Always confirm critical items:
- names, phone numbers, addresses
- dates/times
- amounts

### Domain Vocabulary
Maintain per-customer word/phrase lists:
- product names
- acronyms
- common issues

### Safety & Escalation
Define “handoff” triggers:
- caller requests human
- repeated misunderstandings (e.g., 2 clarifications in a row)
- sensitive categories (medical/legal) depending on policy

---

## Legitimacy Classification (Legit vs Marketing vs Spam)
Use a layered approach for best accuracy:

- **Signals (cheap, fast)**:
  - caller number match against contact list / allowlist / blocklist
  - call frequency patterns (repeat callers)
  - STT keywords (“promotion”, “special offer”, “survey”, etc.)
- **LLM structured classification**:
  - given transcript + metadata, return JSON:
    - `label`: legitimate|marketing|spam|unknown
    - `confidence`: 0..1
    - `rationale`: short (internal)
    - `intent_category`, `intent_text`, `entities`

Important UX policy:
- When uncertain, use `unknown` and ask one clarifying question (“Is this about an existing appointment/order, or is this a promotional call?”).

---

## APIs (Suggested)

### Public (Telephony)
- `POST /telephony/inbound` → returns provider instructions (connect stream)
- `POST /telephony/status` → call state callbacks
- `WS /media` → bidirectional audio + control events

### Internal
- `POST /sessions/{callId}/events` (optional, for replay)
- `POST /summaries/{callId}/finalize` (worker)
- `GET /calls/{callId}` (dashboard)

For the “app view”:
- `GET /screening?from=...&to=...&label=...` (list)
- `GET /screening/{callId}` (detail with transcript + classification)

---

## Deployment Topology

### Minimal (MVP)
- 1 API service (includes Media Gateway + Session Worker)
- 1 Postgres
- 1 Redis (optional but recommended for queues/caching)

### Scaled
- API service (webhooks + auth)
- Media Gateway service (websockets, audio)
- Session Worker service (stateful, autoscaled with sticky routing)
- Async Worker service (summaries, analytics)
- Postgres + Redis + object storage

Observability:
- structured logs with `call_id`
- metrics: STT latency, LLM time-to-first-token, TTS time-to-first-byte, end-to-end “time to first word”
- traces across gateway → STT → LLM → TTS

---

## Security, Privacy, and Compliance
- Encrypt data at rest (Postgres + backups).
- Restrict access via least privilege (service accounts, DB roles).
- Token/signature validation for telephony webhooks.
- PII handling:
  - redact in logs,
  - optionally store redacted transcript separately for analytics,
  - retention policies per customer.
- If recording calls, announce per jurisdiction and store consent flags.

---

## MVP Milestones (Pragmatic Build Order)
- **M1**: Inbound call redirect + “hello” + hangup; store call record.
- **M2**: Streaming STT → LLM → TTS loop; store transcript.
- **M3**: Barge-in support + turn detection tuned for low latency.
- **M4**: Intent extraction + legitimacy classification (structured JSON); store in DB.
- **M5**: Prompt management per number/customer + basic dashboard (list + detail).
- **M6**: Optional enrichment (number reputation) + escalation policy.

---

## Open Questions (to finalize choices)
- Telephony provider preference (Twilio/Vonage/Plivo)?
- Target languages/accents and noise conditions?
- Do you need call recording storage, and what retention period?
- What legitimacy labels do you want beyond `legitimate|marketing|spam|unknown`?
- Should the system **hang up after collecting intent**, or **transfer to you** for legit calls?
- What "specified prompt" format do you want (markdown, JSON, UI builder)?

---

## Productization Roadmap

The goal is to evolve Karen into a **multi-tenant SaaS** phone assistant.

### Target Use Case
Users forward their phone to Karen when unavailable (busy, DND, no answer). Karen:
1. Answers the call with a personalized greeting
2. Captures why the caller is calling
3. Classifies the call (legitimate/marketing/spam)
4. Forwards VIP callers immediately
5. Stores the call summary for later review

### Multi-Tenancy Model
Each tenant (user) has:
- **Dedicated Twilio number**: reliable routing via `To` field
- **Custom system prompt**: personalized assistant personality
- **Custom voice**: ElevenLabs voice ID selection
- **VIP list**: names/numbers to forward immediately
- **Marketing redirect**: email for marketing callers

### Phone Number Routing
**Primary method**: Dedicated Twilio number per tenant
- Twilio number → tenant lookup → apply tenant's config
- Most reliable, ~$1/month per number

**Secondary method**: Call forwarding detection
- Detect `ForwardedFrom` header from Twilio
- Less reliable (carrier-dependent), useful as fallback

### Infrastructure Evolution
1. **MVP (current)**: Coolify, single tenant, hardcoded prompt
2. **Phase 1**: Multi-tenant DB, phone auth (Twilio Verify), routing by number
3. **Phase 2**: Web app (landing, onboarding, dashboard, settings)
4. **Phase 3**: Billing (Stripe), usage metering, horizontal scaling
5. **Phase 4**: Mobile apps (Android + iOS with push notifications)

See [docs/PRODUCTIZATION.md](docs/PRODUCTIZATION.md) for detailed implementation plan.
See [docs/UX.md](docs/UX.md) for user flows and screen designs.


