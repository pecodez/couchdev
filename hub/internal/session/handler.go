package session

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type Handler struct {
	svc *Service
	log *zap.Logger
}

func NewHandler(svc *Service, log *zap.Logger) *Handler { return &Handler{svc: svc, log: log} }

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /projects/{project}/sessions", h.genesis)
	mux.HandleFunc("GET /sessions", h.list)
	mux.HandleFunc("GET /sessions/{project}/{session}", h.status)
	mux.HandleFunc("GET /sessions/{project}/{session}/changes", h.changes)
	mux.HandleFunc("DELETE /sessions/{project}/{session}", h.teardown)
	mux.HandleFunc("POST /sessions/{project}/{session}/resume", h.resume)
}

func (h *Handler) genesis(w http.ResponseWriter, r *http.Request) {
	h.log.Info("POST genesis", zap.String("project", r.PathValue("project")))
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
		h.log.Warn("genesis failed", zap.String("project", r.PathValue("project")), zap.String("session", req.Session), zap.Error(err))
		code := http.StatusInternalServerError
		msg := err.Error()
		if strings.Contains(msg, "not found") {
			code = http.StatusNotFound
		} else if strings.Contains(msg, "already live") {
			code = http.StatusConflict
		} else if strings.Contains(msg, "not authenticated") || strings.Contains(msg, "backoff") {
			code = http.StatusUnprocessableEntity
		}
		http.Error(w, msg, code)
		return
	}
	h.log.Info("genesis ok", zap.String("canonical", sess.CanonicalName), zap.Strings("warnings", sess.Warnings))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sess)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	h.log.Info("GET sessions")
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
	force := r.URL.Query().Get("force") == "true"
	h.log.Info("DELETE session", zap.String("project", r.PathValue("project")), zap.String("session", r.PathValue("session")), zap.Bool("force", force))
	deleted, reason, err := h.svc.Teardown(r.PathValue("project"), r.PathValue("session"), force)
	if err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		http.Error(w, err.Error(), code)
		return
	}
	if !deleted {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(struct {
			Reason string `json:"reason"`
		}{reason})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) resume(w http.ResponseWriter, r *http.Request) {
	h.log.Info("POST resume", zap.String("project", r.PathValue("project")), zap.String("session", r.PathValue("session")))
	sess, err := h.svc.Resume(r.PathValue("project"), r.PathValue("session"))
	if err != nil {
		h.log.Warn("resume failed", zap.Error(err))
		code := http.StatusInternalServerError
		msg := err.Error()
		if strings.Contains(msg, "not found") {
			code = http.StatusNotFound
		} else if strings.Contains(msg, "already live") {
			code = http.StatusConflict
		} else if strings.Contains(msg, "is dead") {
			code = http.StatusUnprocessableEntity
		} else if strings.Contains(msg, "not authenticated") || strings.Contains(msg, "backoff") {
			code = http.StatusUnprocessableEntity
		}
		http.Error(w, msg, code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sess)
}
