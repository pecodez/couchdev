package project

type Project struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	RepoPath      string `json:"repo_path"`
	DefaultBranch string `json:"default_branch"`
	NamePrefix    string `json:"name_prefix"`
	SourceType    string `json:"source_type"`
	RepoURL       string `json:"repo_url"`
	Registry      string `json:"registry"`
	PlansPath     string `json:"plans_path,omitempty"`

	// Computed at read time; not stored in DB.
	SourceMissing bool   `json:"source_missing,omitempty"`
	PlansDir      string `json:"plans_dir,omitempty"`
}
