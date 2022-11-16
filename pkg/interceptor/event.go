package interceptor

// GerritEvent represents a Gerrit event.
type GerritEvent struct {
	Project struct {
		Name string `json:"name"`
	} `json:"project"`
}

// GitHubEvent represents a GitHub event.
type GitHubEvent struct {
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// GitLabEvent represents GitLab event.
type GitLabEvent struct {
	Project struct {
		PathWithNamespace string `json:"path_with_namespace"`
	} `json:"project"`
}

const (
	gitProviderGitHub = "github"
	gitProviderGitLab = "gitlab"
	gitProviderGerrit = "gerrit"
)

type eventInfo struct {
	GitProvider string
	RepoPath    string
}
