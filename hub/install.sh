#!/usr/bin/env bash
set -euo pipefail

REPO="pecodez/couchdev"

VERSION="latest"
LOCAL_BIN=""
FORCE=false
FORCE_CLAUDE_MD=false
EXPLICIT_PREFIX=""

if [[ "${EUID:-$(id -u)}" -eq 0 ]]; then
  MODE="system"
else
  MODE="user"
fi

usage() {
  cat <<EOF
Usage: install.sh [OPTIONS]

Options:
  --system              Install system-wide under /opt/couchdev (default when run as root)
  --user                Install for current user under \$HOME/.local/couchdev
  --prefix=PATH         Override install prefix
  --version=TAG         GitHub release tag to download (default: latest)
  --local-bin=PATH      Skip download; install this pre-built binary instead
  --force               Overwrite existing config
  --force-claude-md     Overwrite existing projects CLAUDE.md (e.g. on upgrade)
  -h, --help            Show this help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --system)        MODE="system" ;;
    --user)          MODE="user" ;;
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

if [[ -n "$EXPLICIT_PREFIX" ]]; then
  PREFIX="$EXPLICIT_PREFIX"
elif [[ "$MODE" == "system" ]]; then
  PREFIX="/opt/couchdev"
else
  PREFIX="${HOME}/.local/couchdev"
fi

BIN_DIR="$PREFIX/bin"
ETC_DIR="$PREFIX/etc"
DATA_DIR="$PREFIX/data"

case "$(uname -m)" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *) echo "error: unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

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

mkdir -p "$BIN_DIR" "$ETC_DIR" "$DATA_DIR/projects"

# ── hub CLAUDE.md ─────────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HUB_CLAUDE_SRC="$SCRIPT_DIR/claude/hub-claude.md"
PROJECTS_DIR="${DATA_DIR}/projects"
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

if [[ "$MODE" == "system" ]]; then
  LINK_DIR="/usr/local/bin"
else
  LINK_DIR="${HOME}/.local/bin"
  mkdir -p "$LINK_DIR"
fi
ln -sf "$BIN_DIR/couchdev" "$LINK_DIR/couchdev"
echo "Symlinked        → $LINK_DIR/couchdev"

# ── starter config ────────────────────────────────────────────────────────────

CONFIG_FILE="$ETC_DIR/config.json"
GENERATED_TOKEN=""
if [[ -f "$CONFIG_FILE" ]] && [[ "$FORCE" == false ]]; then
  echo "Config exists at $CONFIG_FILE (use --force to overwrite)"
else
  GENERATED_TOKEN=$(openssl rand -hex 32)
  TOKEN_HASH=$(printf '%s' "$GENERATED_TOKEN" | sha256sum | cut -d' ' -f1)
  cat > "$CONFIG_FILE" <<EOF
{
  "listen_addr": ":8443",
  "db_path": "${DATA_DIR}/couchdev.db",
  "projects_dir": "${DATA_DIR}/projects",
  "token_hash": "${TOKEN_HASH}"
}
EOF
  echo "Wrote config     → $CONFIG_FILE"
fi

# ── systemd unit ──────────────────────────────────────────────────────────────

install_system_unit() {
  local dest="/etc/systemd/system/couchdev.service"
  sed "s|@PREFIX@|${PREFIX}|g" > "$dest" <<'UNIT'
[Unit]
Description=Couchdev hub
After=network.target

[Service]
Type=simple
User=couchdev
Group=couchdev
ExecStart=@PREFIX@/bin/couchdev serve --config @PREFIX@/etc/config.json
Restart=on-failure
RestartSec=5
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
ReadWritePaths=@PREFIX@/data

[Install]
WantedBy=multi-user.target
UNIT
  echo "Installed unit   → $dest"
  systemctl daemon-reload
}

install_user_unit() {
  local dir="${HOME}/.config/systemd/user"
  local dest="$dir/couchdev.service"
  mkdir -p "$dir"
  sed "s|@PREFIX@|${PREFIX}|g" > "$dest" <<'UNIT'
[Unit]
Description=Couchdev hub
After=network.target

[Service]
Type=simple
ExecStart=@PREFIX@/bin/couchdev serve --config @PREFIX@/etc/config.json
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
UNIT
  echo "Installed unit   → $dest"
  systemctl --user daemon-reload 2>/dev/null || true
}

if command -v systemctl &>/dev/null; then
  if [[ "$MODE" == "system" ]]; then
    install_system_unit
  else
    install_user_unit
  fi
else
  echo "systemctl not found — skipping service install"
fi

# ── post-install instructions ─────────────────────────────────────────────────

cat <<EOF

══════════════════════════════════════════
 Couchdev installed → ${PREFIX}
══════════════════════════════════════════
EOF

if [[ -n "$GENERATED_TOKEN" ]]; then
  cat <<EOF

Bearer token (configure this in your client):

   ${GENERATED_TOKEN}

EOF
fi

if [[ "$MODE" == "system" ]]; then
  cat <<EOF
Next steps:

1. Create the service user and set ownership:

   useradd -r -s /usr/sbin/nologin -d ${DATA_DIR} couchdev
   chown -R couchdev:couchdev ${PREFIX}

2. Enable and start:

   systemctl enable --now couchdev

EOF
else
  cat <<EOF
Next steps:

1. Enable and start:

   systemctl --user enable --now couchdev

EOF
fi

echo "   Or run directly:  couchdev serve --config ${CONFIG_FILE}"
echo ""
