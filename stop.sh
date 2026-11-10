#!/usr/bin/env bash
# AMBA - stop and remove the containers. Safe to run any time.
set -u
cd "$(dirname "$0")"
docker compose down && printf '\nAMBA sudah dimatikan.\n'
