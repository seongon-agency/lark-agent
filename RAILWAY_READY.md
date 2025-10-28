# ✅ Railway Deployment - Configuration Complete

Your Lark-Agent repository is now **100% ready for Railway deployment**!

## Recent Fix: Railway PORT Error ✅

**Fixed**: "PORT is not valid integer" error on Railway deployment

The Go service now:
- ✅ Reads PORT directly from environment variables
- ✅ Handles invalid PORT values gracefully
- ✅ Makes config.yaml optional (uses env vars on Railway)
- ✅ Falls back to default port (9000) safely

**See**: [RAILWAY_PORT_FIX.md](RAILWAY_PORT_FIX.md) for details

**Important**: Do NOT manually set PORT variable in Railway - it's set automatically!

---

## What Was Configured

### 1. Railway Configuration Files ✓
- ✅ `railway.json` - Root service configuration
- ✅ `ai-service/railway.json` - Python service configuration
- ✅ `railway.toml` - Reference file

### 2. Docker Configuration ✓
- ✅ `Dockerfile` - Go bot multi-stage build (optimized for Railway)
- ✅ `ai-service/Dockerfile` - Python Agno service (Railway PORT compatible)
- ✅ `docker-compose.yaml` - Updated with both services for local testing
- ✅ Health checks configured for both services

### 3. Go Service Updates ✓
- ✅ `code/initialization/config.go` - Updated to support Railway's PORT variable
- ✅ Fallback logic: Checks `HTTP_PORT` first, then `PORT`, defaults to 9000
- ✅ Compatible with both local and Railway deployment

### 4. Environment Configuration ✓
- ✅ `.env.example` - Comprehensive template with all variables
- ✅ Separate sections for Feishu, OpenAI, Agno, Azure
- ✅ Detailed comments and Railway deployment notes
- ✅ `.gitignore` already configured to exclude `.env`

### 5. Documentation ✓
- ✅ `GETTING_STARTED.md` - Main navigation guide
- ✅ `RAILWAY_QUICKSTART.md` - 15-minute deployment guide
- ✅ `RAILWAY_DEPLOYMENT.md` - Complete deployment reference
- ✅ `DEPLOYMENT_CHECKLIST.md` - Pre-deployment checklist
- ✅ `CLAUDE.md` - Architecture documentation (existing)

### 6. Verification Tools ✓
- ✅ `verify-deployment.sh` - Bash verification script (Linux/Mac)
- ✅ `verify-deployment.ps1` - PowerShell verification script (Windows)
- ✅ Health check endpoint testing
- ✅ Service communication verification

## Quick Start

### For First-Time Deployment

1. **Read this first**: [RAILWAY_QUICKSTART.md](RAILWAY_QUICKSTART.md)
2. **Follow the 3 steps**:
   - Gather credentials (5 min)
   - Deploy to Railway (5 min)
   - Configure webhooks (3 min)
3. **Test your bot** in Feishu

**Total time**: ~15 minutes

### For Production Deployment

1. **Review**: [DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md)
2. **Complete all checks**
3. **Deploy using**: [RAILWAY_DEPLOYMENT.md](RAILWAY_DEPLOYMENT.md)
4. **Verify with**: `verify-deployment.sh` or `verify-deployment.ps1`

## File Summary

| File | Purpose | Required |
|------|---------|----------|
| `railway.json` | Railway root service config | Yes |
| `ai-service/railway.json` | Railway Python service config | Yes |
| `Dockerfile` | Go bot Docker image | Yes |
| `ai-service/Dockerfile` | Python Agno Docker image | Yes |
| `docker-compose.yaml` | Local multi-service testing | Optional |
| `.env.example` | Environment template | Yes |
| `GETTING_STARTED.md` | Main navigation | Recommended |
| `RAILWAY_QUICKSTART.md` | Quick deploy guide | Recommended |
| `RAILWAY_DEPLOYMENT.md` | Complete reference | Recommended |
| `DEPLOYMENT_CHECKLIST.md` | Pre-deploy checklist | Recommended |
| `verify-deployment.sh` | Verification (Bash) | Optional |
| `verify-deployment.ps1` | Verification (PowerShell) | Optional |

## What You Need

Before deploying, gather these:

### From Feishu Open Platform
- [ ] App ID (starts with `cli_`)
- [ ] App Secret
- [ ] Encrypt Key
- [ ] Verification Token

**Get them here**: https://open.feishu.cn/app

### From OpenAI
- [ ] API Key (starts with `sk-`)
- [ ] Ensure it has available credits

**Get it here**: https://platform.openai.com/api-keys

### From Railway
- [ ] Account created
- [ ] GitHub connected

**Sign up here**: https://railway.app/

## Deployment Steps (Summary)

### Step 1: Deploy Agno Service
```
Railway → New Project → Deploy from GitHub
→ Add Service → Name: agno-service
→ Root: ai-service
→ Set env vars: OPENAI_KEY, OPENAI_MODEL
```

### Step 2: Deploy Feishu Bot
```
Same project → Add Service → Name: feishu-bot
→ Root: (empty)
→ Set env vars: APP_ID, APP_SECRET, OPENAI_KEY, AGNO_SERVICE_URL
→ Generate public domain
```

### Step 3: Configure Webhooks
```
Feishu → Your App → Event Subscriptions
→ Event URL: https://your-railway-domain/webhook/event
→ Card URL: https://your-railway-domain/webhook/card
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Feishu Cloud                        │
│                   (User Messages)                       │
└────────────────────┬────────────────────────────────────┘
                     │ HTTPS Webhook
                     ▼
┌─────────────────────────────────────────────────────────┐
│                  Railway Project                        │
│                                                          │
│  ┌──────────────────┐         ┌──────────────────┐     │
│  │  feishu-bot      │         │  agno-service    │     │
│  │  (Go Service)    │────────▶│  (Python/Agno)   │     │
│  │  Port: 9000      │  HTTP   │  Port: 8000      │     │
│  │  Public: ✓       │         │  Internal: ✓     │     │
│  └──────────────────┘         └────────┬─────────┘     │
│         │                               │               │
└─────────┼───────────────────────────────┼───────────────┘
          │                               │
          │                               ▼
          │                      ┌──────────────────┐
          │                      │   OpenAI API     │
          │                      └──────────────────┘
          │                               │
          └───────────────────────────────┘
                     Response to User
```

## Service Communication

### Local Development
```env
AGNO_SERVICE_URL=http://agno-service:8000
```

### Railway Deployment
```env
AGNO_SERVICE_URL=http://agno-service.railway.internal:8000
```

## Testing

### Local Testing with Docker Compose
```bash
# 1. Copy environment file
cp .env.example .env

# 2. Edit .env with your credentials
nano .env

# 3. Start services
docker-compose up -d

# 4. Check health
curl http://localhost:9000/ping
curl http://localhost:8000/health

# 5. View logs
docker-compose logs -f
```

### Railway Testing
```bash
# Set your Railway URLs
export FEISHU_BOT_URL=https://feishu-bot-production.up.railway.app
export AGNO_SERVICE_URL=https://agno-service-production.up.railway.app

# Run verification
bash verify-deployment.sh

# Or on Windows
.\verify-deployment.ps1 `
  -FeishuBotUrl https://feishu-bot-production.up.railway.app `
  -AgnoServiceUrl https://agno-service-production.up.railway.app
```

## Environment Variables Checklist

### Feishu Bot Service

Required:
- [ ] `APP_ID`
- [ ] `APP_SECRET`
- [ ] `APP_ENCRYPT_KEY`
- [ ] `APP_VERIFICATION_TOKEN`
- [ ] `BOT_NAME`
- [ ] `OPENAI_KEY`
- [ ] `AGNO_SERVICE_URL`

Optional:
- [ ] `OPENAI_MODEL` (default: gpt-3.5-turbo)
- [ ] `USE_AGNO` (default: true)
- [ ] `API_URL` (default: https://api.openai.com)

### Agno Service

Required:
- [ ] `OPENAI_KEY`
- [ ] `OPENAI_MODEL`

Optional:
- [ ] `HOST` (default: 0.0.0.0)
- [ ] `PORT` (Railway sets automatically)
- [ ] `STORAGE_DIR` (default: /app/data)

## Health Check Endpoints

| Service | Endpoint | Expected Response |
|---------|----------|-------------------|
| Feishu Bot | `GET /ping` | `pong` |
| Agno Service | `GET /health` | `{"status":"healthy",...}` |

## Troubleshooting Quick Reference

| Problem | Solution |
|---------|----------|
| Bot not responding | Check Railway logs, verify webhooks |
| OpenAI not configured | Set `OPENAI_KEY` in both services |
| Services can't talk | Use `.railway.internal` URL |
| Webhook fails | Verify `APP_ENCRYPT_KEY` matches |
| Port errors | Railway sets `PORT` automatically |

**Full troubleshooting**: See [RAILWAY_DEPLOYMENT.md](RAILWAY_DEPLOYMENT.md#troubleshooting)

## Next Steps

1. ✅ Configuration complete
2. 📖 Read [RAILWAY_QUICKSTART.md](RAILWAY_QUICKSTART.md)
3. 🚀 Deploy to Railway (15 minutes)
4. ✅ Run verification script
5. 🎉 Start using your bot!

## Support & Resources

- **Quick Start**: [RAILWAY_QUICKSTART.md](RAILWAY_QUICKSTART.md)
- **Full Guide**: [RAILWAY_DEPLOYMENT.md](RAILWAY_DEPLOYMENT.md)
- **Checklist**: [DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md)
- **Architecture**: [CLAUDE.md](CLAUDE.md)
- **Railway Docs**: https://docs.railway.app/
- **Feishu Docs**: https://open.feishu.cn/document/

## Deployment Confidence

✅ **Docker images optimized**
✅ **Railway configuration complete**
✅ **Environment variables documented**
✅ **Health checks configured**
✅ **Verification tools ready**
✅ **Documentation complete**
✅ **Service communication configured**

## You're All Set! 🎉

Your repository is now **100% ready** for Railway deployment.

**Start here**: [RAILWAY_QUICKSTART.md](RAILWAY_QUICKSTART.md)

---

**Configuration completed by**: Claude Code
**Date**: 2025-10-28
**Status**: ✅ Ready for Railway
