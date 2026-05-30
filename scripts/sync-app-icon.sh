#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC="$ROOT/assets/icon.png"
DEST="$ROOT/build/appicon.png"

if [ ! -f "$SRC" ]; then
  echo "sync-app-icon: missing source icon at $SRC" >&2
  exit 1
fi

mkdir -p "$(dirname "$DEST")"
cp "$SRC" "$DEST"
rm -f "$ROOT/build/windows/icon.ico"
