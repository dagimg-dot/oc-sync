package types

type Mapping struct {
	RemoteHostname  string `yaml:"remote_hostname"`
	RemoteProjectID string `yaml:"remote_project_id"`
	RemoteWorktree  string `yaml:"remote_worktree"`
	LocalWorktree   string `yaml:"local_worktree"`
	LocalProjectID  string `yaml:"local_project_id"`
}
