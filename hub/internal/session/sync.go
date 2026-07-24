package session

import (
	"go.uber.org/zap"

	"github.com/pecodez/couchdev/internal/git"
	"github.com/pecodez/couchdev/internal/project"
)

// SyncAllProjects fetches the default branch for every registered project
// that has an "origin" remote, so merge/ahead state checks start from
// up-to-date remote-tracking refs rather than whatever was last fetched
// (possibly never). Intended to run once at server startup. Failures are
// logged per project and do not stop the sweep or abort startup — a project
// whose remote is unreachable just falls back to comparing against its local
// ref, as it did before this fetch existed.
func SyncAllProjects(ps *project.Store, g git.Client, log *zap.Logger) {
	projects, err := ps.List()
	if err != nil {
		log.Warn("sync: failed to list projects", zap.Error(err))
		return
	}
	for _, proj := range projects {
		hasRemote, err := g.HasRemote(proj.RepoPath, "origin")
		if err != nil {
			log.Warn("sync: check remote failed", zap.String("project", proj.Name), zap.Error(err))
			continue
		}
		if !hasRemote {
			continue
		}
		if err := g.Fetch(proj.RepoPath, "origin", proj.DefaultBranch); err != nil {
			log.Warn("sync: fetch failed", zap.String("project", proj.Name), zap.Error(err))
			continue
		}
		log.Info("sync: fetched default branch", zap.String("project", proj.Name), zap.String("branch", proj.DefaultBranch))
	}
}
