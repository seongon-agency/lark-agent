# Railway Deployment Guide

This guide will walk you through deploying your Lark-Agent (Feishu chatbot with Agno AI backend) to Railway.

## Architecture Overview

This project consists of two services that work together:

1. **feishu-bot** (Go): Main webhook handler for Feishu/Lark events
2. **agno-service** (Python): AI backend using Agno framework for intelligent responses

```
Feishu → feishu-bot (Go) → agno-service (Python) → OpenAI
```

## Prerequisites

- [Railway account](https://railway.app/) (free tier available)
- [Railway CLI](https://docs.railway.app/develop/cli) (optional, but recommended)
- Feishu/Lark App credentials (from [Feishu Open Platform](https://open.feishu.cn/app))
- OpenAI API key

## Deployment Steps

### Option 1: Deploy via Railway Dashboard (Recommended for Beginners)

#### Step 1: Create a New Project

1. Log in to [Railway](https://railway.app/)
2. Click **"New Project"**
3. Select **"Deploy from GitHub repo"**
4. Authorize Railway to access your GitHub account
5. Select your `lark-agent` repository

#### Step 2: Create the Agno Service

1. Railway will detect your repository
2. Click **"Add a Service"**
3. Select **"Add Service"** → Choose your repository
4. Name it: `agno-service`
5. Go to **Settings** → **Service**:
   - **Root Directory**: Set to `ai-service`
   - **Build Command**: Leave empty (uses Dockerfile)
   - **Start Command**: `uvicorn main:app --host 0.0.0.0 --port $PORT`

6. Go to **Settings** → **Environment Variables** and add:
   ```
   OPENAI_KEY=sk-your_actual_openai_key
   OPENAI_MODEL=gpt-4
   HOST=0.0.0.0
   STORAGE_DIR=/app/data
   ```

7. Go to **Settings** → **Networking**:
   - **Generate Domain** (optional, for external access)
   - Note the **Private Network** URL: `agno-service.railway.internal`

8. Go to **Settings** → **Health Check**:
   - **Path**: `/health`
   - **Timeout**: 100 seconds

#### Step 3: Create the Feishu Bot Service

1. Click **"New Service"** in your project
2. Select your repository again
3. Name it: `feishu-bot`
4. Go to **Settings** → **Service**:
   - **Root Directory**: Leave empty (uses root)
   - **Build Command**: Leave empty (uses Dockerfile)
   - **Start Command**: Leave empty (Dockerfile defines it)

5. Go to **Settings** → **Environment Variables** and add:

   **Feishu Configuration:**
   ```
   APP_ID=cli_your_app_id
   APP_SECRET=your_app_secret
   APP_ENCRYPT_KEY=your_encrypt_key
   APP_VERIFICATION_TOKEN=your_verification_token
   BOT_NAME=chatGpt
   ```

   **OpenAI Configuration:**
   ```
   OPENAI_KEY=sk-your_openai_key
   OPENAI_MODEL=gpt-3.5-turbo
   OPENAI_MAX_TOKENS=2000
   ```

   **Agno Integration:**
   ```
   AGNO_SERVICE_URL=http://agno-service.railway.internal:8000
   USE_AGNO=true
   ```

   **Optional Settings:**
   ```
   API_URL=https://api.openai.com
   STREAM_MODE=false
   AZURE_ON=false
   ```

6. Go to **Settings** → **Networking**:
   - **Generate Domain** - This gives you a public URL
   - Copy this URL (you'll need it for Feishu webhook configuration)
   - Example: `https://feishu-bot-production.up.railway.app`

7. Go to **Settings** → **Health Check**:
   - **Path**: `/ping`
   - **Timeout**: 100 seconds

#### Step 4: Configure Service Communication

The services need to communicate with each other:

1. In **feishu-bot** service, verify `AGNO_SERVICE_URL` is set to:
   ```
   http://agno-service.railway.internal:8000
   ```

2. Alternatively, you can use Railway's **Service Variables**:
   - Go to feishu-bot → Variables
   - Reference: `${{agno-service.RAILWAY_PRIVATE_DOMAIN}}`

#### Step 5: Configure Feishu Webhooks

1. Go to [Feishu Open Platform](https://open.feishu.cn/app)
2. Select your app
3. Navigate to **Event Subscriptions** → **Event Configuration**
4. Set **Request URL** to: `https://your-railway-domain.up.railway.app/webhook/event`
5. Navigate to **Card Callback**
6. Set **Request URL** to: `https://your-railway-domain.up.railway.app/webhook/card`
7. Save and verify the configuration

#### Step 6: Test Your Deployment

1. Check service logs in Railway dashboard:
   - **agno-service**: Should show "Starting Agno AI Agent Service"
   - **feishu-bot**: Should show "http server started"

2. Test health endpoints:
   ```bash
   curl https://your-feishu-bot-domain.up.railway.app/ping
   # Should return: pong

   curl https://your-agno-service-domain.up.railway.app/health
   # Should return: {"status":"healthy",...}
   ```

3. Test in Feishu:
   - Open your Feishu app
   - Send a message to your bot
   - You should receive an AI-powered response

---

### Option 2: Deploy via Railway CLI (Advanced)

#### Prerequisites

Install Railway CLI:
```bash
npm i -g @railway/cli
# or
brew install railway
```

#### Step 1: Login to Railway

```bash
railway login
```

#### Step 2: Initialize Project

```bash
cd lark-agent
railway init
# Select "Create a new project"
# Name it: lark-agent
```

#### Step 3: Link to Project

```bash
railway link
```

#### Step 4: Deploy Agno Service

```bash
cd ai-service
railway up
# Railway will detect Dockerfile and deploy

# Set environment variables
railway variables set OPENAI_KEY=sk-your_key
railway variables set OPENAI_MODEL=gpt-4
railway variables set STORAGE_DIR=/app/data
```

#### Step 5: Deploy Feishu Bot

```bash
cd ..
railway up
# Railway will detect root Dockerfile and deploy

# Set environment variables
railway variables set APP_ID=cli_your_id
railway variables set APP_SECRET=your_secret
railway variables set APP_ENCRYPT_KEY=your_key
railway variables set APP_VERIFICATION_TOKEN=your_token
railway variables set OPENAI_KEY=sk-your_key
railway variables set AGNO_SERVICE_URL=http://agno-service.railway.internal:8000
railway variables set USE_AGNO=true
```

#### Step 6: Get Deployment URL

```bash
railway domain
# Copy the URL for Feishu webhook configuration
```

---

## Environment Variables Reference

### Required for Feishu Bot

| Variable | Description | Example |
|----------|-------------|---------|
| `APP_ID` | Feishu App ID | `cli_a1234567890abcde` |
| `APP_SECRET` | Feishu App Secret | `your_secret` |
| `APP_ENCRYPT_KEY` | Feishu Encrypt Key | `your_encrypt_key` |
| `APP_VERIFICATION_TOKEN` | Feishu Verification Token | `your_token` |
| `BOT_NAME` | Bot display name | `chatGpt` |
| `OPENAI_KEY` | OpenAI API Key | `sk-xxx,sk-yyy` |
| `AGNO_SERVICE_URL` | Agno service URL | `http://agno-service.railway.internal:8000` |

### Required for Agno Service

| Variable | Description | Example |
|----------|-------------|---------|
| `OPENAI_KEY` | OpenAI API Key | `sk-your_key` |
| `OPENAI_MODEL` | OpenAI Model | `gpt-4` |
| `STORAGE_DIR` | SQLite storage path | `/app/data` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENAI_MODEL` | Model for bot | `gpt-3.5-turbo` |
| `OPENAI_MAX_TOKENS` | Max response tokens | `2000` |
| `API_URL` | OpenAI API URL | `https://api.openai.com` |
| `STREAM_MODE` | Enable streaming | `false` |
| `USE_AGNO` | Use Agno backend | `true` |
| `HTTP_PROXY` | HTTP proxy URL | `` |
| `AZURE_ON` | Use Azure OpenAI | `false` |

---

## Troubleshooting

### Service Won't Start

**Check logs:**
```bash
railway logs
# or in dashboard: Service → Logs
```

**Common issues:**
- Missing environment variables
- Invalid Dockerfile path
- Port conflicts

### Services Can't Communicate

**Verify internal networking:**
- Use `.railway.internal` domain for internal communication
- Check both services are in the same project
- Verify `AGNO_SERVICE_URL` is correct

**Test connectivity:**
```bash
# From feishu-bot service
railway run curl http://agno-service.railway.internal:8000/health
```

### Feishu Webhook Fails

**Verify webhook URL:**
- Must be publicly accessible
- Must use HTTPS (Railway provides this)
- Format: `https://your-domain.up.railway.app/webhook/event`

**Check Railway domain:**
```bash
railway domain
```

**Test endpoint:**
```bash
curl https://your-domain.up.railway.app/ping
```

### Database Issues (Agno Service)

**SQLite storage:**
- Railway provides ephemeral storage by default
- For persistent storage, use Railway Volumes (currently in beta)
- Alternative: Use Railway's PostgreSQL plugin and update Agno storage

**Add volume (if available):**
1. Go to agno-service → Settings → Volumes
2. Add volume: `/app/data`

### High Memory Usage

**Python service optimization:**
- Set `OPENAI_HTTP_CLIENT_TIMEOUT` to lower value
- Limit concurrent requests
- Use lighter model (gpt-3.5-turbo instead of gpt-4)

**Go service optimization:**
- Already optimized with multi-stage build
- Consider reducing `OPENAI_MAX_TOKENS`

### OpenAI API Errors

**Rate limiting:**
- Use multiple OpenAI keys: `OPENAI_KEY=sk-1,sk-2,sk-3`
- Go bot has built-in load balancing

**Timeout errors:**
- Increase `OPENAI_HTTP_CLIENT_TIMEOUT`
- Check `API_URL` is correct
- Verify `HTTP_PROXY` if using one

---

## Monitoring

### View Logs

**Dashboard:**
- Go to your service → **Logs** tab
- Real-time streaming logs
- Filter by log level

**CLI:**
```bash
railway logs
# or
railway logs --follow
```

### Metrics

**Railway provides:**
- CPU usage
- Memory usage
- Network traffic
- Request count

**Access metrics:**
- Dashboard → Service → **Metrics** tab

### Health Checks

**Endpoints:**
- Feishu bot: `GET /ping` → Returns "pong"
- Agno service: `GET /health` → Returns JSON health status

**Test health:**
```bash
# Feishu bot
curl https://your-bot-domain.up.railway.app/ping

# Agno service
curl https://your-agno-domain.up.railway.app/health
```

---

## Scaling

### Vertical Scaling (Increase Resources)

1. Go to Service → **Settings** → **Resources**
2. Upgrade to higher tier for:
   - More RAM
   - More CPU
   - Faster builds

### Horizontal Scaling (Multiple Instances)

**Note:** Current implementation uses in-memory cache
- Not recommended without Redis
- Would require code changes for distributed cache

**For production scaling:**
1. Migrate session cache to Redis
2. Use Railway's Redis plugin
3. Update both services to use Redis
4. Enable multiple replicas

---

## Cost Optimization

### Free Tier Limits

Railway free tier includes:
- $5/month credit
- Unlimited projects
- Limited execution hours

### Tips to Stay Within Free Tier

1. **Use sleep mode:**
   - Services sleep after 30 min inactivity
   - Wake on first request

2. **Optimize Docker images:**
   - Multi-stage builds (already implemented)
   - Minimal base images (already using alpine/slim)

3. **Reduce resource usage:**
   - Use gpt-3.5-turbo instead of gpt-4
   - Lower `OPENAI_MAX_TOKENS`

4. **Monitor usage:**
   - Dashboard → **Usage** tab
   - Set up billing alerts

---

## Production Checklist

Before going to production, ensure:

- [ ] All environment variables are set correctly
- [ ] Webhook URLs are configured in Feishu
- [ ] Health checks are passing
- [ ] Services can communicate internally
- [ ] OpenAI API key is valid and has credits
- [ ] Logs show no errors
- [ ] Test message works end-to-end
- [ ] Domain is configured (optional but recommended)
- [ ] Monitoring is set up
- [ ] Backup plan for database (if using volumes)

---

## Next Steps

1. **Custom Domain (Optional):**
   - Railway → Settings → Networking → Custom Domain
   - Add your domain and configure DNS

2. **Add Database (Optional):**
   - Railway → New → Database → PostgreSQL
   - Update Agno service to use PostgreSQL instead of SQLite

3. **CI/CD (Automatic):**
   - Railway auto-deploys on git push
   - Configure via Settings → GitHub Repo

4. **Monitoring (Advanced):**
   - Integrate with Sentry for error tracking
   - Use Railway webhooks for notifications

---

## Support

- **Railway Docs**: https://docs.railway.app/
- **Feishu Docs**: https://open.feishu.cn/document/
- **Agno Docs**: https://docs.agno.com/ (if available)
- **Project Issues**: File issues on your GitHub repository

---

## Quick Reference

### Railway Commands

```bash
# Login
railway login

# Initialize project
railway init

# Deploy
railway up

# Set variable
railway variables set KEY=value

# View logs
railway logs

# Open dashboard
railway open

# Run command in Railway environment
railway run <command>
```

### Important URLs

- **Railway Dashboard**: https://railway.app/dashboard
- **Feishu Open Platform**: https://open.feishu.cn/app
- **Project Repository**: Your GitHub repo URL

---

**Deployment Status:** ✅ Ready for Railway

Your project is now fully configured for Railway deployment. Follow the steps above to get your Feishu bot with Agno AI backend live!
