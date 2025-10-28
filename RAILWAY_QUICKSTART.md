# Railway Quick Start Guide

Get your Feishu chatbot with Agno AI backend deployed to Railway in 15 minutes.

## Prerequisites Checklist

- [ ] Railway account ([Sign up here](https://railway.app/))
- [ ] GitHub repository with this code
- [ ] Feishu App created ([Feishu Open Platform](https://open.feishu.cn/app))
- [ ] OpenAI API key ([Get one here](https://platform.openai.com/api-keys))

## Step 1: Gather Your Credentials (5 minutes)

### From Feishu Open Platform

1. Go to https://open.feishu.cn/app
2. Select or create your app
3. Note down these values:
   - **App ID** (starts with `cli_`)
   - **App Secret**
   - **Encrypt Key**
   - **Verification Token**

### From OpenAI

1. Go to https://platform.openai.com/api-keys
2. Create a new API key (starts with `sk-`)
3. Copy and save it securely

## Step 2: Deploy to Railway (5 minutes)

### Deploy Agno Service First

1. Go to [Railway Dashboard](https://railway.app/dashboard)
2. Click **"New Project"** → **"Deploy from GitHub repo"**
3. Authorize Railway and select your repository
4. Click **"Add Service"** → Select your repo
5. Name it: `agno-service`
6. **Settings** → **Service**:
   - Root Directory: `ai-service`
   - Start Command: `uvicorn main:app --host 0.0.0.0 --port $PORT`
7. **Settings** → **Variables**, add:
   ```
   OPENAI_KEY=sk-your_actual_key_here
   OPENAI_MODEL=gpt-4
   HOST=0.0.0.0
   STORAGE_DIR=/app/data
   ```
8. **Settings** → **Networking**:
   - Note the internal URL: `agno-service.railway.internal`

### Deploy Feishu Bot Service

1. In the same project, click **"New Service"**
2. Select your repository again
3. Name it: `feishu-bot`
4. **Settings** → **Service**:
   - Root Directory: (leave empty)
   - Start Command: (leave empty, uses Dockerfile)
5. **Settings** → **Variables**, add:
   ```
   APP_ID=cli_your_app_id
   APP_SECRET=your_app_secret
   APP_ENCRYPT_KEY=your_encrypt_key
   APP_VERIFICATION_TOKEN=your_token
   BOT_NAME=chatGpt
   OPENAI_KEY=sk-your_key
   AGNO_SERVICE_URL=http://agno-service.railway.internal:8000
   USE_AGNO=true
   ```
6. **Settings** → **Networking**:
   - Click **"Generate Domain"**
   - Copy the URL (e.g., `https://feishu-bot-production.up.railway.app`)

## Step 3: Configure Feishu Webhooks (3 minutes)

1. Go back to [Feishu Open Platform](https://open.feishu.cn/app)
2. Select your app
3. **Event Subscriptions** → **Event Configuration**:
   - Request URL: `https://your-railway-url.up.railway.app/webhook/event`
   - Click **Save** and verify
4. **Card Callback**:
   - Request URL: `https://your-railway-url.up.railway.app/webhook/card`
   - Click **Save** and verify
5. **Subscribe to Events** (if not already):
   - `im.message.receive_v1` (Receive messages)
   - `im.message.message_read_v1` (Read receipts)

## Step 4: Test Your Bot (2 minutes)

1. Open Feishu/Lark app
2. Find your bot (search by bot name)
3. Send a message: "Hello!"
4. You should receive an AI-powered response

If it works, congratulations! 🎉 Your bot is live.

## Troubleshooting Quick Fixes

### Bot doesn't respond

**Check Railway logs:**
1. Go to Railway → Your Project → feishu-bot → Logs
2. Look for errors

**Common issues:**
- Environment variables not set correctly
- Webhook URL not saved in Feishu
- Services not communicating

**Quick fix:**
```bash
# Test health endpoints
curl https://your-bot-url.up.railway.app/ping
curl https://your-agno-url.up.railway.app/health
```

### "OpenAI not configured" error

**Fix:**
1. Go to Railway → agno-service → Variables
2. Verify `OPENAI_KEY` is set correctly
3. Redeploy the service

### Services can't communicate

**Fix:**
1. Go to Railway → feishu-bot → Variables
2. Update `AGNO_SERVICE_URL`:
   ```
   http://agno-service.railway.internal:8000
   ```
3. Make sure both services are in the same project

### Webhook verification fails

**Fix:**
1. Verify these match in both Railway and Feishu:
   - `APP_ENCRYPT_KEY`
   - `APP_VERIFICATION_TOKEN`
2. Make sure webhook URL uses HTTPS
3. Check Railway logs for decryption errors

## Environment Variables Quick Reference

### Feishu Bot (Required)

```env
APP_ID=cli_your_app_id
APP_SECRET=your_secret
APP_ENCRYPT_KEY=your_encrypt_key
APP_VERIFICATION_TOKEN=your_token
BOT_NAME=chatGpt
OPENAI_KEY=sk-your_key
AGNO_SERVICE_URL=http://agno-service.railway.internal:8000
USE_AGNO=true
```

### Agno Service (Required)

```env
OPENAI_KEY=sk-your_key
OPENAI_MODEL=gpt-4
HOST=0.0.0.0
STORAGE_DIR=/app/data
```

## Next Steps

Once your bot is working:

1. **Add Custom Domain** (Optional):
   - Railway → Settings → Networking → Custom Domain

2. **Monitor Usage**:
   - Railway → Your Project → Usage tab
   - Set up billing alerts

3. **Customize Bot Behavior**:
   - Edit `ai-service/main.py` to change system prompts
   - Update and redeploy

4. **Add More Features**:
   - Check `CLAUDE.md` for architecture details
   - See `RAILWAY_DEPLOYMENT.md` for advanced config

## Support Resources

- **Full Deployment Guide**: See `RAILWAY_DEPLOYMENT.md`
- **Architecture Details**: See `CLAUDE.md`
- **Railway Docs**: https://docs.railway.app/
- **Feishu API Docs**: https://open.feishu.cn/document/

## Deployment Checklist

Use this before going live:

- [ ] Both services deployed and healthy
- [ ] All environment variables set
- [ ] Webhook URLs configured in Feishu
- [ ] Health checks passing
- [ ] Test message works end-to-end
- [ ] Logs show no errors
- [ ] Domain configured (if using custom domain)

---

**Total Time**: ~15 minutes

Your Feishu bot with AI backend is now live on Railway! 🚀
