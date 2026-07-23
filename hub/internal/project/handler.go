package project

import (
	"bufio"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	enry "github.com/go-enry/go-enry/v2"
	"github.com/pecodez/couchdev/internal/git"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
var mdBlockquote = regexp.MustCompile(`^\s*>\s*`)
var mdLink = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
var mdEmphasis = regexp.MustCompile(`\*\*|__|\*|_|` + "`")

func stripMarkdown(s string) string {
	s = mdBlockquote.ReplaceAllString(s, "")
	s = mdLink.ReplaceAllString(s, "$1")
	s = mdEmphasis.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

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
	mux.HandleFunc("POST /projects/{name}/remote", h.connectRemote)
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
	p.RepoPath = filepath.Join(h.projectsDir, p.Name, "source")

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

// connectRemote attaches an existing remote repo URL to a local-only project
// and pushes its default branch. If the push fails after the remote was
// successfully attached, the attachment is still persisted and a warning is
// returned rather than the failure being treated as a full rollback.
func (h *Handler) connectRemote(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	var body struct {
		RepoURL string `json:"repo_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.RepoURL == "" {
		http.Error(w, "repo_url required", http.StatusBadRequest)
		return
	}

	p, err := h.store.GetByName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if p.RepoURL != "" {
		http.Error(w, "project already has a remote configured", http.StatusConflict)
		return
	}

	if err := h.git.AddRemote(p.RepoPath, "origin", body.RepoURL); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	registry := detectRegistry(body.RepoURL)

	var warning string
	if err := h.git.Push(p.RepoPath, "origin", p.DefaultBranch); err != nil {
		warning = "remote attached but push failed: " + err.Error()
	}

	if err := h.store.SetRemote(name, body.RepoURL, registry); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p.RepoURL = body.RepoURL
	p.Registry = registry
	h.enrich(p)

	resp := struct {
		*Project
		Warning string `json:"warning,omitempty"`
	}{Project: p, Warning: warning}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// enrich sets computed fields on p (source_missing, plans_dir, description, languages).
func (h *Handler) enrich(p *Project) {
	if _, err := os.Stat(p.RepoPath); err != nil {
		p.SourceMissing = true
		p.PlansDir = h.plansDir(p)
		return
	}
	p.PlansDir = h.plansDir(p)
	p.Description = readmeDescription(p.RepoPath)
	p.Languages = detectLanguages(p.RepoPath)
}

// readmeDescription returns the first real paragraph line from a README,
// skipping headings, badges, and image lines.
func readmeDescription(repoPath string) string {
	for _, name := range []string{"README.md", "README", "readme.md", "Readme.md"} {
		f, err := os.Open(filepath.Join(repoPath, name))
		if err != nil {
			continue
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "#") {
				continue
			}
			// skip badge/image lines
			if strings.HasPrefix(line, "![") || strings.HasPrefix(line, "[![") {
				continue
			}
			return stripMarkdown(line)
		}
	}
	return ""
}

// detectLanguages walks the repo and returns a sorted list of unique languages (max 8).
func detectLanguages(repoPath string) []string {
	seen := map[string]struct{}{}
	filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(repoPath, path)
		if enry.IsVendor(rel) || enry.IsDocumentation(rel) || enry.IsTest(rel) {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil || enry.IsBinary(content) || enry.IsGenerated(rel, content) {
			return nil
		}
		if lang := enry.GetLanguage(rel, content); lang != "" {
			seen[lang] = struct{}{}
		}
		return nil
	})
	langs := make([]string, 0, len(seen))
	for l := range seen {
		langs = append(langs, l)
	}
	sort.Strings(langs)
	if len(langs) > 8 {
		langs = langs[:8]
	}
	return langs
}

// plansDir returns the effective plans directory for p.
// When PlansPath is empty the hub owns plans outside source/;
// when set the path is relative to source/ enabling SCM commits.
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
