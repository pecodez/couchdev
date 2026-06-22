package project

import (
	"encoding/json"
	"net/http"
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
	p.RepoPath = filepath.Join(h.projectsDir, p.Name)

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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
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
