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

## Coolify Deployment (Recommended)
Import `deploy/docker-compose.coolify.yml` into Coolify (or create services from the Dockerfiles).

Production domains:
- `api.zvednu.cz` → backend service
- `zvednu.cz` / `www.zvednu.cz` → frontend service

Then configure Twilio:
- Voice webhook for your Twilio number: `https://api.zvednu.cz/telephony/inbound` (POST)
- Media stream URL used by backend: `wss://api.zvednu.cz/media`

See [deploy/SETUP.md](deploy/SETUP.md) for detailed deployment instructions.


