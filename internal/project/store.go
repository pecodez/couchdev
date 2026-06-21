package project

import (
	"database/sql"
	"fmt"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) Create(p Project) (*Project, error) {
	if p.DefaultBranch == "" {
		p.DefaultBranch = "main"
	}
	res, err := s.db.Exec(
		`INSERT INTO projects (name, repo_path, default_branch, name_prefix) VALUES (?,?,?,?)`,
		p.Name, p.RepoPath, p.DefaultBranch, p.NamePrefix,
	)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	id, _ := res.LastInsertId()
	p.ID = id
	return &p, nil
}

func (s *Store) List() ([]Project, error) {
	rows, err := s.db.Query(
		`SELECT id, name, repo_path, default_branch, name_prefix FROM projects ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.RepoPath, &p.DefaultBranch, &p.NamePrefix); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetByName(name string) (*Project, error) {
	var p Project
	err := s.db.QueryRow(
		`SELECT id, name, repo_path, default_branch, name_prefix FROM projects WHERE name=?`, name,
	).Scan(&p.ID, &p.Name, &p.RepoPath, &p.DefaultBranch, &p.NamePrefix)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project %q not found", name)
	}
	return &p, err
}
