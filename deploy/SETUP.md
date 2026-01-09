# Karen Deployment Setup Guide

## Prerequisites

1. **Server**: Coolify running on 46.224.75.8 (coolify.mechant.cz)
2. **Domain**: karen.coollamer.com pointed to 46.224.75.8

## DNS Configuration

Add these DNS records to coollamer.com:

| Type | Name | Value |
|------|------|-------|
| A | karen | 46.224.75.8 |
| A | api.karen | 46.224.75.8 |

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
POSTGRES_PASSWORD=your_secure_password_here
TWILIO_AUTH_TOKEN=your_twilio_auth_token_here
EOF

# Secure the .env file
chmod 600 /opt/karen/.env
```

## Deployment

Deployment happens automatically when you push to `main` branch via GitHub Actions.

To manually trigger a deployment:
1. Go to GitHub repo → Actions → "Deploy to Production"
2. Click "Run workflow"

## URLs

After deployment:
- **Frontend**: https://karen.coollamer.com
- **Backend API**: https://api.karen.coollamer.com
- **Health check**: https://api.karen.coollamer.com/healthz

## Twilio Configuration

Configure your Twilio phone number with:
- **Voice webhook**: `https://api.karen.coollamer.com/telephony/inbound` (POST)
- **Status callback**: `https://api.karen.coollamer.com/telephony/status` (POST)

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

# Check Traefik routing
ssh root@46.224.75.8 "docker logs coolify-proxy 2>&1 | tail -50"
```
