@echo off
REM AMBA - quick start (Windows, Docker Desktop). Mirrors "make up".
setlocal
cd /d "%~dp0"

docker info >nul 2>&1
if errorlevel 1 (
  echo Docker is not running. Start Docker Desktop, then run this again.
  exit /b 1
)

echo Building and starting AMBA...
docker compose up --build
