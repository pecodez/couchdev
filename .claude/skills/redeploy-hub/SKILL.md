---
name: redeploy-hub
description: Rebuild the couchdev hub binary from current worktree source and hot-swap it into the live, non-systemd server process running on this machine — for verifying local changes without cutting a GitHub release. Use when the user asks to "redeploy", "rebuild and restart the server", "hot-swap the binary", or wants to see their latest hub changes running live.
---

# Redeploy hub

Rebuilds the `couchdev` binary from the current worktree's source and
replaces the binary backing the live hub server process on this machine,
then restarts it. This is for local dev iteration only — it is not a
replacement for `make release-*` / GitHub Releases, and the process is not
supervised (no systemd, no crash auto-restart).

**This restarts a shared process.** The hub server backs every active
Claude Code session on this machine (tmux sessions persist independently
and are not killed by this), but the web/mobile UI and any in-flight
`/api/` requests will briefly disconnect during the restart. Always get
explicit confirmation from the user immediately before step 4 (stopping the
live process) — even if they already said "go ahead" on the rebuild in
general, confirm the specific restart action separately.

## Steps

1. **Build.** From the repo root:
   ```
   cd hub && make build
   ```
   This builds the frontend (`web/dist`, embedded via `go:embed`) and then
   the Go binary with `VERSION`/`COMMIT` stamped via ldflags (see
   `hub/Makefile`), producing `hub/bin/couchdev`.

2. **Locate the live process.**
   ```
   pgrep -af 'couchdev serve'
   ```
   Extract the PID and the `--config`/`-c` path from the matched command
   line — don't guess these, they can differ between machines/installs. If
   nothing matches, stop and tell the user no live server was found to
   replace.

3. **Sanity-check the new binary before touching anything live.**
   ```
   hub/bin/couchdev version
   ```
   Confirm it's a real, successful build (prints a version/commit string,
   doesn't error).

4. **Confirm with the user**, stating plainly: "This will restart the live
   hub server backing all active sessions on this machine — the UI will be
   briefly unreachable during the restart. Proceed?" Wait for an explicit
   yes before continuing.

5. **Back up the current live binary** for rollback:
   ```
   cp <live-bin-path> <live-bin-path>.bak
   ```

6. **Stop the running process** gracefully:
   ```
   kill <PID>
   ```
   Poll `kill -0 <PID>` until it fails (process gone) — give it a few
   seconds. Do not reach for `kill -9` unless a plain `kill` fails to stop
   it after a reasonable wait.

7. **Replace the binary:**
   ```
   cp hub/bin/couchdev <live-bin-path>
   chmod +x <live-bin-path>
   ```

8. **Restart it detached**, using the same config path found in step 2, so
   it survives after your shell exits:
   ```
   nohup <live-bin-path> serve --config <config-path> > <live-log-path> 2>&1 &
   disown
   ```

9. **Verify.** Poll until it responds (a few seconds is normal):
   ```
   curl -sS http://127.0.0.1:<port>/api/version
   ```
   (host/port come from the config's `listen_addr`). `/api/version` is
   reachable without a bearer token by design even when `require_auth` is
   on (see `hub/internal/api/server.go`), so no token juggling is needed.
   Confirm the returned version/commit matches the binary built in step 3.

10. **On any failure** in steps 6–9 (process won't stop, won't come back
    up, `/api/version` doesn't respond after a reasonable wait), restore
    immediately from the backup and restart it the same way as step 8, then
    report the failure and what you rolled back.

## Notes

- Never hardcode the live binary/config/port — always derive them from the
  running process (step 2) and its config file.
- If the rebuild only touched frontend or backend code, `make build` still
  rebuilds both — keep it simple, don't try to skip half the build.
