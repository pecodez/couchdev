#!/usr/bin/env bash
set -euo pipefail

# couchdev must run as the same user as claude and tmux — system-wide installs
# under a separate service account cannot access the user's Claude credentials
# or tmux sessions.  User-mode only.
if [[ "${EUID:-$(id -u)}" -eq 0 ]]; then
  echo "error: do not run install.sh as root — couchdev must run as your own user" >&2
  echo "       Re-run without sudo." >&2
  exit 1
fi

REPO="pecodez/couchdev"

VERSION="latest"
LOCAL_BIN=""
FORCE=false
FORCE_CLAUDE_MD=false
EXPLICIT_PREFIX=""

usage() {
  cat <<EOF
Usage: install.sh [OPTIONS]

Options:
  --prefix=PATH         Override install prefix (default: \$HOME/.local/couchdev)
  --version=TAG         GitHub release tag to download (default: latest)
  --local-bin=PATH      Skip download; install this pre-built binary instead
  --force               Overwrite existing config
  --force-claude-md     Overwrite existing projects CLAUDE.md (e.g. on upgrade)
  -h, --help            Show this help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --prefix=*)      EXPLICIT_PREFIX="${1#*=}" ;;
    --prefix)        EXPLICIT_PREFIX="$2"; shift ;;
    --version=*)     VERSION="${1#*=}" ;;
    --version)       VERSION="$2"; shift ;;
    --local-bin=*)   LOCAL_BIN="${1#*=}" ;;
    --local-bin)     LOCAL_BIN="$2"; shift ;;
    --force)         FORCE=true ;;
    --force-claude-md) FORCE_CLAUDE_MD=true ;;
    -h|--help)       usage; exit 0 ;;
    *) echo "Unknown flag: $1" >&2; usage >&2; exit 1 ;;
  esac
  shift
done

PREFIX="${EXPLICIT_PREFIX:-${HOME}/.local/couchdev}"
BIN_DIR="$PREFIX/bin"
ETC_DIR="$PREFIX/etc"
DATA_DIR="$PREFIX/data"
PROJECTS_DIR="$DATA_DIR/projects"

case "$(uname -m)" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *) echo "error: unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

# ── dependency checks ─────────────────────────────────────────────────────────

DEPS_OK=true

if ! command -v tmux &>/dev/null; then
  echo "error: tmux is not installed — sessions cannot be created without it" >&2
  echo "       Install with: apt install tmux  (or brew install tmux)" >&2
  DEPS_OK=false
fi

if ! command -v claude &>/dev/null; then
  echo "warning: claude CLI not found in PATH — sessions will fail until it is installed" >&2
  echo "         See https://claude.ai/code to install Claude Code" >&2
fi

# On Debian/Ubuntu, libsecret-1-0 is required for Claude Code to persist OAuth credentials
# in the keyring. Without it the remote-control bridge token is not stored between restarts
# and remote control will silently fail.
if command -v dpkg &>/dev/null; then
  if ! dpkg -l libsecret-1-0 2>/dev/null | grep -q "^ii"; then
    echo "error: libsecret-1-0 is not installed — Claude Code cannot persist OAuth credentials" >&2
    echo "" >&2
    echo "       Install it first, then re-run this installer:" >&2
    echo "" >&2
    echo "         sudo apt install libsecret-1-0 gnome-keyring" >&2
    echo "" >&2
    exit 1
  fi
fi

# Check whether Claude Code has been authenticated. Remote control (the feature that lets
# you connect to sessions from Claude mobile) requires a valid login — it will silently
# fail if this step is skipped.
_EMAIL=""
if command -v python3 &>/dev/null && command -v claude &>/dev/null; then
  _CLAUDE_JSON="${HOME}/.claude.json"
  if [[ -f "$_CLAUDE_JSON" ]]; then
    _EMAIL=$(python3 -c "
import json, sys
try:
    d = json.load(open(sys.argv[1]))
    acct = d.get('oauthAccount') or {}
    print(acct.get('emailAddress', ''))
except Exception:
    pass
" "$_CLAUDE_JSON" 2>/dev/null)
  fi
  if [[ -n "$_EMAIL" ]]; then
    echo "Claude auth   → authenticated as ${_EMAIL}"
  else
    echo "warning: Claude Code is not authenticated — remote control will not work" >&2
    echo "         Run:  claude login" >&2
  fi
fi

if ! command -v python3 &>/dev/null; then
  echo "warning: python3 not found — Claude workspace trust registration will be skipped" >&2
fi

if [[ "$DEPS_OK" == false ]]; then
  exit 1
fi

# ── acquire binary ────────────────────────────────────────────────────────────

TMP_BIN="$(mktemp)"
trap 'rm -f "$TMP_BIN"' EXIT

if [[ -n "$LOCAL_BIN" ]]; then
  if [[ ! -f "$LOCAL_BIN" ]]; then
    echo "error: --local-bin path not found: $LOCAL_BIN" >&2
    exit 1
  fi
  cp "$LOCAL_BIN" "$TMP_BIN"
  echo "Using local binary: $LOCAL_BIN"
else
  if command -v curl &>/dev/null; then
    FETCH() { curl -fsSL "$1"; }
  elif command -v wget &>/dev/null; then
    FETCH() { wget -qO- "$1"; }
  else
    echo "error: neither curl nor wget is available — install one or use --local-bin" >&2
    exit 1
  fi

  if [[ "$VERSION" == "latest" ]]; then
    URL="https://github.com/${REPO}/releases/latest/download/couchdev-linux-${ARCH}"
  else
    URL="https://github.com/${REPO}/releases/download/${VERSION}/couchdev-linux-${ARCH}"
  fi

  echo "Downloading couchdev ${VERSION} (linux/${ARCH}) ..."
  FETCH "$URL" > "$TMP_BIN"
fi

chmod +x "$TMP_BIN"

if ! "$TMP_BIN" --help >/dev/null 2>&1; then
  echo "error: binary sanity check failed — may be wrong architecture or corrupt" >&2
  exit 1
fi

# ── directory layout ──────────────────────────────────────────────────────────

mkdir -p "$BIN_DIR" "$ETC_DIR" "$PROJECTS_DIR"

# ── hub CLAUDE.md ─────────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HUB_CLAUDE_SRC="$SCRIPT_DIR/claude/hub-claude.md"
PROJECTS_CLAUDE="$PROJECTS_DIR/CLAUDE.md"

if [[ -f "$HUB_CLAUDE_SRC" ]]; then
  if [[ ! -f "$PROJECTS_CLAUDE" ]] || [[ "$FORCE_CLAUDE_MD" == true ]]; then
    cp "$HUB_CLAUDE_SRC" "$PROJECTS_CLAUDE"
    echo "Installed CLAUDE.md → $PROJECTS_CLAUDE"
  else
    echo "Skipping CLAUDE.md  (already exists; use --force-claude-md to overwrite)"
  fi
fi

# ── install binary ────────────────────────────────────────────────────────────

cp "$TMP_BIN" "$BIN_DIR/couchdev"
chmod +x "$BIN_DIR/couchdev"
echo "Installed binary → $BIN_DIR/couchdev"

LINK_DIR="${HOME}/.local/bin"
mkdir -p "$LINK_DIR"
ln -sf "$BIN_DIR/couchdev" "$LINK_DIR/couchdev"
echo "Symlinked        → $LINK_DIR/couchdev"

# ── starter config ────────────────────────────────────────────────────────────

CONFIG_FILE="$ETC_DIR/config.json"
GENERATED_TOKEN=""
if [[ -f "$CONFIG_FILE" ]] && [[ "$FORCE" == false ]]; then
  echo "Config exists at $CONFIG_FILE (use --force to overwrite)"
else
  TOKEN_OUTPUT=$("$BIN_DIR/couchdev" token generate)
  GENERATED_TOKEN=$(echo "$TOKEN_OUTPUT" | awk 'NR==2{print $1}')
  TOKEN_HASH=$(echo "$TOKEN_OUTPUT" | awk 'NR==5{print $1}')
  cat > "$CONFIG_FILE" <<EOF
{
  "listen_addr": ":8080",
  "db_path": "${DATA_DIR}/couchdev.db",
  "projects_dir": "${PROJECTS_DIR}",
  "require_auth": true,
  "token_hash": "${TOKEN_HASH}"
}
EOF
  echo "Wrote config     → $CONFIG_FILE"
fi

# ── register projects dir as trusted in Claude settings ───────────────────────
# Claude shows a workspace trust dialog for directories it hasn't seen before.
# Adding the projects dir to additionalDirectories means all project worktrees
# (children of this path) are trusted from the first session — no dialog needed.

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
    print(f"Trusted projects → {projects_dir}")
else:
    print(f"Already trusted  → {projects_dir}")
PYEOF
fi

# ── post-install instructions ─────────────────────────────────────────────────

cat <<EOF

══════════════════════════════════════════
 Couchdev installed → ${PREFIX}
══════════════════════════════════════════
EOF

if [[ -n "$GENERATED_TOKEN" ]]; then
  cat <<EOF

Bearer token (add this in the Claude android app):

   ${GENERATED_TOKEN}

EOF
fi

if [[ -z "$_EMAIL" ]]; then
  cat <<EOF
Next steps:

1. Authenticate Claude Code (required for remote control):

   claude login

2. Start couchdev:

   couchdev serve --config ${CONFIG_FILE}

EOF
else
  cat <<EOF
Next steps:

Start couchdev:

   couchdev serve --config ${CONFIG_FILE}

EOF
fi
