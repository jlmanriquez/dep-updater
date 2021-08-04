package cfg

// Config is a json file configuration
type Config struct {
	// WorkspaceHome is a path to local repository, with al projects
	WorkspaceHome string `json:"workspace_home,omitempty"`
	// Repository contain the info of remote repository
	Repository Repository `json:"repository,omitempty"`
	// Projects is a list de project to update
	Projects []Project `json:"projects,omitempty"`
	// BranchSection specifies how work with project branches
	// WorkingBranch is the branch name over to work
	WorkingBranch string `json:"working_branch,omitempty"`
	// CreateFrom is the name of the branch from where WorkingBranch will be created
	CreateFrom string `json:"create_from,omitempty"`
	// Libraries contain all libraries and versions to update
	Libraries map[string]string `json:"libraries,omitempty"`
}

type Project struct {
	Enabled bool   `json:"enabled,omitempty"`
	Name    string `json:"name,omitempty"`
	Push    bool   `json:"push,omitempty"`
	Clone   bool   `json:"clone,omitempty"`
}

type Repository struct {
	URL      string `json:"url,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}
