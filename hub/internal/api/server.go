package api

import (
	"database/sql"
	"encoding/json"
	"io/fs"
	"net/http"

	"go.uber.org/zap"

	"github.com/pecodez/couchdev/internal/auth"
	"github.com/pecodez/couchdev/internal/git"
	"github.com/pecodez/couchdev/internal/project"
	"github.com/pecodez/couchdev/internal/session"
	"github.com/pecodez/couchdev/internal/tmux"
)

func New(tokenHash []byte, db *sql.DB, t tmux.Client, webFS fs.FS, projectsDir string, g git.Client, log *zap.Logger, version string) http.Handler {
	ps := project.NewStore(db)
	ss := session.NewStore(db)
	svc := session.NewService(ps, ss, t, g, log)

	apiMux := http.NewServeMux()
	project.NewHandler(ps, projectsDir, g).Register(apiMux)
	session.NewHandler(svc, log).Register(apiMux)

	apiHandler := http.Handler(http.StripPrefix("/api", apiMux))
	if tokenHash != nil {
		apiHandler = auth.Middleware(tokenHash)(apiHandler)
	}

	root := http.NewServeMux()
	root.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"version": version})
	})
	root.Handle("/api/", apiHandler)
	root.Handle("/", http.FileServer(http.FS(webFS)))
	return root
}
