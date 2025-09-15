# ============================================================
# Nexus Commerce — Rebuild & Reseed Script
# Rebuilds backend, runs migrations, and reseeds the database
# ============================================================

param(
    [switch]$Full,       # Also rebuild gateway + frontend
    [switch]$SkipSeed    # Skip seeding (just rebuild + migrate)
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot\..

Write-Host ""
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "  Nexus Commerce — Rebuild & Reseed" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan

# ── Step 1: Rebuild ──────────────────────────────────────────

if ($Full) {
    Write-Host ""
    Write-Host "[1/4] Rebuilding ALL services..." -ForegroundColor Yellow
    docker compose -f docker-compose.dev.yml build backend frontend
} else {
    Write-Host ""
    Write-Host "[1/4] Rebuilding backend..." -ForegroundColor Yellow
    docker compose -f docker-compose.dev.yml build backend
}

if ($LASTEXITCODE -ne 0) {
    Write-Host "  BUILD FAILED" -ForegroundColor Red
    exit 1
}
Write-Host "  Build complete" -ForegroundColor Green

# ── Step 2: Restart services ────────────────────────────────

Write-Host ""
Write-Host "[2/4] Restarting services..." -ForegroundColor Yellow

if ($Full) {
    docker compose -f docker-compose.dev.yml up -d postgres redis backend frontend
} else {
    docker compose -f docker-compose.dev.yml up -d postgres redis backend
}

# Wait for postgres to be healthy
Write-Host "  Waiting for postgres..." -ForegroundColor DarkGray
$retries = 0
do {
    Start-Sleep -Seconds 2
    $health = docker compose -f docker-compose.dev.yml ps postgres --format json | ConvertFrom-Json
    $retries++
} while ($health.Health -ne "healthy" -and $retries -lt 15)

if ($health.Health -ne "healthy") {
    Write-Host "  Postgres not healthy after 30s" -ForegroundColor Red
    exit 1
}
Write-Host "  Services running" -ForegroundColor Green

# ── Step 3: Run Migrations ──────────────────────────────────

Write-Host ""
Write-Host "[3/4] Running migrations..." -ForegroundColor Yellow
docker compose -f docker-compose.dev.yml run --rm migrate

if ($LASTEXITCODE -ne 0) {
    Write-Host "  MIGRATION FAILED" -ForegroundColor Red
    exit 1
}
Write-Host "  Migrations applied" -ForegroundColor Green

# ── Step 4: Seed Database ───────────────────────────────────

if (-not $SkipSeed) {
    Write-Host ""
    Write-Host "[4/4] Seeding database..." -ForegroundColor Yellow
    Get-Content "app\backend-api\seed.sql" | docker exec -i (
        docker compose -f docker-compose.dev.yml ps -q postgres
    ) psql -U user -d saas_db

    if ($LASTEXITCODE -ne 0) {
        Write-Host "  SEED FAILED" -ForegroundColor Red
        exit 1
    }
    Write-Host "  Database seeded" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "[4/4] Skipping seed (--SkipSeed)" -ForegroundColor DarkGray
}

# ── Done ────────────────────────────────────────────────────

Write-Host ""
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "  Ready!  " -ForegroundColor Green -NoNewline
Write-Host "           Backend: http://localhost:8080" -ForegroundColor White
Write-Host "           Frontend: http://localhost:3000" -ForegroundColor White
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "  Test users (password: password123):" -ForegroundColor DarkGray
Write-Host "    owner@coolshoes.com  → coolshoes.localhost:3000" -ForegroundColor DarkGray
Write-Host "    owner@techstore.com  → techstore.localhost:3000" -ForegroundColor DarkGray
Write-Host ""
