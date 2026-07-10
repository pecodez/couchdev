package api

import (
	"database/sql"
	"io/fs"
	"net/http"

	"github.com/pecodez/couchdev/internal/auth"
	"github.com/pecodez/couchdev/internal/git"
	"github.com/pecodez/couchdev/internal/project"
	"github.com/pecodez/couchdev/internal/session"
	"github.com/pecodez/couchdev/internal/tmux"
)

func New(tokenHash []byte, db *sql.DB, t tmux.Client, webFS fs.FS, projectsDir string, g git.Client) http.Handler {
	ps := project.NewStore(db)
	ss := session.NewStore(db)
	svc := session.NewService(ps, ss, t, g)

	apiMux := http.NewServeMux()
	project.NewHandler(ps, projectsDir, g).Register(apiMux)
	session.NewHandler(svc).Register(apiMux)

	root := http.NewServeMux()
	root.Handle("/api/", auth.Middleware(tokenHash)(http.StripPrefix("/api", apiMux)))
	root.Handle("/", http.FileServer(http.FS(webFS)))
	return root
}
