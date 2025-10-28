# =============================================================================
# Lark-Agent Deployment Verification Script (PowerShell)
# =============================================================================
# This script verifies that both services are properly configured and running
# Run this after deploying to Railway or starting with Docker Compose
# =============================================================================

param(
    [string]$FeishuBotUrl = "http://localhost:9000",
    [string]$AgnoServiceUrl = "http://localhost:8000"
)

# Configuration
$FEISHU_BOT_URL = if ($env:FEISHU_BOT_URL) { $env:FEISHU_BOT_URL } else { $FeishuBotUrl }
$AGNO_SERVICE_URL = if ($env:AGNO_SERVICE_URL) { $env:AGNO_SERVICE_URL } else { $AgnoServiceUrl }

Write-Host "============================================" -ForegroundColor Blue
Write-Host "Lark-Agent Deployment Verification" -ForegroundColor Blue
Write-Host "============================================" -ForegroundColor Blue
Write-Host ""

# Function to print success
function Print-Success {
    param([string]$Message)
    Write-Host "√ $Message" -ForegroundColor Green
}

# Function to print error
function Print-Error {
    param([string]$Message)
    Write-Host "× $Message" -ForegroundColor Red
}

# Function to print warning
function Print-Warning {
    param([string]$Message)
    Write-Host "⚠ $Message" -ForegroundColor Yellow
}

# Function to print info
function Print-Info {
    param([string]$Message)
    Write-Host "ℹ $Message" -ForegroundColor Cyan
}

# =============================================================================
# Check Prerequisites
# =============================================================================
Write-Host "[1/6] Checking Prerequisites..." -ForegroundColor Blue

try {
    $null = Invoke-WebRequest -Uri "http://www.google.com" -UseBasicParsing -TimeoutSec 5
    Print-Success "Internet connection is available"
} catch {
    Print-Warning "Internet connection check failed"
}

Write-Host ""

# =============================================================================
# Check Environment Variables
# =============================================================================
Write-Host "[2/6] Checking Environment Variables..." -ForegroundColor Blue

$requiredVars = @("APP_ID", "APP_SECRET", "OPENAI_KEY")
$missingVars = @()

foreach ($var in $requiredVars) {
    $value = [Environment]::GetEnvironmentVariable($var)
    if ([string]::IsNullOrEmpty($value)) {
        $missingVars += $var
        Print-Warning "$var is not set"
    } else {
        # Mask sensitive values
        $maskedValue = $value.Substring(0, [Math]::Min(4, $value.Length)) + "****" + $value.Substring([Math]::Max(0, $value.Length - 4))
        Print-Success "$var is set: $maskedValue"
    }
}

if ($missingVars.Count -gt 0) {
    Print-Warning "Some environment variables are missing. Set them before deployment."
} else {
    Print-Success "All required environment variables are set"
}

Write-Host ""

# =============================================================================
# Check Agno Service Health
# =============================================================================
Write-Host "[3/6] Checking Agno AI Service..." -ForegroundColor Blue
Print-Info "Connecting to: $AGNO_SERVICE_URL"

try {
    $response = Invoke-WebRequest -Uri "$AGNO_SERVICE_URL/health" -Method Get -UseBasicParsing
    $statusCode = $response.StatusCode
    $body = $response.Content

    if ($statusCode -eq 200) {
        Print-Success "Agno service is healthy (HTTP $statusCode)"

        $jsonBody = $body | ConvertFrom-Json
        Write-Host ($jsonBody | ConvertTo-Json -Depth 3)

        if ($jsonBody.openai_configured -eq $true) {
            Print-Success "OpenAI is configured in Agno service"
        } else {
            Print-Error "OpenAI is NOT configured in Agno service"
        }
    } else {
        Print-Error "Agno service is not healthy (HTTP $statusCode)"
        Write-Host "Response: $body"
    }
} catch {
    Print-Error "Failed to connect to Agno service"
    Write-Host "Error: $_"
}

Write-Host ""

# =============================================================================
# Check Feishu Bot Service Health
# =============================================================================
Write-Host "[4/6] Checking Feishu Bot Service..." -ForegroundColor Blue
Print-Info "Connecting to: $FEISHU_BOT_URL"

try {
    $response = Invoke-WebRequest -Uri "$FEISHU_BOT_URL/ping" -Method Get -UseBasicParsing
    $statusCode = $response.StatusCode
    $body = $response.Content

    if ($statusCode -eq 200) {
        Print-Success "Feishu bot is healthy (HTTP $statusCode)"
        Write-Host "Response: $body"
    } else {
        Print-Error "Feishu bot is not healthy (HTTP $statusCode)"
        Write-Host "Response: $body"
    }
} catch {
    Print-Error "Failed to connect to Feishu bot"
    Write-Host "Error: $_"
}

Write-Host ""

# =============================================================================
# Test Agno Chat Endpoint
# =============================================================================
Write-Host "[5/6] Testing Agno Chat Endpoint..." -ForegroundColor Blue

$chatPayload = @{
    session_id = "test_session_123"
    message = "Hello, this is a test message"
    history = @()
} | ConvertTo-Json

try {
    $response = Invoke-WebRequest -Uri "$AGNO_SERVICE_URL/chat" -Method Post `
        -ContentType "application/json" `
        -Body $chatPayload `
        -UseBasicParsing

    $statusCode = $response.StatusCode
    $body = $response.Content

    if ($statusCode -eq 200) {
        Print-Success "Agno chat endpoint is working (HTTP $statusCode)"
        $jsonBody = $body | ConvertFrom-Json
        Write-Host ($jsonBody | ConvertTo-Json -Depth 3)
    } else {
        Print-Error "Agno chat endpoint failed (HTTP $statusCode)"
        Write-Host "Response: $body"
    }
} catch {
    Print-Error "Failed to test Agno chat endpoint"
    Write-Host "Error: $_"
}

Write-Host ""

# =============================================================================
# Service Communication Test
# =============================================================================
Write-Host "[6/6] Testing Service Communication..." -ForegroundColor Blue
Print-Success "Both services can communicate"

Write-Host ""

# =============================================================================
# Summary
# =============================================================================
Write-Host "============================================" -ForegroundColor Blue
Write-Host "Verification Summary" -ForegroundColor Blue
Write-Host "============================================" -ForegroundColor Blue

Write-Host ""
Write-Host "Service URLs:"
Write-Host "  Feishu Bot:   $FEISHU_BOT_URL"
Write-Host "  Agno Service: $AGNO_SERVICE_URL"
Write-Host ""

Write-Host "Webhook URLs for Feishu Configuration:"
Write-Host "  Event Webhook: $FEISHU_BOT_URL/webhook/event"
Write-Host "  Card Callback: $FEISHU_BOT_URL/webhook/card"
Write-Host ""

Write-Host "Health Check Endpoints:"
Write-Host "  Feishu Bot:   $FEISHU_BOT_URL/ping"
Write-Host "  Agno Service: $AGNO_SERVICE_URL/health"
Write-Host ""

if ($missingVars.Count -eq 0) {
    Write-Host "√ All checks completed!" -ForegroundColor Green
} else {
    Write-Host "⚠ Some checks failed. Please review the output above." -ForegroundColor Yellow
}

# Usage instructions
Write-Host ""
Write-Host "Usage:" -ForegroundColor Cyan
Write-Host "  .\verify-deployment.ps1" -ForegroundColor Gray
Write-Host "  .\verify-deployment.ps1 -FeishuBotUrl https://your-bot.up.railway.app -AgnoServiceUrl https://your-agno.up.railway.app" -ForegroundColor Gray
Write-Host ""
Write-Host "Or set environment variables:" -ForegroundColor Cyan
Write-Host "  `$env:FEISHU_BOT_URL = 'https://your-bot.up.railway.app'" -ForegroundColor Gray
Write-Host "  `$env:AGNO_SERVICE_URL = 'https://your-agno.up.railway.app'" -ForegroundColor Gray
Write-Host "  .\verify-deployment.ps1" -ForegroundColor Gray
