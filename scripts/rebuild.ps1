param(
    [switch]$Full,
    [switch]$SkipSeed
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot\..

if ($Full) {
    Write-Host "[1/4] Building backend and frontend..." -ForegroundColor Yellow
    docker compose -f docker-compose.dev.yml build backend frontend
} else {
    Write-Host "[1/4] Building backend..." -ForegroundColor Yellow
    docker compose -f docker-compose.dev.yml build backend
}

if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed." -ForegroundColor Red
    exit 1
}

Write-Host "[2/4] Starting containers..." -ForegroundColor Yellow
if ($Full) {
    docker compose -f docker-compose.dev.yml up -d postgres redis backend frontend
} else {
    docker compose -f docker-compose.dev.yml up -d postgres redis backend
}

$retryCount = 0
do {
    Start-Sleep -Seconds 2
    $postgresStatus = docker compose -f docker-compose.dev.yml ps postgres --format json | ConvertFrom-Json
    $retryCount++
} while ($postgresStatus.Health -ne "healthy" -and $retryCount -lt 15)

if ($postgresStatus.Health -ne "healthy") {
    Write-Host "PostgreSQL failed health check after 30s." -ForegroundColor Red
    exit 1
}

Write-Host "[3/4] Applying migrations..." -ForegroundColor Yellow
docker compose -f docker-compose.dev.yml run --rm migrate
if ($LASTEXITCODE -ne 0) {
    Write-Host "Migration failed." -ForegroundColor Red
    exit 1
}

if (-not $SkipSeed) {
    Write-Host "[4/4] Seeding database..." -ForegroundColor Yellow
    Get-Content "app\backend-api\seed.sql" | docker exec -i (
        docker compose -f docker-compose.dev.yml ps -q postgres
    ) psql -U user -d saas_db

    if ($LASTEXITCODE -ne 0) {
        Write-Host "Database seed failed." -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "[4/4] Skipping seed." -ForegroundColor DarkGray
}

Write-Host "Stack ready: Backend http://localhost:8080 | Frontend http://localhost:3000" -ForegroundColor Green
