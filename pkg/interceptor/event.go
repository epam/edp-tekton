package interceptor

// GerritEvent represents a Gerrit event.
type GerritEvent struct {
	Project struct {
		Name string `json:"name"`
	} `json:"project"`
	Change struct {
		Branch string `json:"branch"`
	} `json:"change"`
}

// GitHubEvent represents a GitHub event.
type GitHubEvent struct {
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	PullRequest struct {
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
	} `json:"pull_request"`
}

// GitLabEvent represents GitLab event.
type GitLabEvent struct {
	Project struct {
		PathWithNamespace string `json:"path_with_namespace"`
	} `json:"project"`
	ObjectAttributes struct {
		TargetBranch string `json:"target_branch"`
	} `json:"object_attributes"`
}

const (
	gitProviderGitHub = "github"
	gitProviderGitLab = "gitlab"
	gitProviderGerrit = "gerrit"
)

type eventInfo struct {
	GitProvider string
	RepoPath    string
	Branch      string
}
