#!/usr/bin/env bash
# dev.sh — build backend + frontend dev server with a layout that mirrors install.sh
#
# Layout (under hub/.dev-run/, gitignored):
#   .dev-run/bin/couchdev        built binary
#   .dev-run/etc/config.json     config + token (persisted between runs)
#   .dev-run/data/couchdev.db    database
#   .dev-run/data/projects/      projects dir (registered in Claude trust)
#
# Vite proxies /api → localhost:8443, so the frontend dev server and the
# backend can run side-by-side without CORS issues.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEV_DIR="$SCRIPT_DIR/.dev-run"
BIN="$DEV_DIR/bin/couchdev"
CONFIG="$DEV_DIR/etc/config.json"
PROJECTS_DIR="$DEV_DIR/data/projects"

# ── find go ───────────────────────────────────────────────────────────────────

find_go() {
  if command -v go &>/dev/null; then echo go; return; fi
  local found
  found=$(find "${GOPATH:-$HOME/go}/pkg/mod" -name "go" -path "*/toolchain*/bin/go" \
          2>/dev/null | sort -V | tail -1)
  if [[ -n "$found" ]]; then echo "$found"; return; fi
  echo "error: go not found in PATH or GOPATH — install from https://go.dev/dl/" >&2
  exit 1
}

GO=$(find_go)

# ── build backend ─────────────────────────────────────────────────────────────

echo "→ building backend..."
mkdir -p "$DEV_DIR/bin"
GOPATH="${GOPATH:-$HOME/go}" "$GO" build -o "$BIN" "$SCRIPT_DIR/cmd/couchdev"
echo "  built → $BIN"

# ── dev layout (mirrors install.sh) ──────────────────────────────────────────

mkdir -p "$DEV_DIR/etc" "$PROJECTS_DIR"

CLAUDE_SRC="$SCRIPT_DIR/claude/hub-claude.md"
if [[ -f "$CLAUDE_SRC" ]] && [[ ! -f "$PROJECTS_DIR/CLAUDE.md" ]]; then
  cp "$CLAUDE_SRC" "$PROJECTS_DIR/CLAUDE.md"
fi

if [[ ! -f "$CONFIG" ]]; then
  echo "→ generating dev config and token..."
  TOKEN_OUTPUT=$("$BIN" token generate)
  TOKEN=$(echo "$TOKEN_OUTPUT" | awk 'NR==2{print $1}')
  TOKEN_HASH=$(echo "$TOKEN_OUTPUT" | awk 'NR==5{print $1}')
  cat > "$CONFIG" <<EOF
{
  "listen_addr": ":8443",
  "db_path": "${DEV_DIR}/data/couchdev.db",
  "projects_dir": "${PROJECTS_DIR}",
  "token_hash": "${TOKEN_HASH}"
}
EOF
  echo ""
  echo "  ┌─────────────────────────────────────────────────────┐"
  echo "  │ Bearer token (save this — it won't be shown again): │"
  echo "  │                                                     │"
  echo "  │   $TOKEN  │"
  echo "  └─────────────────────────────────────────────────────┘"
  echo ""
fi

# ── register projects dir as trusted in Claude settings ───────────────────────

if command -v python3 &>/dev/null; then
  CLAUDE_SETTINGS="${HOME}/.claude/settings.json"
  python3 - "$PROJECTS_DIR" "$CLAUDE_SETTINGS" <<'PYEOF'
import sys, json, os
projects_dir, settings_path = sys.argv[1], sys.argv[2]
settings = {}
if os.path.exists(settings_path):
    try:
        with open(settings_path) as f:
            settings = json.load(f)
    except (json.JSONDecodeError, OSError):
        pass
dirs = settings.setdefault("permissions", {}).setdefault("additionalDirectories", [])
if projects_dir not in dirs:
    dirs.append(projects_dir)
    with open(settings_path, "w") as f:
        json.dump(settings, f, indent=4)
    print(f"  trusted → {projects_dir} in ~/.claude/settings.json")
PYEOF
fi

# ── frontend deps ─────────────────────────────────────────────────────────────

if [[ ! -d "$SCRIPT_DIR/web/node_modules" ]]; then
  echo "→ installing frontend dependencies..."
  npm --prefix "$SCRIPT_DIR/web" install
fi

# ── start servers ─────────────────────────────────────────────────────────────

cleanup() {
  echo ""
  echo "→ shutting down..."
  [[ -n "${BACKEND_PID:-}" ]]  && kill "$BACKEND_PID"  2>/dev/null || true
  [[ -n "${FRONTEND_PID:-}" ]] && kill "$FRONTEND_PID" 2>/dev/null || true
}
trap cleanup INT TERM EXIT

echo "→ starting backend on :8443..."
"$BIN" serve --config "$CONFIG" &
BACKEND_PID=$!

echo "→ starting frontend dev server on :5174..."
npm --prefix "$SCRIPT_DIR/web" run dev -- --host --port 5174 &
FRONTEND_PID=$!

echo ""
echo "  Hub UI → http://localhost:5174"
echo "  API    → http://localhost:8443"
echo ""
echo "  Press Ctrl+C to stop."
echo ""

wait "$BACKEND_PID" "$FRONTEND_PID"
