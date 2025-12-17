# karen (Call Screener)

Single-repo project for a **Twilio-forwarded call screener** that captures **intent + legitimacy label** and stores results in **Postgres**, with an **admin web app** to review calls.

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

You’ll need:
- `api.yourdomain.com` → backend service
- `app.yourdomain.com` → frontend service

Then configure Twilio:
- Voice webhook for your Twilio number: `https://api.yourdomain.com/telephony/inbound` (POST)
- Media stream URL used by backend: `wss://api.yourdomain.com/media`


