# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based Feishu/Lark bot that integrates with OpenAI services (GPT-4, GPT-3.5, DALL-E-3, Whisper, GPT-4V). It provides conversational AI, image generation, image analysis, and voice transcription capabilities within Feishu messaging platform.

**Module name:** `start-feishubot`
**Go version:** 1.18

## Build & Run Commands

### Local Development
```bash
cd code

# Copy and configure settings
cp config.example.yaml config.yaml
# Edit config.yaml with your API keys and settings

# Run locally
go run main.go

# Build binary
go build -o feishu_chatgpt main.go

# Run tests
go test ./services/openai -v
go test ./utils -v
```

### Docker
```bash
# Build image
docker build -t feishu-chatgpt:latest .

# Run container (adjust environment variables)
docker run -d --name feishu-chatgpt -p 9000:9000 \
  --env APP_ID=xxx \
  --env APP_SECRET=xxx \
  --env APP_ENCRYPT_KEY=xxx \
  --env APP_VERIFICATION_TOKEN=xxx \
  --env BOT_NAME=chatGpt \
  --env OPENAI_KEY="sk-xxx1,sk-xxx2" \
  --env API_URL="https://api.openai.com" \
  feishu-chatgpt:latest

# Or use docker-compose
docker compose up -d
```

### Webhook Endpoints
After deployment, configure these in Feishu bot backend:
- **Event callback:** `http://YOUR_DOMAIN:9000/webhook/event`
- **Card callback:** `http://YOUR_DOMAIN:9000/webhook/card`
- **Health check:** `http://YOUR_DOMAIN:9000/ping`

## Architecture

### Event-Driven Chain of Responsibility Pattern

The application uses a **chain-of-responsibility pattern** where incoming messages flow through a series of action handlers. Each handler decides whether to process the message and stop propagation or pass to the next handler.

**Main entry point:** `main.go:18-53`
- Initializes role list, config, Lark client, and OpenAI service
- Sets up event dispatcher for message and read receipts
- Sets up card action handler for interactive buttons
- Starts Gin HTTP server on port 9000 (or configured port)

### Message Processing Flow

```
Webhook → Event Dispatcher → MessageHandler → Action Chain → Service Layer → OpenAI → Response
```

**Action chain** (`handlers/handler.go:94-110`):
1. `ProcessedUniqueAction` - Deduplicate via msgCache
2. `ProcessMentionAction` - Verify bot @mention in groups
3. `AudioAction` - Voice-to-text (Whisper)
4. `ClearAction` - `/clear` command handling
5. `VisionAction` - Image analysis (GPT-4V)
6. `PicAction` - Image generation (DALL-E-3)
7. `AIModeAction` - `/ai_mode` temperature selection
8. `RoleListAction` - List available role templates
9. `HelpAction` - Show help information
10. `BalanceAction` - Query OpenAI token balance
11. `RolePlayAction` - System prompt role-playing
12. `MessageAction` - Standard GPT chat completion
13. `EmptyAction` - Handle empty messages

Each action returns `bool`: `false` stops chain, `true` continues.

### Key Services

#### SessionCache (`services/sessionCache.go`)
**Purpose:** Maintains per-conversation state with 12-hour TTL.

**SessionMeta structure:**
- `Mode` - Operating mode: `gpt`, `pic_create`, `pic_vary`, `vision`
- `Msg` - Message history (pruned to stay under 4096 tokens)
- `PicSetting` - Image generation settings (resolution, style)
- `AIMode` - Temperature levels: `Fresh` (0.1), `Warmth` (0.7), `Balance` (1.2), `Creativity` (1.7)
- `VisionDetail` - Image analysis detail: `high` or `low`

**Key methods:**
- `Get(sessionId)` - Retrieve session state
- `SetMsg()` - Append message to history with token pruning
- `SetMode()` - Switch between gpt/pic_create/pic_vary/vision
- `Clear()` - Reset conversation context

#### MsgCache (`services/msgCache.go`)
**Purpose:** Prevent duplicate message processing (30-minute TTL).

#### LoadBalancer (`services/loadbalancer/loadbalancer.go`)
**Purpose:** Distribute API calls across multiple OpenAI keys.

**Features:**
- Round-robin to least-used available key
- Marks keys unavailable on failure
- Auto-revival if all keys unavailable
- Thread-safe with RWMutex

### OpenAI Integration (`services/openai/`)

**ChatGPT service** (`common.go`) supports:
- OpenAI native API
- Azure OpenAI (toggle via `AZURE_ON` config)

**API methods:**
- `Completions(msg, aiMode)` - GPT chat with temperature control
- `GetVisionInfo(visionMsg)` - GPT-4V image analysis
- `AudioToText(audioFile)` - Whisper transcription
- `GenerateOneImage(prompt, size, style)` - DALL-E-3 creation
- `GenerateOneImageVariation(imagePath, resolution)` - Image variants
- `QueryBalance()` - Check account balance

**Retry logic:** Max 3 attempts with exponential backoff (`doAPIRequestWithRetry`)

### Configuration (`initialization/config.go`)

Configuration loaded via Viper from `config.yaml` + environment variables (env vars override YAML).

**Critical settings:**
- `OPENAI_KEY` - Comma-separated keys for load balancing
- `OPENAI_MODEL` - Default: `gpt-3.5-turbo`
- `OPENAI_MAX_TOKENS` - Default: 2000
- `STREAM_MODE` - Enable streaming responses (default: false)
- `HTTP_PROXY` - Optional proxy for API calls
- `API_URL` - Custom endpoint for reverse proxy (default: `https://api.openai.com`)

**Azure settings** (when `AZURE_ON=true`):
- `AZURE_API_VERSION`
- `AZURE_RESOURCE_NAME`
- `AZURE_DEPLOYMENT_NAME`
- `AZURE_OPENAI_TOKEN`

### Response Generation (`handlers/msg.go`)

The bot uses **Feishu interactive cards** for rich UI:
- `newSendCard()` - Full card with header + elements
- `replyCard()` - Reply with interactive card
- `PatchCard()` - Update existing card in-place

**Card elements:**
- Headers with color templates (Blue, Indigo, Green, Grey)
- Markdown/plain text fields
- Image divs with preview
- Action buttons with payloads
- Dropdown menus for selections

## Session Modes

Each conversation operates in one of four modes:

| Mode | Trigger | Behavior |
|------|---------|----------|
| `gpt` | Default | Standard chat with GPT models |
| `pic_create` | `/picture` command | DALL-E-3 image generation from text |
| `pic_vary` | User submits image in pic mode | DALL-E-3 creates variations |
| `vision` | `/vision` command | GPT-4V analyzes uploaded images |

Mode is stored in SessionCache and persists across messages until changed.

## Important Implementation Notes

### Message Context Management
- History is pruned to stay under 4096 tokens using `tokenizer-go`
- Prunes from oldest messages first (preserves system prompt)
- Token counting happens in `sessionCache.SetMsg()` (lines 163-186)

### Duplicate Message Prevention
- All messages tagged in `msgCache` before processing
- 30-minute cache window prevents double-processing
- Implemented in `ProcessedUniqueAction` (first in chain)

### Group Chat vs Private Chat
- Group chats require bot @mention (validated in `ProcessMentionAction`)
- Private chats bypass mention check
- Chat type determined from `event.Message.ChatType`

### Audio Processing
- Feishu audio format: OGG Opus
- Conversion pipeline: OGG → decode Opus → encode WAV/MP3
- Code: `utils/audio/ogg.go` and `utils/audio/wav.go`
- Whisper API requires WAV or MP3

### Image Handling
- Vision mode accepts both single images (`image` msgType) and multiple images (`post` msgType)
- Images uploaded to Feishu first, then passed as URLs to GPT-4V
- DALL-E generates images which are uploaded to Feishu, then embedded in cards

### Role Templates (`role_list.yaml`)
System prompts for role-playing scenarios (translator, interviewer, etc.). Loaded at startup via `initialization.InitRoleList()`.

## File Structure

```
code/
├── main.go                    # Entry point, server setup
├── config.example.yaml        # Configuration template
├── role_list.yaml            # Role-play system prompts
├── handlers/                 # Event and card handlers
│   ├── handler.go            # Action chain definition
│   ├── init.go               # Handler initialization
│   ├── msg.go                # Card/message builders
│   ├── event_*_action.go     # Event action implementations
│   └── card_*_action.go      # Card action implementations
├── services/                 # Business logic layer
│   ├── sessionCache.go       # Conversation state
│   ├── msgCache.go           # Deduplication
│   ├── loadbalancer/         # API key distribution
│   └── openai/               # OpenAI API client
│       ├── common.go         # Core ChatGPT service
│       ├── gpt3.go           # Chat completions
│       ├── vision.go         # GPT-4V integration
│       ├── picture.go        # DALL-E image gen
│       ├── audio.go          # Whisper transcription
│       └── stream.go         # Streaming responses
├── initialization/           # Startup initialization
│   ├── config.go             # Viper config loading
│   ├── lark_client.go        # Feishu SDK client
│   ├── roles_load.go         # Role template loading
│   └── gin.go                # HTTP server setup
├── utils/                    # Utility functions
│   ├── strings.go            # String processing
│   └── audio/                # Audio conversion
└── logger/                   # Logging utilities
```

## Adding New Actions

To add a new message handler:

1. Create action file in `handlers/` (e.g., `event_newfeature_action.go`)
2. Implement `Action` interface:
   ```go
   type NewFeatureAction struct{ /* fields */ }

   func (a *NewFeatureAction) Execute(info *ActionInfo) bool {
       // Return false to stop chain, true to continue
   }
   ```
3. Add to chain in `handlers/handler.go` (lines 94-110):
   ```go
   msgHandler.actionHandlers = []ActionHandler{
       // ... existing handlers
       &NewFeatureAction{},
       // ... remaining handlers
   }
   ```
4. Position matters: earlier actions run first

## Testing

Tests are located alongside source files with `_test.go` suffix:
- `services/openai/gpt3_test.go` - OpenAI integration tests
- `utils/strings_test.go` - String utility tests

Run with: `go test ./... -v`

## Deployment Considerations

### Single-Instance Limitation
- SessionCache and MsgCache use in-memory storage (go-cache)
- Not suitable for multi-instance deployments without shared state store
- For scaling: migrate to Redis or similar distributed cache

### API Key Management
- Multiple keys enable load balancing and higher rate limits
- Keys separated by commas: `sk-xxx1,sk-xxx2,sk-xxx3`
- Failed keys marked unavailable but not permanently removed

### Webhook Security
- Feishu events verified via `APP_VERIFICATION_TOKEN` and `APP_ENCRYPT_KEY`
- Handled automatically by Larksuite SDK middleware
- Ensure these match Feishu bot backend configuration

### Proxy Support
- Set `HTTP_PROXY` for network-restricted environments
- Example: `http://127.0.0.1:7890`
- Leave empty if no proxy needed

## Troubleshooting

### Common Issues

**Bot doesn't respond in groups:**
- Verify bot is @mentioned
- Check `ProcessMentionAction` isn't blocking
- Ensure bot has `im:message.group_at_msg:readonly` permission in Feishu backend

**OpenAI API errors:**
- Check LoadBalancer has valid keys (`initialization/config.go`)
- Verify `API_URL` is correct (default: `https://api.openai.com`)
- Test with `/balance` command to verify connectivity

**Duplicate message processing:**
- MsgCache should prevent this automatically
- Check `ProcessedUniqueAction` is first in chain
- Verify msgCache TTL isn't expired (30 minutes default)

**Context not maintained:**
- SessionCache expires after 12 hours
- User can manually clear with `/clear` command
- Check `sessionCache.SetMsg()` for token pruning issues

**Audio transcription fails:**
- Ensure audio conversion pipeline works (OGG→WAV)
- Whisper API requires audio files < 25MB
- Check `utils/audio/` conversion logic
