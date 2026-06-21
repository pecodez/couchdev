#!/bin/sh
set -e

# Usage: install.sh <hub-ip> <nfs-export-path> [data-root]
# Example: install.sh 192.168.1.10 /srv/couchdev /mnt/couchdev

HUB_IP="${1:?Usage: install.sh <hub-ip> <nfs-export-path> [data-root]}"
NFS_EXPORT="${2:?NFS export path required}"
DATA_ROOT="${3:-/mnt/couchdev}"

SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "==> Installing Node.js LTS"
if ! command -v node > /dev/null 2>&1; then
  curl -fsSL https://deb.nodesource.com/setup_lts.x | sh -
  apt-get install -y nodejs
fi

echo "==> Installing npm dependencies"
cd "$SCRIPT_DIR"
npm install

echo "==> Installing Playwright Chromium"
npx playwright install chromium --with-deps

echo "==> Building TypeScript"
npm run build

echo "==> Mounting NFS share"
apt-get install -y nfs-common
mkdir -p "$DATA_ROOT"
FSTAB_ENTRY="${HUB_IP}:${NFS_EXPORT} ${DATA_ROOT} nfs defaults,_netdev 0 0"
if ! grep -qF "$FSTAB_ENTRY" /etc/fstab; then
  echo "$FSTAB_ENTRY" >> /etc/fstab
fi
mount "$DATA_ROOT" 2>/dev/null || true

echo "==> Writing DATA_ROOT to environment"
if ! grep -q "^DATA_ROOT=" /etc/environment 2>/dev/null; then
  echo "DATA_ROOT=${DATA_ROOT}" >> /etc/environment
fi

echo "==> Registering MCP server with Claude Code"
SETTINGS="$HOME/.claude/settings.json"
if [ ! -f "$SETTINGS" ]; then
  mkdir -p "$(dirname "$SETTINGS")"
  echo '{}' > "$SETTINGS"
fi
node -e "
const fs = require('fs');
const s = JSON.parse(fs.readFileSync('$SETTINGS', 'utf8'));
s.mcpServers = s.mcpServers || {};
s.mcpServers['webdev'] = {
  command: 'node',
  args: ['${SCRIPT_DIR}/dist/index.js'],
  env: { DATA_ROOT: '${DATA_ROOT}' }
};
fs.writeFileSync('$SETTINGS', JSON.stringify(s, null, 2));
"

echo ""
echo "==> Done. WebDev MCP node is ready."
echo "    DATA_ROOT:  ${DATA_ROOT}"
echo "    MCP server: node ${SCRIPT_DIR}/dist/index.js"
