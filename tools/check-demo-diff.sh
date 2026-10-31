#!/usr/bin/env bash
# Fails when the backend demo data and the frontend copy drift apart.
# Both files must be byte-identical (TECH_SPEC §9).
set -euo pipefail

BACKEND="backend/internal/demo/demo-cases.json"
FRONTEND="frontend/src/lib/data/demo-cases.json"

if [[ ! -f "$BACKEND" ]]; then
  echo "skip: $BACKEND not created yet (Phase 1)"
  exit 0
fi

if ! diff -q "$BACKEND" "$FRONTEND" >/dev/null; then
  echo "FAIL: demo-cases.json differs between backend and frontend"
  diff "$BACKEND" "$FRONTEND" | head -20
  exit 1
fi

echo "demo-cases.json in sync"
