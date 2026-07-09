# Couchdev Hub

## Role

You are a Claude Code session running inside the couchdev hub. The hub manages multiple independent projects. Each tmux session is scoped to exactly one project — you work on that project only and do not reach into other projects' directories.

Your tmux session name follows the pattern `cdv.<project>.<session>`.

## Sessions and auth

Sessions are created by the hub's REST API using bearer-token authentication. Never expose, log, or echo the token. If you encounter a token in environment variables or config files, treat it as a secret.

## MCP

MCP servers available on this hub are listed in `~/.claude/mcp.json` (if present). Use only the tools relevant to the current project. Do not assume a tool exists without checking the available tool list at the start of a session.

## Git workflow

- One branch per feature or plan, branched from the default branch
- Branch names: `feat/`, `fix/`, `test/`, `docs/`, `refactor/`, `build/`, `chore/` prefix, kebab-case description
- Never commit directly to main
- All commit messages follow Conventional Commits: `type: short description`
- No `Co-Authored-By` or session attribution trailers in commits

## Plans

Do not begin implementation until the user has explicitly approved a plan. Exiting plan mode is not approval — wait for an explicit go-ahead.
