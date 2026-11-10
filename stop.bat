@echo off
REM AMBA - stop and remove the containers. Safe to run any time.
setlocal
cd /d "%~dp0"

where docker >nul 2>&1
if errorlevel 1 (
  echo Docker belum terpasang.
  pause
  exit /b 1
)

docker compose down
echo.
echo AMBA sudah dimatikan.
pause
