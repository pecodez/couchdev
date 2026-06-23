package project

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pecodez/couchdev/internal/git"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

type Handler struct {
	store       *Store
	projectsDir string
	git         git.Client
}

func NewHandler(store *Store, projectsDir string, g git.Client) *Handler {
	return &Handler{store: store, projectsDir: projectsDir, git: g}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /projects", h.create)
	mux.HandleFunc("GET /projects", h.list)
	mux.HandleFunc("DELETE /projects/{name}", h.delete)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var p Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if !validName.MatchString(p.Name) {
		http.Error(w, "name must start with a letter or number and contain only letters, numbers, - and _", http.StatusBadRequest)
		return
	}
	p.RepoPath = filepath.Join(h.projectsDir, p.Name, "src")

	switch p.SourceType {
	case "clone":
		if p.RepoURL == "" {
			http.Error(w, "repo_url required for clone", http.StatusBadRequest)
			return
		}
		p.Registry = detectRegistry(p.RepoURL)
		if err := h.git.Clone(p.RepoURL, p.RepoPath); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
	case "greenfield":
		if err := h.git.Init(p.RepoPath); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
	default:
		http.Error(w, "source_type must be 'clone' or 'greenfield'", http.StatusBadRequest)
		return
	}

	created, err := h.store.Create(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	h.enrich(created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	projects, err := h.store.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if projects == nil {
		projects = []Project{}
	}
	for i := range projects {
		h.enrich(&projects[i])
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	retainSource := r.URL.Query().Get("retain_source") == "true"

	p, err := h.store.GetByName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := h.store.Delete(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !retainSource {
		projectDir := filepath.Dir(p.RepoPath) // <projects_dir>/<name>
		os.RemoveAll(projectDir)
	}
	w.WriteHeader(http.StatusNoContent)
}

// enrich sets computed fields on p (source_missing, plans_dir).
func (h *Handler) enrich(p *Project) {
	if _, err := os.Stat(p.RepoPath); err != nil {
		p.SourceMissing = true
	}
	p.PlansDir = h.plansDir(p)
}

// plansDir returns the effective plans directory for p.
// When PlansPath is empty the hub owns plans outside src/;
// when set the path is relative to src/ enabling SCM commits.
func (h *Handler) plansDir(p *Project) string {
	if p.PlansPath == "" {
		return filepath.Join(h.projectsDir, p.Name, "plans")
	}
	return filepath.Join(p.RepoPath, p.PlansPath)
}

func detectRegistry(url string) string {
	switch {
	case strings.Contains(url, "github.com"):
		return "github"
	case strings.Contains(url, "gitlab.com"):
		return "gitlab"
	default:
		return "custom"
	}
}
