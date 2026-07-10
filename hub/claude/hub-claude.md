# Couchdev Hub — Session Rules

This file is installed by couchdev into the projects root and applies to every
session the hub creates. It is intentionally opinionated. A CLAUDE.md inside
your project repo can refine or override these rules for that project.

---

## Session identity

Your session canonical name (`<project>/<session>`) was passed via `--rc` when
this session was created. Confirm it with:

    git branch --show-current

Your working directory is a git worktree on that branch. You work on this
project and this branch only. Your tmux session follows the pattern
`cdv.<project>.<session>`.

---

## One session, one purpose

Every session is scoped to **exactly one feature or bug fix**. The session name
is your scope boundary. You do not take on unrelated work within the same
session — see the Scope gate section below.

---

## Session startup protocol

Run this at the start of every session before doing anything else:

1. Read `git branch --show-current` to confirm your branch/session name.
2. Derive your plan file path: `../../plans/<session-name>.md` relative to your
   working directory. If the session name contains `/` (e.g. `feat/fix-login`),
   preserve it as a path (→ `../../plans/feat/fix-login.md`).
3. **No plan file found** → create one using the Plan format below. Do not
   write code, edit files, or run commands until the user has explicitly
   approved the plan.
4. **Plan file exists, status `planning`** → show the plan, ask for approval.
   Do not start work until the user gives an explicit go-ahead.
5. **Plan file exists, status `approved` or `in-progress`** → summarise the
   plan in one sentence and confirm you are ready to continue.

---

## Plan format

Store the plan at `../../plans/<session-name>.md` from your working directory:

```markdown
# <canonical session name>

## Purpose
One sentence: what feature is being added or what bug is being fixed.

## Scope
**In scope:** what this session will deliver.
**Out of scope:** what it will not do (list anything the user might reasonably
expect but that is excluded).

## Acceptance criteria
- [ ] Criterion 1
- [ ] Criterion 2

## Approach
High-level implementation notes. Filled in during planning, before approval.

## Status
planning
```

Progress notes:
- Change `planning` → `approved` when the user approves.
- Change `approved` → `in-progress` when you begin implementation.
- Tick criteria (`[x]`) as they are met.
- Change `in-progress` → `complete` when all criteria are met and a PR is raised.

---

## Scope gate

Before acting on any user request that involves writing, editing, running, or
configuring anything, ask: **"Is this directly required to deliver the purpose
in my plan?"**

If the answer is **no** or **uncertain**:

1. State the out-of-scope request explicitly.
2. Offer the user two options:

   > **Option A — Update the plan:** I'll revise the plan to include this work.
   > You'll need to re-approve it before I continue.
   >
   > **Option B — New session:** This belongs in a separate session. Create one
   > in the couchdev hub, then I can pick it up there.

3. Do not proceed until the user has chosen. If they choose Option A, update
   the plan file and wait for explicit re-approval before continuing.

Simple questions, status checks, and read-only exploration do not trigger the
scope gate.

---

## Git rules

- Work only on the branch already checked out in your worktree.
- Never commit to the default branch directly.
- Commit messages: Conventional Commits — `type: short description`.
- No `Co-Authored-By` or session-attribution trailers.
- Raise a PR when all acceptance criteria are met. Reference the plan purpose
  in the PR description.

---

## Secrets

Never expose, log, or echo bearer tokens or credentials. Treat any token found
in environment variables or config files as a secret.

---

## MCP

Available MCP servers are listed in `~/.claude/mcp.json` (if present). Use
only tools relevant to the current project; do not assume a tool exists without
checking the available tool list.
