# WebDev MCP Node — Design Spec

## Overview

The WebDev node is a stateless stdio MCP server that executes web development jobs on behalf of the hub. It is one node in the CouchDev distributed remote development cluster. Its sole responsibility is to run jobs — it owns no project state, no session state, and no routing logic. The hub is the manager; the node is a dumb executor.

The node is designed to run on minimal hardware (including a Pi Zero) and incurs zero server overhead by using stdio as its MCP transport.

---

## Cluster Architecture Context

```
Hub (Claude Code)
  │
  ├── NFS export: /mnt/couchdev/projects/<project>  (source, read/write)
  ├── NFS export: /mnt/couchdev/logs/<project>      (job logs, written by node)
  └── NFS export: /mnt/couchdev/screenshots/<project> (screenshots, written by node)
        │
        └── WebDev Node (stdio MCP)
              ├── mounts /mnt/couchdev at boot (via /etc/fstab)
              └── DATA_ROOT = /mnt/couchdev
```

The hub manages all NFS exports. Nodes mount the share at boot. No coordination protocol beyond the filesystem.

---

## Transport

**stdio** — Claude connects to the node over SSH and communicates with the MCP process via stdin/stdout. No persistent server process, no open ports, no connection management. The MCP process starts on demand and exits when the connection closes.

---

## NFS Mount Convention

All paths derive from a single `DATA_ROOT` configured on the node at install time (default: `/mnt/couchdev`).

| Purpose | Path |
|---------|------|
| Project source | `DATA_ROOT/projects/<project>/` |
| Job logs | `DATA_ROOT/logs/<project>/<pid>/<tool>.log` |
| Screenshots | `DATA_ROOT/screenshots/<project>/<pid>/<name>.png` |

`<project>` is derived by the node as `basename(project_path)`. PID subdirectories ensure multiple concurrent jobs on the same project never collide. The hub derives all artifact paths from `project_name + pid` — no path is returned over MCP.

---

## Fire & Redirect Execution Pattern

Every tool call:

1. Receives `project_path` and tool-specific parameters
2. Spawns the job as a detached background process (`nohup`, stdout/stderr → log file)
3. Returns `{ "pid": <n> }` immediately
4. The background process writes verbose output to the log file
5. The background process writes a terminal marker as its final line:
   - `COUCHDEV:EXIT:0` — success
   - `COUCHDEV:EXIT:1` — failure

The hub's sub-agent tails the log file and watches for the terminal marker. It never pulls the raw log into the hub's context — it extracts only the summary or failing output.

---

## Tools

All tools are prefixed `webdev_`. All tools return `{ "pid": <n> }` immediately. All tools accept `project_path` as their first parameter — the absolute path to the project source on the NFS mount.

### `webdev_run_storybook`

Runs the Storybook test runner in headless mode against the project.

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `project_path` | string | yes | Absolute path to project source on NFS mount |

**Execution:** `npx storybook test --headless`

**Artifacts:**
- Log: `DATA_ROOT/logs/<project>/<pid>/storybook.log`
- Screenshots: `DATA_ROOT/screenshots/<project>/<pid>/storybook-<story-id>.png` (one per failing story)

---

### `webdev_run_playwright`

Runs Playwright tests with optional filters. All filters are additive — unset filters are omitted from the command.

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `project_path` | string | yes | Absolute path to project source on NFS mount |
| `spec` | string | no | Path to a specific spec file (relative to project root) |
| `test_name` | string | no | Test name pattern passed to `--grep` |
| `tag` | string | no | Tag filter passed to `--grep` as `@<tag>`; mutually exclusive with `test_name` |

**Execution:** `npx playwright test [spec] [--grep <test_name or @tag>]`

If both `test_name` and `tag` are provided, the tool returns an error without spawning a process.

**Artifacts:**
- Log: `DATA_ROOT/logs/<project>/<pid>/playwright.log`
- Screenshots: `DATA_ROOT/screenshots/<project>/<pid>/playwright-<spec-or-tag>.png` — Playwright's built-in `screenshot: 'only-on-failure'` captures these automatically; the node configures this via `playwright.config.ts` or CLI flag

---

### `webdev_build`

Runs a Vite production build with the specified mode.

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `project_path` | string | yes | Absolute path to project source on NFS mount |
| `mode` | string | yes | Vite mode flag (e.g., `production`, `staging`) |

**Execution:** `vite build --mode <mode>`

**Artifacts:**
- Log: `DATA_ROOT/logs/<project>/<pid>/build-<mode>.log`
- Build output written to project's configured `outDir` on the NFS mount

---

### `webdev_preview`

Builds with Vite, serves the output with `vite preview`, takes a headless browser screenshot, then exits.

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `project_path` | string | yes | Absolute path to project source on NFS mount |
| `mode` | string | yes | Vite mode flag passed to both build and preview |

**Execution:**
1. `vite build --mode <mode>`
2. `vite preview` (binds to a free local port)
3. Headless Chromium navigates to the preview URL, captures screenshot
4. `vite preview` process killed
5. Terminal marker written

**Artifacts:**
- Log: `DATA_ROOT/logs/<project>/<pid>/preview-<mode>.log`
- Screenshot: `DATA_ROOT/screenshots/<project>/<pid>/preview-<mode>.png`

---

## Node Configuration

Configured once at install time via a `.env` file or environment variables. No runtime config, no registry, no project state.

| Variable | Default | Description |
|----------|---------|-------------|
| `DATA_ROOT` | `/mnt/couchdev` | Root of the NFS mount |

---

## Implementation Stack

- **Runtime:** Node.js (LTS)
- **Language:** TypeScript
- **MCP SDK:** `@modelcontextprotocol/sdk`
- **Headless browser:** Playwright's bundled Chromium (reused for both `webdev_preview` screenshots and `webdev_run_playwright`)
- **Package manager:** npm

---

## Directory Layout

```
mcp/webdev/
├── src/
│   ├── index.ts          # MCP server entry point, tool registration
│   ├── tools/
│   │   ├── storybook.ts  # webdev_run_storybook
│   │   ├── playwright.ts # webdev_run_playwright
│   │   ├── build.ts      # webdev_build
│   │   └── preview.ts    # webdev_preview
│   └── lib/
│       ├── spawn.ts      # detached process helper, log path derivation
│       └── paths.ts      # DATA_ROOT convention, path builders
├── package.json
├── tsconfig.json
└── setup/
    └── install.sh        # wires NFS mount, installs deps, registers MCP with Claude
```

---

## Setup Script Responsibilities (`install.sh`)

1. Install Node.js LTS if not present
2. `npm install` in `mcp/webdev/`
3. Install Playwright browsers: `npx playwright install chromium`
4. Write `/etc/fstab` entry for the NFS mount
5. Mount `DATA_ROOT`
6. Write `DATA_ROOT` to node's environment
7. Register the MCP server with Claude Code (`~/.claude/settings.json` or equivalent)

The script is the only thing that knows about the cluster topology. The MCP server itself is topology-agnostic.
