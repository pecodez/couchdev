# Couchdev — Claude Code rules

## Git workflow

- **One branch per feature/plan.** Each new feature, fix, or implementation plan gets its own branch, created fresh from main. Never accumulate unrelated work on an existing branch.
- **Always use a git worktree** for the new branch so it is isolated from the main workspace. Create the worktree before touching any file.
- Branch names must use a semantic prefix: `feat/`, `fix/`, `test/`, `docs/`, `refactor/`, `build/`, `chore/` followed by a kebab-case description.
- All commit messages must follow Conventional Commits: `type: short description`.
- Raise a PR from the branch to main once the work is complete.
- Never commit directly to main.

## Commits

- No `Co-Authored-By` or `Claude-Session` trailers.
- Terse messages — one line, present tense, lowercase after the colon.

## Planning

- Exiting plan mode does not mean the plan is approved. Always wait for an explicit go-ahead before implementing.

## Tests

- Do not use TDD (no write-failing-test-first). Only write test files when explicitly requested.

## Style

- No superpowers skills for implementation — write files and run commands directly.
- Do not look at or draw from the `glimpse` project; it is unrelated.
