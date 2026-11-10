@echo off
REM AMBA - one-click start (Windows, Docker Desktop). Just double-click this file.
setlocal
cd /d "%~dp0"

where docker >nul 2>&1
if errorlevel 1 (
  echo Docker belum terpasang.
  echo Pasang Docker Desktop dari https://www.docker.com/products/docker-desktop/
  echo lalu jalankan file ini lagi.
  pause
  exit /b 1
)

docker info >nul 2>&1
if errorlevel 1 (
  echo Docker Desktop belum berjalan.
  echo Buka Docker Desktop, tunggu sampai muncul "Engine running",
  echo lalu jalankan file ini lagi.
  pause
  exit /b 1
)

REM Already running? Report where and change nothing. Without this a second
REM double-click would look like "port 80 is busy" and silently move the app.
docker compose ps --services --filter status=running 2>nul | findstr /x "frontend" >nul
if not errorlevel 1 (
  set APP_PORT=80
  for /f "tokens=2 delims=:" %%p in ('docker compose port frontend 80 2^>nul') do set APP_PORT=%%p
  echo AMBA sudah jalan: http://localhost:%APP_PORT%
  echo.
  echo Mau mematikan? Jalankan stop.bat
  pause
  exit /b 0
)

REM A busy port 80 must not turn into a cryptic error.
netstat -ano | findstr /r /c:":80 .*LISTENING" >nul 2>&1
if not errorlevel 1 (
  set FRONTEND_PORT=3000
  echo Port 80 sedang dipakai, AMBA memakai port 3000.
)

echo Menyalakan AMBA. Saat pertama kali, proses build bisa memakan beberapa menit...
echo.
docker compose up -d --build
if errorlevel 1 (
  echo.
  echo Gagal menyalakan AMBA. Lihat pesan di atas.
  pause
  exit /b 1
)

echo.
echo Menunggu AMBA siap...
timeout /t 8 /nobreak >nul

if defined FRONTEND_PORT (
  echo.
  echo   AMBA sudah jalan: http://localhost:%FRONTEND_PORT%
) else (
  echo.
  echo   AMBA sudah jalan: http://localhost
)
echo.
echo   Untuk mematikan nanti, jalankan stop.bat
echo   Jendela ini boleh ditutup.
echo.
pause
