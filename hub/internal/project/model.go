package project

type Project struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	RepoPath      string `json:"repo_path"`
	DefaultBranch string `json:"default_branch"`
	NamePrefix    string `json:"name_prefix"`
}
