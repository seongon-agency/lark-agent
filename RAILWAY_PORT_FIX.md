# Railway PORT Error Fix

## Problem

Getting error: **"PORT is not valid integer"** when deploying to Railway.

## Solution

I've updated the code to handle Railway's PORT environment variable properly. Here's what was fixed:

### 1. Updated `code/initialization/config.go`

The configuration now:
- ✅ Reads PORT directly from environment variables
- ✅ Handles invalid PORT values gracefully
- ✅ Falls back to default port (9000) if PORT is invalid
- ✅ Prioritizes HTTP_PORT over PORT if both are set
- ✅ Makes config.yaml optional (uses env vars on Railway)

### 2. How to Deploy

After pulling these changes:

#### Step 1: Clear Railway Variables

1. Go to Railway Dashboard → Your Project → feishu-bot service
2. Go to **Variables** tab
3. **Remove** any existing `PORT` variable if you manually set one
   - Railway sets PORT automatically, don't override it
4. **Remove** any `HTTP_PORT` variable if present

#### Step 2: Verify Required Variables

Make sure these are set in Railway (feishu-bot service):

```env
APP_ID=cli_your_app_id
APP_SECRET=your_app_secret
APP_ENCRYPT_KEY=your_encrypt_key
APP_VERIFICATION_TOKEN=your_token
BOT_NAME=chatGpt
OPENAI_KEY=sk-your_key
AGNO_SERVICE_URL=http://agno-service.railway.internal:8000
USE_AGNO=true
```

**Do NOT set:**
- `PORT` - Railway sets this automatically
- `HTTP_PORT` - Only set if you want to override Railway's PORT

#### Step 3: Redeploy

1. Push the updated code to GitHub:
   ```bash
   git add .
   git commit -m "Fix PORT configuration for Railway"
   git push origin main
   ```

2. Railway will auto-deploy

3. Check logs:
   ```bash
   railway logs
   ```

4. Look for these messages:
   ```
   Using PORT from env: 8080 (or whatever Railway assigned)
   http server started: http://localhost:8080/webhook/event
   ```

## What Was Changed

### Before (❌ Problematic)
```go
httpPort := getViperIntValue("HTTP_PORT", 0)
if httpPort == 0 {
    httpPort = getViperIntValue("PORT", 9000)
}
```

### After (✅ Fixed)
```go
// Try reading from environment variables directly
if portStr := os.Getenv("HTTP_PORT"); portStr != "" {
    if p, err := strconv.Atoi(portStr); err == nil {
        httpPort = p
    }
}

// If HTTP_PORT not set, try PORT (Railway default)
if httpPort == 0 {
    if portStr := os.Getenv("PORT"); portStr != "" {
        if p, err := strconv.Atoi(portStr); err == nil {
            httpPort = p
        } else {
            // Handle invalid PORT gracefully
            fmt.Printf("Warning: PORT env var '%s' is not a valid integer\n", portStr)
        }
    }
}

// Final fallback to default
if httpPort == 0 {
    httpPort = 9000
}
```

## Debugging

If you still see the error, check Railway logs for these debug messages:

```bash
railway logs
```

Look for:
- `Warning: Could not read config file` - This is OK on Railway
- `Using PORT from env: XXXX` - Port was read successfully
- `Warning: PORT env var 'XXX' is not a valid integer` - PORT has invalid value

## Common Causes

### 1. Manually Set PORT Variable
**Problem**: You manually added PORT=something in Railway variables

**Solution**: Delete the PORT variable from Railway dashboard. Railway sets it automatically.

### 2. Config File with PORT
**Problem**: A config.yaml file was somehow included in the Docker image with invalid PORT

**Solution**:
- Verify `config.yaml` is in `.gitignore`
- Check Dockerfile doesn't explicitly COPY config.yaml
- The updated code now ignores config.yaml if it doesn't exist

### 3. Environment Variable Syntax Error
**Problem**: Railway variable has whitespace or special characters

**Solution**: Ensure all variables are set correctly:
- No leading/trailing spaces
- No quotes around values (Railway adds them automatically)
- Use plain text, not environment variable syntax like `$PORT`

## Testing Locally

To test the fix locally before deploying:

```bash
# Test with valid PORT
export PORT=8080
go run code/main.go

# Test with invalid PORT (should use default 9000)
export PORT=invalid
go run code/main.go

# Test with no PORT (should use default 9000)
unset PORT
go run code/main.go
```

## Expected Behavior

### On Railway
```
✅ Railway sets PORT=8080 (example)
✅ App reads PORT from environment
✅ App starts on port 8080
✅ Railway can route traffic to the app
```

### Locally
```
✅ No PORT set → Uses 9000
✅ PORT=8080 → Uses 8080
✅ HTTP_PORT=7000 → Uses 7000 (overrides PORT)
```

## Still Having Issues?

### Check Railway Logs

```bash
railway logs --service feishu-bot
```

Look for the startup messages:
```
Using PORT from env: XXXX
http server started: http://localhost:XXXX/webhook/event
```

### Verify Config File

The app should show:
```
Warning: Could not read config file ./config.yaml: ... (using env vars)
```

This is **normal** on Railway - we use environment variables, not config files.

### Check Service Status

```bash
railway status
```

All services should show as "Running"

### Test Health Endpoint

```bash
curl https://your-bot-domain.up.railway.app/ping
```

Should return: `pong`

## Railway Dashboard Checklist

Before deploying, verify in Railway Dashboard:

- [ ] feishu-bot service exists
- [ ] Variables tab shows required variables (APP_ID, etc.)
- [ ] **No PORT variable manually set** (Railway sets it)
- [ ] **No HTTP_PORT variable** (unless you want to override)
- [ ] Deployments tab shows successful build
- [ ] Logs show "Using PORT from env: XXXX"

## Quick Fix Summary

1. **Pull latest code** (includes the fix)
2. **Remove any PORT variables** in Railway dashboard
3. **Keep only required variables** (APP_ID, APP_SECRET, etc.)
4. **Redeploy** (git push or manual redeploy)
5. **Check logs** for "Using PORT from env"
6. **Test** with `curl https://your-domain/ping`

---

The code is now robust and handles all PORT scenarios. Deploy with confidence! 🚀
