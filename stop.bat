@echo off
REM AMBA - stop and remove containers. Mirrors "make down".
setlocal
cd /d "%~dp0"

docker info >nul 2>&1
if errorlevel 1 (
  echo Docker is not running. Start Docker Desktop, then run this again.
  exit /b 1
)

docker compose down
