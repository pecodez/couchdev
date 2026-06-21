package session

import (
	"database/sql"
	"fmt"
	"time"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) Create(sess Session) (*Session, error) {
	res, err := s.db.Exec(`
		INSERT INTO sessions (project_id, session, canonical_name, passed_name, cwd, tmux_name, pid)
		VALUES (?,?,?,?,?,?,?)`,
		sess.ProjectID, sess.Session, sess.CanonicalName, sess.PassedName,
		sess.CWD, sess.TmuxName, sess.PID,
	)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	id, _ := res.LastInsertId()
	sess.ID = id
	return &sess, nil
}

func (s *Store) GetByCanonical(canonical string) (*Session, error) {
	var sess Session
	var started string
	var pid sql.NullInt64
	err := s.db.QueryRow(`
		SELECT id, project_id, session, canonical_name, passed_name, cwd, tmux_name, pid, killed, started_at
		FROM sessions WHERE canonical_name=?`, canonical,
	).Scan(&sess.ID, &sess.ProjectID, &sess.Session, &sess.CanonicalName,
		&sess.PassedName, &sess.CWD, &sess.TmuxName, &pid, &sess.Killed, &started)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session %q not found", canonical)
	}
	if err != nil {
		return nil, err
	}
	if pid.Valid {
		v := int(pid.Int64)
		sess.PID = &v
	}
	sess.StartedAt, _ = time.Parse("2006-01-02 15:04:05", started)
	return &sess, nil
}

func (s *Store) ListAll() ([]Session, error) {
	rows, err := s.db.Query(`
		SELECT id, project_id, session, canonical_name, passed_name, cwd, tmux_name, pid, killed, started_at
		FROM sessions ORDER BY started_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Session
	for rows.Next() {
		var sess Session
		var started string
		var pid sql.NullInt64
		if err := rows.Scan(&sess.ID, &sess.ProjectID, &sess.Session, &sess.CanonicalName,
			&sess.PassedName, &sess.CWD, &sess.TmuxName, &pid, &sess.Killed, &started); err != nil {
			return nil, err
		}
		if pid.Valid {
			v := int(pid.Int64)
			sess.PID = &v
		}
		sess.StartedAt, _ = time.Parse("2006-01-02 15:04:05", started)
		out = append(out, sess)
	}
	return out, rows.Err()
}

func (s *Store) MarkKilled(canonical string) error {
	_, err := s.db.Exec(`UPDATE sessions SET killed=1 WHERE canonical_name=?`, canonical)
	return err
}
