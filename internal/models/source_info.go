package models

// SourceInfo locates an application's source code, both on disk and in git.
type SourceInfo struct {
	LocalPath        string `json:"local_path,omitempty"`
	RepositoryURL    string `json:"repository_url,omitempty"`
	MainBranch       string `json:"main_branch,omitempty"`
	WorkingDirectory string `json:"working_directory,omitempty"`
}
