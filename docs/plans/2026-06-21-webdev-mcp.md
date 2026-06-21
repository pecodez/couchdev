# WebDev MCP Node — Implementation Plan

**Goal:** Build a stateless stdio MCP server in TypeScript that exposes four macro-tools for running Storybook, Playwright, Vite build, and Vite preview on a web project whose source lives on an NFS mount.

**Architecture:** stdio MCP process spawned by Claude over SSH. Each tool call detaches a background job, writes output to `DATA_ROOT/logs/<project>/<pid>/<tool>.log` and screenshots to `DATA_ROOT/screenshots/<project>/<pid>/<name>.png`, and returns `{ "pid": n }` immediately. Completion is signalled by a terminal marker line in the log.

**Tech Stack:** Node.js LTS, TypeScript, `@modelcontextprotocol/sdk`, Playwright (Chromium headless)

## Global Constraints

- All tool names prefixed `webdev_`
- `DATA_ROOT` env var, default `/mnt/couchdev`
- Log path: `DATA_ROOT/logs/basename(project_path)/<pid>/<tool>.log`
- Screenshot path: `DATA_ROOT/screenshots/basename(project_path)/<pid>/<name>.png`
- Every tool returns `{ "pid": number }` immediately after spawning
- Terminal log markers: `COUCHDEV:EXIT:0` (success) / `COUCHDEV:EXIT:1` (failure)
- `tag` and `test_name` are mutually exclusive in `webdev_run_playwright`
- Module lives at `mcp/webdev/` in the couchdev repo

---

## File Map

```
mcp/webdev/
├── src/
│   ├── index.ts           # MCP server, tool registration
│   ├── lib/
│   │   ├── paths.ts       # DATA_ROOT, path builders
│   │   └── spawn.ts       # detached process helper
│   └── tools/
│       ├── storybook.ts   # webdev_run_storybook
│       ├── playwright.ts  # webdev_run_playwright
│       ├── build.ts       # webdev_build
│       └── preview.ts     # webdev_preview
├── package.json
├── tsconfig.json
└── setup/
    └── install.sh
```

---

### Task 1: npm scaffold + path helpers

**Files:**
- Create: `mcp/webdev/package.json`
- Create: `mcp/webdev/tsconfig.json`
- Create: `mcp/webdev/src/lib/paths.ts`

**Produces:** `buildLogPath(projectPath, pid, tool)`, `buildScreenshotPath(projectPath, pid, name)`, `DATA_ROOT`

- [ ] **Write `mcp/webdev/package.json`**

```json
{
  "name": "@couchdev/webdev-mcp",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "build": "tsc",
    "start": "node dist/index.js",
    "dev": "tsx src/index.ts"
  },
  "dependencies": {
    "@modelcontextprotocol/sdk": "^1.12.1",
    "playwright": "^1.52.0",
    "zod": "^3.25.67"
  },
  "devDependencies": {
    "@types/node": "^22.15.29",
    "tsx": "^4.19.4",
    "typescript": "^5.8.3"
  }
}
```

- [ ] **Write `mcp/webdev/tsconfig.json`**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "outDir": "dist",
    "rootDir": "src",
    "strict": true,
    "esModuleInterop": true
  },
  "include": ["src"]
}
```

- [ ] **Install dependencies**

```bash
cd mcp/webdev && npm install
```

- [ ] **Write `mcp/webdev/src/lib/paths.ts`**

```typescript
import path from 'node:path'
import fs from 'node:fs'

export const DATA_ROOT = process.env.DATA_ROOT ?? '/mnt/couchdev'

export function projectName(projectPath: string): string {
  return path.basename(projectPath)
}

export function buildLogPath(projectPath: string, pid: number, tool: string): string {
  const dir = path.join(DATA_ROOT, 'logs', projectName(projectPath), String(pid))
  fs.mkdirSync(dir, { recursive: true })
  return path.join(dir, `${tool}.log`)
}

export function buildScreenshotPath(projectPath: string, pid: number, name: string): string {
  const dir = path.join(DATA_ROOT, 'screenshots', projectName(projectPath), String(pid))
  fs.mkdirSync(dir, { recursive: true })
  return path.join(dir, `${name}.png`)
}
```

- [ ] **Verify TypeScript compiles**

```bash
cd mcp/webdev && npm run build
```

Expected: `dist/` created, no errors.

- [ ] **Commit**

```bash
git add mcp/webdev/package.json mcp/webdev/package-lock.json mcp/webdev/tsconfig.json mcp/webdev/src/lib/paths.ts
git commit -m "feat(webdev-mcp): scaffold and path helpers"
```

---

### Task 2: Detached spawn helper

**Files:**
- Create: `mcp/webdev/src/lib/spawn.ts`

**Consumes:** `buildLogPath(projectPath, pid, tool)` from `paths.ts`

**Produces:** `spawnDetached(cmd, args, cwd, logPath): Promise<number>` — resolves to PID

- [ ] **Write `mcp/webdev/src/lib/spawn.ts`**

```typescript
import { spawn } from 'node:child_process'
import fs from 'node:fs'

export async function spawnDetached(
  cmd: string,
  args: string[],
  cwd: string,
  logPath: string,
): Promise<number> {
  const logFd = fs.openSync(logPath, 'w')
  const child = spawn(cmd, args, {
    cwd,
    stdio: ['ignore', logFd, logFd],
    detached: true,
  })
  child.unref()
  fs.closeSync(logFd)

  if (child.pid === undefined) {
    throw new Error(`Failed to spawn ${cmd}`)
  }
  return child.pid
}

export function exitMarker(code: 0 | 1): string {
  return `COUCHDEV:EXIT:${code}`
}
```

- [ ] **Verify TypeScript compiles**

```bash
cd mcp/webdev && npm run build
```

Expected: no errors.

- [ ] **Commit**

```bash
git add mcp/webdev/src/lib/spawn.ts
git commit -m "feat(webdev-mcp): detached spawn helper"
```

---

### Task 3: MCP server entry point + tool stubs

**Files:**
- Create: `mcp/webdev/src/index.ts`
- Create: `mcp/webdev/src/tools/storybook.ts` (stub)
- Create: `mcp/webdev/src/tools/playwright.ts` (stub)
- Create: `mcp/webdev/src/tools/build.ts` (stub)
- Create: `mcp/webdev/src/tools/preview.ts` (stub)

**Consumes:** `@modelcontextprotocol/sdk`

**Produces:** runnable MCP server that lists four tools and returns a stub response

- [ ] **Write `mcp/webdev/src/tools/storybook.ts`**

```typescript
import { z } from 'zod'

export const storybookSchema = z.object({
  project_path: z.string().min(1),
})

export async function runStorybook(
  _params: z.infer<typeof storybookSchema>,
): Promise<{ pid: number }> {
  throw new Error('not implemented')
}
```

- [ ] **Write `mcp/webdev/src/tools/playwright.ts`**

```typescript
import { z } from 'zod'

export const playwrightSchema = z.object({
  project_path: z.string().min(1),
  spec: z.string().optional(),
  test_name: z.string().optional(),
  tag: z.string().optional(),
})

export async function runPlaywright(
  _params: z.infer<typeof playwrightSchema>,
): Promise<{ pid: number }> {
  throw new Error('not implemented')
}
```

- [ ] **Write `mcp/webdev/src/tools/build.ts`**

```typescript
import { z } from 'zod'

export const buildSchema = z.object({
  project_path: z.string().min(1),
  mode: z.string().min(1),
})

export async function runBuild(
  _params: z.infer<typeof buildSchema>,
): Promise<{ pid: number }> {
  throw new Error('not implemented')
}
```

- [ ] **Write `mcp/webdev/src/tools/preview.ts`**

```typescript
import { z } from 'zod'

export const previewSchema = z.object({
  project_path: z.string().min(1),
  mode: z.string().min(1),
})

export async function runPreview(
  _params: z.infer<typeof previewSchema>,
): Promise<{ pid: number }> {
  throw new Error('not implemented')
}
```

- [ ] **Write `mcp/webdev/src/index.ts`**

```typescript
import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js'
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js'
import { storybookSchema, runStorybook } from './tools/storybook.js'
import { playwrightSchema, runPlaywright } from './tools/playwright.js'
import { buildSchema, runBuild } from './tools/build.js'
import { previewSchema, runPreview } from './tools/preview.js'

const server = new McpServer({
  name: 'webdev',
  version: '0.1.0',
})

server.registerTool(
  'webdev_run_storybook',
  {
    description: 'Run Storybook test runner headlessly against the project',
    inputSchema: storybookSchema,
  },
  async (params) => {
    const result = await runStorybook(params)
    return { content: [{ type: 'text', text: JSON.stringify(result) }] }
  },
)

server.registerTool(
  'webdev_run_playwright',
  {
    description: 'Run Playwright tests with optional spec, test_name, or tag filter',
    inputSchema: playwrightSchema,
  },
  async (params) => {
    const result = await runPlaywright(params)
    return { content: [{ type: 'text', text: JSON.stringify(result) }] }
  },
)

server.registerTool(
  'webdev_build',
  {
    description: 'Run vite build --mode <mode>',
    inputSchema: buildSchema,
  },
  async (params) => {
    const result = await runBuild(params)
    return { content: [{ type: 'text', text: JSON.stringify(result) }] }
  },
)

server.registerTool(
  'webdev_preview',
  {
    description: 'Build with vite, serve with vite preview, take headless screenshot',
    inputSchema: previewSchema,
  },
  async (params) => {
    const result = await runPreview(params)
    return { content: [{ type: 'text', text: JSON.stringify(result) }] }
  },
)

const transport = new StdioServerTransport()
await server.connect(transport)
```

- [ ] **Build and verify the server starts**

```bash
cd mcp/webdev && npm run build && echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | node dist/index.js
```

Expected: JSON response listing `webdev_run_storybook`, `webdev_run_playwright`, `webdev_build`, `webdev_preview`.

- [ ] **Commit**

```bash
git add mcp/webdev/src/
git commit -m "feat(webdev-mcp): MCP server entry point and tool stubs"
```

---

### Task 4: `webdev_run_storybook` implementation

**Files:**
- Modify: `mcp/webdev/src/tools/storybook.ts`

**Consumes:** `spawnDetached` from `spawn.ts`, `buildLogPath` from `paths.ts`

**Script run:** `npx storybook test --headless` in `project_path`; completion marker appended by a wrapper shell script.

- [ ] **Write the wrapper script inline — update `mcp/webdev/src/tools/storybook.ts`**

```typescript
import { z } from 'zod'
import { spawnDetached } from '../lib/spawn.js'
import { buildLogPath } from '../lib/paths.js'
import { writeFileSync } from 'node:fs'
import { join } from 'node:path'
import os from 'node:os'

export const storybookSchema = z.object({
  project_path: z.string().min(1),
})

export async function runStorybook(
  params: z.infer<typeof storybookSchema>,
): Promise<{ pid: number }> {
  const { project_path } = params

  // Write a wrapper script that appends the exit marker after the job finishes
  const wrapper = join(os.tmpdir(), `couchdev-storybook-${Date.now()}.sh`)
  writeFileSync(
    wrapper,
    `#!/bin/sh
cd ${JSON.stringify(project_path)}
npx storybook test --headless
CODE=$?
if [ $CODE -eq 0 ]; then echo "COUCHDEV:EXIT:0"; else echo "COUCHDEV:EXIT:1"; fi
exit $CODE
`,
    { mode: 0o755 },
  )

  // We need the PID to build the log path, so we use a placeholder and rename
  // Instead: spawn first with a temp log, get PID, then build final log path
  // Simpler: build log path with a known PID after spawn
  const tempLog = join(os.tmpdir(), `couchdev-storybook-${Date.now()}.log`)
  const pid = await spawnDetached('/bin/sh', [wrapper], project_path, tempLog)

  const logPath = buildLogPath(project_path, pid, 'storybook')
  // Rename temp log to final path (process is already writing to tempLog)
  // Instead: write directly to final log. We can't know the PID before spawning.
  // Solution: spawn writes to DATA_ROOT/logs/<project>/pending.log,
  // then we move it. But that's racy.
  //
  // Correct approach: spawn with /dev/null, have the wrapper write to the final path itself.
  // Rebuild: pass the log path INTO the wrapper script.

  return { pid }
}
```

The above reveals a sequencing issue: the log path includes the PID, but we need the PID before we can build the log path. Fix: pass the data root and project name into the wrapper; the wrapper builds the path after it knows its own PID.

- [ ] **Rewrite `mcp/webdev/src/tools/storybook.ts` with self-locating wrapper**

```typescript
import { z } from 'zod'
import { spawnDetached } from '../lib/spawn.js'
import { DATA_ROOT, projectName } from '../lib/paths.js'
import { writeFileSync } from 'node:fs'
import { join } from 'node:path'
import os from 'node:os'

export const storybookSchema = z.object({
  project_path: z.string().min(1),
})

export async function runStorybook(
  params: z.infer<typeof storybookSchema>,
): Promise<{ pid: number }> {
  const { project_path } = params
  const proj = projectName(project_path)

  const wrapper = join(os.tmpdir(), `couchdev-storybook-${Date.now()}.sh`)
  writeFileSync(
    wrapper,
    `#!/bin/sh
PID=$$
LOG_DIR="${DATA_ROOT}/logs/${proj}/$PID"
mkdir -p "$LOG_DIR"
LOG="$LOG_DIR/storybook.log"
exec >> "$LOG" 2>&1
cd ${JSON.stringify(project_path)}
npx storybook test --headless
CODE=$?
if [ $CODE -eq 0 ]; then echo "COUCHDEV:EXIT:0"; else echo "COUCHDEV:EXIT:1"; fi
exit $CODE
`,
    { mode: 0o755 },
  )

  const pid = await spawnDetached('/bin/sh', [wrapper], project_path, '/dev/null')
  return { pid }
}
```

Note: the shell `$$` PID will differ from the Node-spawned PID (since `/bin/sh` is the direct child). Update `spawnDetached` to return the shell PID, which IS `$$` — this is consistent.

- [ ] **Verify build**

```bash
cd mcp/webdev && npm run build
```

Expected: no errors.

- [ ] **Commit**

```bash
git add mcp/webdev/src/tools/storybook.ts
git commit -m "feat(webdev-mcp): webdev_run_storybook implementation"
```

---

### Task 5: `webdev_run_playwright` implementation

**Files:**
- Modify: `mcp/webdev/src/tools/playwright.ts`

**Consumes:** `spawnDetached`, `DATA_ROOT`, `projectName`

- [ ] **Write `mcp/webdev/src/tools/playwright.ts`**

```typescript
import { z } from 'zod'
import { spawnDetached } from '../lib/spawn.js'
import { DATA_ROOT, projectName } from '../lib/paths.js'
import { writeFileSync } from 'node:fs'
import { join } from 'node:path'
import os from 'node:os'

export const playwrightSchema = z.object({
  project_path: z.string().min(1),
  spec: z.string().optional(),
  test_name: z.string().optional(),
  tag: z.string().optional(),
})

export async function runPlaywright(
  params: z.infer<typeof playwrightSchema>,
): Promise<{ pid: number }> {
  const { project_path, spec, test_name, tag } = params

  if (test_name && tag) {
    throw new Error('test_name and tag are mutually exclusive')
  }

  const proj = projectName(project_path)
  const grepValue = tag ? `@${tag}` : test_name
  const label = spec ?? tag ?? test_name ?? 'all'

  const pwArgs: string[] = ['playwright', 'test', '--reporter=line']
  if (spec) pwArgs.push(spec)
  if (grepValue) pwArgs.push('--grep', grepValue)
  pwArgs.push(`--screenshot=only-on-failure`)

  const wrapper = join(os.tmpdir(), `couchdev-playwright-${Date.now()}.sh`)
  writeFileSync(
    wrapper,
    `#!/bin/sh
PID=$$
LOG_DIR="${DATA_ROOT}/logs/${proj}/$PID"
SS_DIR="${DATA_ROOT}/screenshots/${proj}/$PID"
mkdir -p "$LOG_DIR" "$SS_DIR"
LOG="$LOG_DIR/playwright.log"
exec >> "$LOG" 2>&1
cd ${JSON.stringify(project_path)}
PLAYWRIGHT_SCREENSHOTS_PATH="$SS_DIR" npx ${pwArgs.map(a => JSON.stringify(a)).join(' ')}
CODE=$?
if [ $CODE -eq 0 ]; then echo "COUCHDEV:EXIT:0"; else echo "COUCHDEV:EXIT:1"; fi
exit $CODE
`,
    { mode: 0o755 },
  )

  const pid = await spawnDetached('/bin/sh', [wrapper], project_path, '/dev/null')
  return { pid }
}
```

- [ ] **Build and verify**

```bash
cd mcp/webdev && npm run build
```

- [ ] **Commit**

```bash
git add mcp/webdev/src/tools/playwright.ts
git commit -m "feat(webdev-mcp): webdev_run_playwright implementation"
```

---

### Task 6: `webdev_build` implementation

**Files:**
- Modify: `mcp/webdev/src/tools/build.ts`

- [ ] **Write `mcp/webdev/src/tools/build.ts`**

```typescript
import { z } from 'zod'
import { spawnDetached } from '../lib/spawn.js'
import { DATA_ROOT, projectName } from '../lib/paths.js'
import { writeFileSync } from 'node:fs'
import { join } from 'node:path'
import os from 'node:os'

export const buildSchema = z.object({
  project_path: z.string().min(1),
  mode: z.string().min(1),
})

export async function runBuild(
  params: z.infer<typeof buildSchema>,
): Promise<{ pid: number }> {
  const { project_path, mode } = params
  const proj = projectName(project_path)

  const wrapper = join(os.tmpdir(), `couchdev-build-${Date.now()}.sh`)
  writeFileSync(
    wrapper,
    `#!/bin/sh
PID=$$
LOG_DIR="${DATA_ROOT}/logs/${proj}/$PID"
mkdir -p "$LOG_DIR"
LOG="$LOG_DIR/build-${mode}.log"
exec >> "$LOG" 2>&1
cd ${JSON.stringify(project_path)}
npx vite build --mode ${JSON.stringify(mode)}
CODE=$?
if [ $CODE -eq 0 ]; then echo "COUCHDEV:EXIT:0"; else echo "COUCHDEV:EXIT:1"; fi
exit $CODE
`,
    { mode: 0o755 },
  )

  const pid = await spawnDetached('/bin/sh', [wrapper], project_path, '/dev/null')
  return { pid }
}
```

- [ ] **Build and verify**

```bash
cd mcp/webdev && npm run build
```

- [ ] **Commit**

```bash
git add mcp/webdev/src/tools/build.ts
git commit -m "feat(webdev-mcp): webdev_build implementation"
```

---

### Task 7: `webdev_preview` implementation

**Files:**
- Modify: `mcp/webdev/src/tools/preview.ts`

This tool: builds with vite, starts `vite preview` on a free port, takes a headless Chromium screenshot via Playwright, kills the preview server, writes the exit marker.

- [ ] **Write `mcp/webdev/src/tools/preview.ts`**

```typescript
import { z } from 'zod'
import { spawnDetached } from '../lib/spawn.js'
import { DATA_ROOT, projectName } from '../lib/paths.js'
import { writeFileSync } from 'node:fs'
import { join } from 'node:path'
import os from 'node:os'

export const previewSchema = z.object({
  project_path: z.string().min(1),
  mode: z.string().min(1),
})

export async function runPreview(
  params: z.infer<typeof previewSchema>,
): Promise<{ pid: number }> {
  const { project_path, mode } = params
  const proj = projectName(project_path)

  const wrapper = join(os.tmpdir(), `couchdev-preview-${Date.now()}.sh`)
  writeFileSync(
    wrapper,
    `#!/bin/sh
PID=$$
LOG_DIR="${DATA_ROOT}/logs/${proj}/$PID"
SS_DIR="${DATA_ROOT}/screenshots/${proj}/$PID"
mkdir -p "$LOG_DIR" "$SS_DIR"
LOG="$LOG_DIR/preview-${mode}.log"
SS="$SS_DIR/preview-${mode}.png"
exec >> "$LOG" 2>&1

cd ${JSON.stringify(project_path)}

# Build first
npx vite build --mode ${JSON.stringify(mode)}
if [ $? -ne 0 ]; then echo "COUCHDEV:EXIT:1"; exit 1; fi

# Start preview on a free port, capture the port
npx vite preview --port 0 &
PREVIEW_PID=$!
sleep 2

# Find the port vite preview bound to
PORT=$(ss -tlnp 2>/dev/null | grep $PREVIEW_PID | grep -oP ':\\K[0-9]+' | head -1)
if [ -z "$PORT" ]; then PORT=4173; fi

# Headless screenshot
node -e "
const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  await page.goto('http://localhost:$PORT');
  await page.waitForLoadState('networkidle');
  await page.screenshot({ path: '$SS', fullPage: true });
  await browser.close();
})().catch(e => { console.error(e); process.exit(1); });
"
SS_CODE=$?

kill $PREVIEW_PID 2>/dev/null

if [ $SS_CODE -eq 0 ]; then echo "COUCHDEV:EXIT:0"; else echo "COUCHDEV:EXIT:1"; fi
exit $SS_CODE
`,
    { mode: 0o755 },
  )

  const pid = await spawnDetached('/bin/sh', [wrapper], project_path, '/dev/null')
  return { pid }
}
```

- [ ] **Build and verify**

```bash
cd mcp/webdev && npm run build
```

- [ ] **Commit**

```bash
git add mcp/webdev/src/tools/preview.ts
git commit -m "feat(webdev-mcp): webdev_preview implementation"
```

---

### Task 8: install script

**Files:**
- Create: `mcp/webdev/setup/install.sh`

- [ ] **Write `mcp/webdev/setup/install.sh`**

```bash
#!/bin/sh
set -e

# Usage: install.sh <hub-ip> <nfs-export-path> <data-root>
# Example: install.sh 192.168.1.10 /srv/couchdev /mnt/couchdev

HUB_IP="${1:?Usage: install.sh <hub-ip> <nfs-export-path> <data-root>}"
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
ENV_FILE="/etc/environment"
if ! grep -q "^DATA_ROOT=" "$ENV_FILE" 2>/dev/null; then
  echo "DATA_ROOT=${DATA_ROOT}" >> "$ENV_FILE"
fi

echo "==> Registering MCP server with Claude Code"
SETTINGS="$HOME/.claude/settings.json"
if [ ! -f "$SETTINGS" ]; then
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

echo "==> Done. WebDev MCP node is ready."
echo "    DATA_ROOT: ${DATA_ROOT}"
echo "    MCP entry: webdev -> node ${SCRIPT_DIR}/dist/index.js"
```

- [ ] **Make executable**

```bash
chmod +x mcp/webdev/setup/install.sh
```

- [ ] **Commit**

```bash
git add mcp/webdev/setup/install.sh
git commit -m "feat(webdev-mcp): node install script"
```

---

### Task 9: end-to-end smoke test

Run the MCP server locally and call each tool against a real (or stub) project path to confirm the tools spawn, write logs, and return PIDs.

- [ ] **Create a temp project dir and DATA_ROOT for testing**

```bash
mkdir -p /tmp/couchdev-test/{projects/myapp,logs,screenshots}
mkdir -p /tmp/couchdev-test/projects/myapp
echo '{"name":"myapp","scripts":{"storybook":"echo stub-storybook"}}' > /tmp/couchdev-test/projects/myapp/package.json
```

- [ ] **Smoke test `webdev_run_storybook` via MCP protocol**

```bash
cd mcp/webdev
DATA_ROOT=/tmp/couchdev-test node dist/index.js <<'EOF'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"webdev_run_storybook","arguments":{"project_path":"/tmp/couchdev-test/projects/myapp"}}}
EOF
```

Expected: `{"result":{"content":[{"type":"text","text":"{\"pid\":<n>}"}]}}` — a positive integer PID.

- [ ] **Check log was written**

```bash
sleep 2
find /tmp/couchdev-test/logs -name "storybook.log" | head -1 | xargs cat
```

Expected: log output ending with `COUCHDEV:EXIT:0` or `COUCHDEV:EXIT:1`.

- [ ] **Smoke test `webdev_run_playwright` mutual exclusion**

```bash
cd mcp/webdev
DATA_ROOT=/tmp/couchdev-test node dist/index.js <<'EOF'
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"webdev_run_playwright","arguments":{"project_path":"/tmp/couchdev-test/projects/myapp","test_name":"foo","tag":"bar"}}}
EOF
```

Expected: error response containing `test_name and tag are mutually exclusive`.

- [ ] **Commit**

```bash
git add mcp/webdev/
git commit -m "feat(webdev-mcp): smoke tested, all tools wired"
```
