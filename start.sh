#!/usr/bin/env bash
# AMBA - one-click start for Linux and macOS. Needs Docker only.
set -u
cd "$(dirname "$0")"

fail() {
  printf '\n%s\n\nTekan Enter untuk menutup...' "$1"
  read -r _
  exit 1
}

command -v docker >/dev/null 2>&1 || fail "Docker belum terpasang.
Pasang Docker Desktop dari https://www.docker.com/products/docker-desktop/
lalu jalankan file ini lagi."

docker info >/dev/null 2>&1 || fail "Docker terpasang tapi belum berjalan.
Buka Docker Desktop (Linux: sudo systemctl start docker), tunggu sampai
statusnya running, lalu jalankan file ini lagi."

# Already running? Report where, and change nothing. Otherwise a second
# double-click would look like "port 80 is busy" and silently move the app.
if docker compose ps --status running --services 2>/dev/null | grep -qx frontend; then
  PORT=$(docker compose port frontend 80 2>/dev/null | sed 's/.*://')
  [ -n "$PORT" ] || PORT=80
  URL="http://localhost"
  [ "$PORT" != "80" ] && URL="http://localhost:$PORT"
  printf 'AMBA sudah jalan: %s\n\nMau mematikan? Jalankan ./stop.sh\n' "$URL"
  exit 0
fi

# A busy port 80 must not turn into a cryptic error for a non-technical user.
PORT=80
if (ss -ltn 2>/dev/null || netstat -ltn 2>/dev/null) | grep -qE '[:.]80[[:space:]]'; then
  PORT=3000
  printf 'Port 80 sedang dipakai, AMBA memakai port %s.\n' "$PORT"
fi

printf 'Menyalakan AMBA. Saat pertama kali, proses build bisa memakan beberapa menit...\n\n'
FRONTEND_PORT=$PORT docker compose up -d --build || fail "Gagal menyalakan AMBA. Lihat pesan di atas."

URL="http://localhost"
[ "$PORT" != "80" ] && URL="http://localhost:$PORT"
printf 'Menunggu AMBA siap'
for _ in $(seq 1 60); do
  curl -fsS -o /dev/null "$URL" 2>/dev/null && break
  printf '.'
  sleep 2
done
printf '\n\n  AMBA sudah jalan: %s\n\n  Untuk mematikan nanti, jalankan ./stop.sh\n  Jendela ini boleh ditutup.\n\n' "$URL"

