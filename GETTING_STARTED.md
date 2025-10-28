# Getting Started with Lark-Agent

Welcome! This guide will help you get your Feishu/Lark chatbot with Agno AI backend up and running.

## What is Lark-Agent?

Lark-Agent is a powerful Feishu (Lark) chatbot that combines:
- **Go backend**: Fast, efficient webhook handler for Feishu events
- **Python AI service**: Intelligent responses using Agno framework and OpenAI
- **Persistent memory**: Conversation context stored in SQLite
- **Railway-ready**: Optimized for cloud deployment

```
User in Feishu → Go Bot (webhook) → Python Agno (AI) → OpenAI → Response
```

## Quick Links

| Document | Purpose | Read Time |
|----------|---------|-----------|
| **[RAILWAY_QUICKSTART.md](RAILWAY_QUICKSTART.md)** | Deploy to Railway in 15 minutes | 5 min |
| **[RAILWAY_DEPLOYMENT.md](RAILWAY_DEPLOYMENT.md)** | Complete Railway deployment guide | 15 min |
| **[DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md)** | Pre-deployment checklist | 10 min |
| **[CLAUDE.md](CLAUDE.md)** | Technical architecture details | 20 min |
| **[readme.md](readme.md)** | Original project README (Chinese) | 10 min |

## Choose Your Path

### Path 1: Quick Deploy to Railway (Recommended)
**Best for**: First-time users who want to get started fast

1. Read [RAILWAY_QUICKSTART.md](RAILWAY_QUICKSTART.md)
2. Follow the 3-step deployment process
3. Test your bot in Feishu

**Time**: ~15 minutes

### Path 2: Local Development First
**Best for**: Developers who want to test locally before deploying

1. Read [Local Development Setup](#local-development-setup) below
2. Test with Docker Compose
3. Deploy using [RAILWAY_DEPLOYMENT.md](RAILWAY_DEPLOYMENT.md)

**Time**: ~30 minutes

### Path 3: Production Deployment
**Best for**: Teams deploying to production

1. Review [DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md)
2. Complete all checklist items
3. Follow [RAILWAY_DEPLOYMENT.md](RAILWAY_DEPLOYMENT.md)
4. Run verification script

**Time**: ~1 hour

## Local Development Setup

### Prerequisites

- Docker and Docker Compose installed
- Feishu app credentials
- OpenAI API key

### Step 1: Clone and Configure

```bash
# Clone repository
git clone https://github.com/your-username/lark-agent.git
cd lark-agent

# Copy environment template
cp .env.example .env

# Edit .env with your credentials
nano .env  # or use your favorite editor
```

### Step 2: Set Environment Variables

Edit `.env` and set these required variables:

```env
# Feishu App Configuration
APP_ID=cli_your_app_id
APP_SECRET=your_app_secret
APP_ENCRYPT_KEY=your_encrypt_key
APP_VERIFICATION_TOKEN=your_verification_token
BOT_NAME=chatGpt

# OpenAI Configuration
OPENAI_KEY=sk-your_openai_key_here
OPENAI_MODEL=gpt-4

# Agno Service URL (for local: localhost)
AGNO_SERVICE_URL=http://agno-service:8000
USE_AGNO=true
```

### Step 3: Start Services

```bash
# Start both services
docker-compose up -d

# View logs
docker-compose logs -f

# Check health
curl http://localhost:9000/ping
curl http://localhost:8000/health
```

### Step 4: Test Locally (Optional)

To test with Feishu locally, you'll need to expose your local server:

```bash
# Install ngrok (https://ngrok.com)
ngrok http 9000

# Use the ngrok URL in Feishu webhook configuration
# Example: https://abc123.ngrok.io/webhook/event
```

## Repository Structure

```
lark-agent/
├── code/                           # Go Feishu bot service
│   ├── handlers/                   # Event and card handlers
│   ├── services/                   # OpenAI, session, cache services
│   ├── initialization/             # Config, Lark client, server setup
│   ├── utils/                      # Utility functions
│   └── main.go                     # Entry point
│
├── ai-service/                     # Python Agno AI service
│   ├── main.py                     # FastAPI application
│   ├── requirements.txt            # Python dependencies
│   ├── Dockerfile                  # Python service Docker image
│   ├── README.md                   # Agno service documentation
│   ├── QUICKSTART.md              # 5-minute setup guide
│   ├── INTEGRATION_GUIDE.md       # Go-Python integration
│   └── go-client-example/         # Go HTTP client for Agno
│
├── Dockerfile                      # Go bot Docker image
├── docker-compose.yaml             # Multi-service orchestration
├── railway.json                    # Railway configuration
│
├── GETTING_STARTED.md             # This file
├── RAILWAY_QUICKSTART.md          # 15-minute Railway setup
├── RAILWAY_DEPLOYMENT.md          # Complete deployment guide
├── DEPLOYMENT_CHECKLIST.md        # Pre-deployment checklist
├── CLAUDE.md                       # Architecture documentation
│
├── .env.example                    # Environment variables template
├── verify-deployment.sh            # Bash verification script
├── verify-deployment.ps1           # PowerShell verification script
│
└── readme.md                       # Original README (Chinese)
```

## Architecture Overview

### Two-Service Architecture

**Service 1: Feishu Bot (Go)**
- **Location**: `/code`
- **Port**: 9000
- **Purpose**: Handle Feishu webhooks, message routing
- **Tech**: Go 1.18, Gin, Larksuite SDK

**Service 2: Agno AI (Python)**
- **Location**: `/ai-service`
- **Port**: 8000
- **Purpose**: AI-powered responses with memory
- **Tech**: Python 3.11, FastAPI, Agno, OpenAI

### Communication Flow

```
┌─────────────┐
│   Feishu    │
│   (User)    │
└──────┬──────┘
       │ HTTPS Webhook
       ▼
┌─────────────────┐
│  Go Feishu Bot  │  Port 9000
│  (Webhook)      │  ← Public
└────────┬────────┘
         │ HTTP Request
         ▼
┌─────────────────┐
│ Python Agno AI  │  Port 8000
│  (AI Service)   │  ← Internal
└────────┬────────┘
         │ API Call
         ▼
┌─────────────────┐
│   OpenAI API    │
└─────────────────┘
         │ AI Response
         ▼
    Back to User
```

## Environment Variables

See [.env.example](.env.example) for complete list.

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `APP_ID` | Feishu App ID | `cli_a1234567890abcde` |
| `APP_SECRET` | Feishu App Secret | `your_secret_here` |
| `APP_ENCRYPT_KEY` | Feishu Encrypt Key | `your_encrypt_key` |
| `APP_VERIFICATION_TOKEN` | Feishu Verification Token | `your_token` |
| `OPENAI_KEY` | OpenAI API Key(s) | `sk-xxx,sk-yyy` |
| `AGNO_SERVICE_URL` | Agno service URL | `http://agno-service:8000` |

## Verification

After deployment, verify everything works:

### Using Bash (Linux/Mac)
```bash
export FEISHU_BOT_URL=https://your-bot.up.railway.app
export AGNO_SERVICE_URL=https://your-agno.up.railway.app
bash verify-deployment.sh
```

### Using PowerShell (Windows)
```powershell
.\verify-deployment.ps1 `
  -FeishuBotUrl https://your-bot.up.railway.app `
  -AgnoServiceUrl https://your-agno.up.railway.app
```

## Common Tasks

### View Logs (Docker Compose)
```bash
docker-compose logs -f feishu-bot
docker-compose logs -f agno-service
```

### View Logs (Railway)
```bash
railway logs --service feishu-bot
railway logs --service agno-service
```

### Restart Services (Docker Compose)
```bash
docker-compose restart
```

### Update and Redeploy (Railway)
```bash
# Railway auto-deploys on git push
git add .
git commit -m "Update configuration"
git push origin main
```

### Clear Conversation History
Send `/clear` command to the bot in Feishu

## Troubleshooting

### Bot doesn't respond

1. **Check logs**:
   ```bash
   docker-compose logs -f
   # or
   railway logs
   ```

2. **Verify webhooks**:
   - Feishu webhook URL must be HTTPS
   - URL format: `https://your-domain/webhook/event`

3. **Test health**:
   ```bash
   curl https://your-bot-url/ping
   curl https://your-agno-url/health
   ```

### "OpenAI not configured" error

- Verify `OPENAI_KEY` is set in both services
- Check API key is valid and has credits
- Ensure API key starts with `sk-`

### Services can't communicate

- For local: Use `http://agno-service:8000`
- For Railway: Use `http://agno-service.railway.internal:8000`
- Verify both services are in same Docker network/Railway project

### More troubleshooting

See [RAILWAY_DEPLOYMENT.md#troubleshooting](RAILWAY_DEPLOYMENT.md#troubleshooting) for detailed solutions.

## Next Steps

1. **Deploy to Railway**: Follow [RAILWAY_QUICKSTART.md](RAILWAY_QUICKSTART.md)
2. **Customize AI behavior**: Edit `ai-service/main.py`
3. **Add features**: See [CLAUDE.md](CLAUDE.md) for architecture
4. **Production deployment**: Use [DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md)

## Resources

### Documentation
- [Feishu Open Platform](https://open.feishu.cn/document/)
- [Railway Documentation](https://docs.railway.app/)
- [Agno Framework](https://docs.agno.com/) (if available)
- [OpenAI API](https://platform.openai.com/docs/)

### Support
- **GitHub Issues**: Report bugs and request features
- **Railway Community**: https://discord.gg/railway
- **Feishu Developer Forum**: https://open.feishu.cn/community/

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

See [LICENSE](LICENSE) file for details.

---

**Ready to get started?**

Choose your path above and start building your AI-powered Feishu chatbot!

For fastest deployment: [RAILWAY_QUICKSTART.md](RAILWAY_QUICKSTART.md) →
