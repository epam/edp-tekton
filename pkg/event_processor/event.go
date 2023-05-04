package event_processor

import (
	"strings"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/pkg/apis/edp/v1"
)

// GerritEvent represents a Gerrit event.
type GerritEvent struct {
	Project struct {
		Name string `json:"name"`
	} `json:"project"`
	Change struct {
		Branch string `json:"branch"`
	} `json:"change"`
	Comment string `json:"comment"`
	Type    string `json:"type"`
}

// GitLabMergeRequestsEvent represents GitLab merge request event.
type GitLabMergeRequestsEvent struct {
	Project          GitLabProject      `json:"project"`
	ObjectAttributes GitLabMergeRequest `json:"object_attributes"`
}

type GitLabProject struct {
	PathWithNamespace string `json:"path_with_namespace"`
}

type GitLabMergeRequest struct {
	TargetBranch string       `json:"target_branch"`
	Title        string       `json:"title"`
	LastCommit   GitLabCommit `json:"last_commit"`
}

type GitLabCommit struct {
	ID string `json:"id"`
}

// GitLabCommentEvent represents GitLab comment event.
type GitLabCommentEvent struct {
	Project          GitLabProject      `json:"project"`
	MergeRequest     GitLabMergeRequest `json:"merge_request"`
	ObjectAttributes GitLabComment      `json:"object_attributes"`
}

type GitLabComment struct {
	Note string `json:"note"`
}

const (
	GitProviderGitHub = "github"
	GitProviderGitLab = "gitlab"
	GitProviderGerrit = "gerrit"

	GerritEventTypeCommentAdded = "comment-added"
	GitHubEventTypeCommentAdded = "issue_comment"
	GitLabEventTypeCommentAdded = "Note Hook"

	EventTypeReviewComment = "comment"
	EventTypeMerge         = "merge"

	recheckComment = "/recheck"
)

// EventInfo represents information about an event.
type EventInfo struct {
	GitProvider        string
	RepoPath           string
	Branch             string
	Type               string
	Codebase           *codebaseApi.Codebase
	HasPipelineRecheck bool
	PullRequest        *PullRequest
}

type PullRequest struct {
	HeadRef string `json:"headRef"`
	HeadSha string `json:"headSha"`
	Title   string `json:"title"`
}

// IsReviewCommentEvent returns true if the event is a review comment event.
func (e *EventInfo) IsReviewCommentEvent() bool {
	return e.Type == EventTypeReviewComment
}

// convertBranchName converts the branch name to the format used in the codebase for Kubernetes resource naming.
func convertBranchName(branch string) string {
	r := strings.NewReplacer("/", "-")

	return r.Replace(branch)
}

func containsPipelineRecheck(s string) bool {
	return strings.Contains(s, recheckComment)
}
