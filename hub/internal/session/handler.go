package session

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /projects/{project}/sessions", h.genesis)
	mux.HandleFunc("GET /sessions", h.list)
	mux.HandleFunc("GET /sessions/{project}/{session}", h.status)
	mux.HandleFunc("GET /sessions/{project}/{session}/changes", h.changes)
	mux.HandleFunc("DELETE /sessions/{project}/{session}", h.teardown)
}

func (h *Handler) genesis(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Session string `json:"session"`
		CWD     string `json:"cwd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Session == "" {
		http.Error(w, "session required", http.StatusBadRequest)
		return
	}
	sess, err := h.svc.Genesis(r.PathValue("project"), req.Session, req.CWD)
	if err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		} else if strings.Contains(err.Error(), "already live") {
			code = http.StatusConflict
		}
		http.Error(w, err.Error(), code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sess)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.svc.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if sessions == nil {
		sessions = []Session{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	s, err := h.svc.Status(r.PathValue("project"), r.PathValue("session"))
	if err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		http.Error(w, err.Error(), code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func (h *Handler) changes(w http.ResponseWriter, r *http.Request) {
	project, sessionName := r.PathValue("project"), r.PathValue("session")
	ahead, files, err := h.svc.Changes(project, sessionName)
	if err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		http.Error(w, err.Error(), code)
		return
	}
	if files == nil {
		files = []string{}
	}

	// look up branch name from the session record
	sess, _ := h.svc.Status(project, sessionName)
	branch := ""
	if sess != nil {
		branch = sess.Branch
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Branch       string   `json:"branch"`
		Ahead        int      `json:"ahead"`
		ChangedFiles []string `json:"changed_files"`
	}{branch, ahead, files})
}

func (h *Handler) teardown(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Teardown(r.PathValue("project"), r.PathValue("session")); err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		http.Error(w, err.Error(), code)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
