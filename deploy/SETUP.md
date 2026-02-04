# Zvednu Deployment Setup Guide

## Prerequisites

1. **Server**: VPS with Docker and Traefik at 46.224.75.8
2. **Domain**: zvednu.cz pointed to 46.224.75.8

## DNS Configuration

Add these DNS records to zvednu.cz:

| Type | Name | Value |
|------|------|-------|
| A | @ | 46.224.75.8 |
| A | www | 46.224.75.8 |
| A | api | 46.224.75.8 |

## GitHub Repository Secrets

Go to GitHub repo → Settings → Secrets and variables → Actions, and add:

| Secret Name | Value |
|-------------|-------|
| `SERVER_HOST` | `46.224.75.8` |
| `SERVER_SSH_KEY` | Your SSH private key for root@46.224.75.8 |

## Initial Server Setup

SSH into the server and run:

```bash
# Create deployment directory
mkdir -p /opt/karen

# Create .env file with your credentials
cat > /opt/karen/.env << 'EOF'
# Database
POSTGRES_PASSWORD=your_secure_password_here

# Twilio (required)
TWILIO_AUTH_TOKEN=your_twilio_auth_token_here
TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_VERIFY_SERVICE_SID=VAxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Voice AI Providers (required for calls)
DEEPGRAM_API_KEY=your_deepgram_api_key
OPENAI_API_KEY=your_openai_api_key
ELEVENLABS_API_KEY=your_elevenlabs_api_key

# JWT Authentication (required - generate a secure random string!)
JWT_SECRET=generate-a-secure-random-string-here

# Admin access (comma-separated phone numbers in E.164 format)
ADMIN_PHONES=+420777123456,+420777654321

# Error monitoring (optional)
SENTRY_DSN=

# Hotjar analytics (optional, used during build)
VITE_HOTJAR_ID=
EOF

# Secure the .env file
chmod 600 /opt/karen/.env
```

### Environment Variable Reference

| Variable | Required | Description |
|----------|----------|-------------|
| `POSTGRES_PASSWORD` | Yes | PostgreSQL database password |
| `TWILIO_AUTH_TOKEN` | Yes | Twilio auth token for webhook validation |
| `TWILIO_ACCOUNT_SID` | Yes | Twilio account SID |
| `TWILIO_VERIFY_SERVICE_SID` | Yes | Twilio Verify service for SMS OTP |
| `JWT_SECRET` | Yes | Secret key for JWT signing (generate random!) |
| `DEEPGRAM_API_KEY` | Yes | Deepgram API key for speech-to-text |
| `OPENAI_API_KEY` | Yes | OpenAI API key for LLM |
| `ELEVENLABS_API_KEY` | Yes | ElevenLabs API key for text-to-speech |
| `ADMIN_PHONES` | Recommended | Comma-separated admin phone numbers |
| `SENTRY_DSN` | No | Sentry DSN for error tracking |
| `VITE_HOTJAR_ID` | No | Hotjar Site ID (set in GitHub secrets) |

## Deployment

Deployment happens automatically when you push to `main` branch via GitHub Actions.

To manually trigger a deployment:
1. Go to GitHub repo → Actions → "CI"
2. Click "Run workflow"

## URLs

After deployment:
- **Frontend**: https://zvednu.cz
- **Backend API**: https://api.zvednu.cz
- **Health check**: https://api.zvednu.cz/healthz

## Twilio Configuration

Configure your Twilio phone number with:
- **Voice webhook**: `https://api.zvednu.cz/telephony/inbound` (POST)
- **Status callback**: `https://api.zvednu.cz/telephony/status` (POST)

## Logs

```bash
# View all logs
ssh root@46.224.75.8 "cd /opt/karen && docker compose logs -f"

# View specific service logs
ssh root@46.224.75.8 "docker logs -f karen-backend"
ssh root@46.224.75.8 "docker logs -f karen-frontend"
ssh root@46.224.75.8 "docker logs -f karen-db"
```

## Troubleshooting

```bash
# Check container status
ssh root@46.224.75.8 "cd /opt/karen && docker compose ps"

# Restart services
ssh root@46.224.75.8 "cd /opt/karen && docker compose restart"

# Check Traefik routing (Traefik runs as coolify-proxy from the original Coolify install)
ssh root@46.224.75.8 "docker logs coolify-proxy 2>&1 | tail -50"
```

## Maintenance

No scheduled tasks or cron jobs are currently needed. Notes for future scale:

- **Session cleanup**: Expired JWT sessions could be pruned periodically (currently handled at query time)
- **Event log retention**: `call_events` table may need retention policies at high call volumes
- **Push token cleanup**: Stale APNs device tokens could be removed after delivery failures
- **Billing period resets**: If usage-based billing is added, periodic aggregation jobs would be needed

PostgreSQL autovacuum handles routine database maintenance automatically.
