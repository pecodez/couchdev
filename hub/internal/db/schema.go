package db

const schema = `
CREATE TABLE IF NOT EXISTS projects (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    name           TEXT    UNIQUE NOT NULL,
    repo_path      TEXT    NOT NULL,
    default_branch TEXT    NOT NULL DEFAULT 'main',
    name_prefix    TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS sessions (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id     INTEGER NOT NULL REFERENCES projects(id),
    session        TEXT    NOT NULL,
    canonical_name TEXT    UNIQUE NOT NULL,
    passed_name    TEXT    NOT NULL,
    cwd            TEXT    NOT NULL,
    tmux_name      TEXT    UNIQUE NOT NULL,
    pid            INTEGER,
    killed         INTEGER NOT NULL DEFAULT 0,
    started_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, session)
);
`

// migrations are applied in order, gated by PRAGMA user_version.
var migrations = []string{
	// v1: add source type, remote URL, and registry metadata to projects
	`ALTER TABLE projects ADD COLUMN source_type TEXT NOT NULL DEFAULT 'clone'`,
	`ALTER TABLE projects ADD COLUMN repo_url    TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE projects ADD COLUMN registry    TEXT NOT NULL DEFAULT ''`,
}
