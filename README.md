# Zvednu (zvednu.cz)

**"Zvednu to za tebe"** - Czech AI phone assistant that answers calls when you can't.

Single-repo project for a **Twilio-forwarded call screener** that captures **intent + legitimacy label** and stores results in **Postgres**, with a **web app** to review calls.

## Product Vision

**Zvednu** is a **multi-tenant SaaS** phone assistant with AI assistant **Karen**:
- Each user gets their own phone number with a customized AI assistant
- Handles calls when you're unavailable (like a smart voicemail)
- Captures caller intent, screens spam/marketing, forwards VIP calls
- Web app first, mobile apps (Android + iOS) planned

**Documentation:**
- [docs/PRODUCTIZATION.md](docs/PRODUCTIZATION.md) - Product roadmap & architecture
- [docs/UX.md](docs/UX.md) - User flows & screen designs
- [docs/MARKETING_CZ.md](docs/MARKETING_CZ.md) - Czech marketing strategy

## Repo Structure
- `backend/` Go API (Twilio webhooks + WebSocket media endpoint + DB)
- `frontend/` Vite + React + TypeScript admin app
- `deploy/` Docker/Coolify deployment assets

## Quick Start (Local)
1) Copy env files:
- `cp backend/env.example backend/.env`
- `cp frontend/env.example frontend/.env`

2) Start services:

```bash
docker compose up --build
```

3) Open:
- App: `http://localhost:5173`
- API: `http://localhost:8080/healthz`

## Voice AI Configuration

Karen uses a multi-provider voice AI stack for natural Czech conversation:
- **Deepgram Nova-3** for speech-to-text (STT)
- **OpenAI GPT-4o-mini** for conversation intelligence
- **ElevenLabs Flash v2.5** for text-to-speech (TTS)

### Voice Settings

Configure in `backend/.env`:

| Variable | Default | Description |
|----------|---------|-------------|
| `TTS_VOICE_ID` | Rachel | ElevenLabs voice ID |
| `TTS_STABILITY` | 0.5 | Voice stability (0.0-1.0). Lower = more expressive |
| `TTS_SIMILARITY` | 0.75 | Voice similarity boost (0.0-1.0). Higher = closer to original |

### Conversation Features

- **Sentence streaming**: LLM responses are streamed to TTS sentence-by-sentence, reducing latency by ~500ms-1.5s
- **Filler words**: Karen says brief acknowledgments ("Jasně...", "Rozumím...") immediately after user finishes speaking
- **Barge-in support**: Users can interrupt Karen mid-sentence; audio stops immediately and new input is processed
- **Turn detection**: 800ms silence threshold for natural conversation flow (configurable per tenant)
- **Per-tenant customization**: Each tenant can configure greeting, system prompt, voice, and conversation timing

### Per-Tenant Configuration

Each tenant can customize Karen's behavior through the tenant settings API (`PATCH /api/tenant`):

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `greeting_text` | string | (env default) | Custom greeting when answering calls |
| `system_prompt` | string | (built-in) | Custom AI personality and instructions |
| `voice_id` | string | (env default) | ElevenLabs voice ID for TTS |
| `endpointing` | int | 800 | Silence detection threshold in ms (200-2000) |
| `language` | string | cs-CZ | Language code for STT |
| `vip_names` | array | [] | Names to recognize as VIP callers |
| `forward_number` | string | null | Number to forward VIP calls to |

**Endpointing** controls how long Karen waits for silence before considering the user done speaking. Lower values (e.g., 500ms) make conversations faster but may cut off slow speakers. Higher values (e.g., 1200ms) are more forgiving but feel slower.

## Coolify Deployment (Recommended)
Import `deploy/docker-compose.coolify.yml` into Coolify (or create services from the Dockerfiles).

Production domains:
- `api.zvednu.cz` → backend service
- `zvednu.cz` / `www.zvednu.cz` → frontend service

Then configure Twilio:
- Voice webhook for your Twilio number: `https://api.zvednu.cz/telephony/inbound` (POST)
- Media stream URL used by backend: `wss://api.zvednu.cz/media`

See [deploy/SETUP.md](deploy/SETUP.md) for detailed deployment instructions.


