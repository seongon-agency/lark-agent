# Railway Deployment Checklist

Use this checklist to ensure your Lark-Agent is 100% ready for Railway deployment.

## Pre-Deployment Checklist

### 1. Repository Setup

- [ ] Code is in GitHub repository
- [ ] Repository is public or Railway has access
- [ ] All sensitive files are in `.gitignore`:
  - `.env` files
  - `config.yaml`
  - `*.pem` certificates
- [ ] Latest changes are committed and pushed

### 2. Feishu/Lark App Configuration

- [ ] App created in [Feishu Open Platform](https://open.feishu.cn/app)
- [ ] App ID obtained (starts with `cli_`)
- [ ] App Secret obtained
- [ ] Encrypt Key obtained
- [ ] Verification Token obtained
- [ ] Bot name configured
- [ ] Required permissions granted:
  - [ ] `im:message` - Send and receive messages
  - [ ] `im:message.group_at_msg` - Receive @ mentions in groups
  - [ ] `im:resource` - Upload images/files
- [ ] App is published or in development mode

### 3. OpenAI Configuration

- [ ] OpenAI account created
- [ ] API key obtained (starts with `sk-`)
- [ ] API key has available credits
- [ ] Model access verified (gpt-3.5-turbo or gpt-4)
- [ ] (Optional) Multiple keys for load balancing

### 4. Railway Account Setup

- [ ] Railway account created
- [ ] GitHub account connected to Railway
- [ ] Payment method added (required after free tier)
- [ ] Understood Railway pricing model

### 5. Docker Configuration

- [ ] `Dockerfile` (root) - Go bot service ✓
- [ ] `ai-service/Dockerfile` - Python Agno service ✓
- [ ] `docker-compose.yaml` updated with both services ✓
- [ ] Both Dockerfiles tested locally (optional but recommended)

### 6. Environment Variables

#### Feishu Bot Service (feishu-bot)

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
- [ ] `OPENAI_MAX_TOKENS` (default: 2000)
- [ ] `API_URL` (default: https://api.openai.com)
- [ ] `USE_AGNO` (default: true)
- [ ] `STREAM_MODE` (default: false)
- [ ] `HTTP_PROXY` (if needed)
- [ ] Azure OpenAI variables (if using Azure)

#### Agno Service (agno-service)

Required:
- [ ] `OPENAI_KEY`
- [ ] `OPENAI_MODEL`
- [ ] `STORAGE_DIR`

Optional:
- [ ] `HOST` (default: 0.0.0.0)
- [ ] `PORT` (Railway sets automatically)

### 7. Railway Service Configuration

#### Agno Service

- [ ] Service name: `agno-service`
- [ ] Root directory: `ai-service`
- [ ] Dockerfile path: `ai-service/Dockerfile`
- [ ] Start command: `uvicorn main:app --host 0.0.0.0 --port $PORT`
- [ ] Health check path: `/health`
- [ ] Health check timeout: 100s
- [ ] All environment variables set
- [ ] Internal networking enabled

#### Feishu Bot Service

- [ ] Service name: `feishu-bot`
- [ ] Root directory: (empty / root)
- [ ] Dockerfile path: `Dockerfile`
- [ ] Start command: (uses Dockerfile CMD)
- [ ] Health check path: `/ping`
- [ ] Health check timeout: 100s
- [ ] All environment variables set
- [ ] Public domain generated
- [ ] `AGNO_SERVICE_URL` points to internal service

### 8. Networking Configuration

- [ ] Agno service accessible internally
- [ ] Feishu bot has public domain
- [ ] Public domain copied for webhook configuration
- [ ] HTTPS is enabled (Railway default)
- [ ] (Optional) Custom domain configured

### 9. Local Testing (Recommended)

Before deploying to Railway, test locally:

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your credentials
# Then run:
docker-compose up

# Test health endpoints
curl http://localhost:9000/ping
curl http://localhost:8000/health

# Check logs
docker-compose logs -f
```

- [ ] Docker Compose starts without errors
- [ ] Both services are healthy
- [ ] Services can communicate
- [ ] (Optional) Test with Feishu webhook locally using ngrok

### 10. Deployment Verification

After deploying to Railway:

- [ ] Both services deployed successfully
- [ ] No errors in deployment logs
- [ ] Health checks passing
- [ ] Public URLs accessible
- [ ] Environment variables loaded correctly

Use verification script:
```bash
export FEISHU_BOT_URL=https://your-bot.up.railway.app
export AGNO_SERVICE_URL=https://your-agno.up.railway.app
bash verify-deployment.sh
```

### 11. Feishu Webhook Configuration

- [ ] Event webhook URL configured:
  - URL: `https://your-railway-url.up.railway.app/webhook/event`
  - Verified successfully
- [ ] Card callback URL configured:
  - URL: `https://your-railway-url.up.railway.app/webhook/card`
  - Verified successfully
- [ ] Event subscriptions enabled:
  - `im.message.receive_v1`
  - `im.message.message_read_v1`

### 12. End-to-End Testing

- [ ] Bot appears in Feishu
- [ ] Can send message to bot in private chat
- [ ] Bot responds with AI-generated message
- [ ] Can @ mention bot in group chat
- [ ] Bot responds in group
- [ ] Conversation context is maintained
- [ ] `/clear` command works
- [ ] Response time is acceptable (< 10 seconds)

### 13. Monitoring Setup

- [ ] Railway logs accessible
- [ ] Health check endpoints monitored
- [ ] (Optional) Uptime monitoring set up
- [ ] (Optional) Error tracking (Sentry, etc.)
- [ ] (Optional) Usage alerts configured

### 14. Documentation

- [ ] Team members have access to:
  - Environment variables
  - Railway project
  - Feishu app console
  - OpenAI API keys
- [ ] Deployment procedure documented
- [ ] Troubleshooting guide available
- [ ] Contact information for support

### 15. Production Readiness

- [ ] All tests passing
- [ ] No errors in logs for 24 hours
- [ ] Performance is acceptable
- [ ] Costs are within budget
- [ ] Backup plan in place
- [ ] Rollback procedure documented
- [ ] On-call schedule defined (if applicable)

## Post-Deployment

### Week 1

- [ ] Monitor logs daily
- [ ] Track usage metrics
- [ ] Gather user feedback
- [ ] Fix any critical bugs
- [ ] Optimize performance if needed

### Month 1

- [ ] Review Railway costs
- [ ] Analyze usage patterns
- [ ] Consider scaling options
- [ ] Plan feature additions
- [ ] Update documentation

## Quick Reference

### Essential URLs

- **Railway Dashboard**: https://railway.app/dashboard
- **Feishu Open Platform**: https://open.feishu.cn/app
- **OpenAI Platform**: https://platform.openai.com

### Health Check Endpoints

```bash
# Feishu Bot
curl https://your-bot-url.up.railway.app/ping

# Agno Service
curl https://your-agno-url.up.railway.app/health
```

### Railway CLI Commands

```bash
# View logs
railway logs

# List services
railway status

# Set variable
railway variables set KEY=value

# Deploy manually
railway up

# Open dashboard
railway open
```

### Common Issues

| Issue | Quick Fix |
|-------|-----------|
| Bot not responding | Check Railway logs for errors |
| "OpenAI not configured" | Verify OPENAI_KEY in both services |
| Webhook verification fails | Check APP_ENCRYPT_KEY matches Feishu |
| Services can't communicate | Use internal URL: `http://agno-service.railway.internal:8000` |
| Out of memory | Upgrade Railway plan or optimize code |

## Sign-Off

Before going to production, get sign-off from:

- [ ] **Technical Lead**: Architecture and code review
- [ ] **DevOps**: Deployment and monitoring setup
- [ ] **Product Owner**: Feature completeness
- [ ] **QA**: Testing complete
- [ ] **Security**: Security review passed

---

**Deployment Status**:

- [ ] Ready for Railway deployment
- [ ] Deployed to Railway
- [ ] Production ready

**Last Updated**: _________

**Deployed By**: _________

**Production URL**: _________
