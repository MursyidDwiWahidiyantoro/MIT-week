@echo off
REM AMBA - dev mode with hot reload (Windows, Docker Desktop). Mirrors "make dev".
setlocal
cd /d "%~dp0"

docker info >nul 2>&1
if errorlevel 1 (
  echo Docker is not running. Start Docker Desktop, then run this again.
  exit /b 1
)

echo Starting AMBA in dev mode (Go Air + Vite HMR)...
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
