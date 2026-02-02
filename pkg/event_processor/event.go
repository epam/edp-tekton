package event_processor

import (
	"strings"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"
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
	User             GitLabUser         `json:"user"`
}

type GitLabProject struct {
	PathWithNamespace string `json:"path_with_namespace"`
}

type GitLabUser struct {
	Username  string `json:"username"`
	AvatarUrl string `json:"avatar_url"`
}

type GitLabMergeRequest struct {
	TargetBranch   string       `json:"target_branch"`
	Title          string       `json:"title"`
	LastCommit     GitLabCommit `json:"last_commit"`
	SourceBranch   string       `json:"source_branch"`
	ChangeNumber   int          `json:"iid"`
	Url            string       `json:"url"`
	MergeCommitSha string       `json:"merge_commit_sha"`
}

type GitLabCommit struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// GitLabCommentEvent represents GitLab comment event.
type GitLabCommentEvent struct {
	Project          GitLabProject      `json:"project"`
	MergeRequest     GitLabMergeRequest `json:"merge_request"`
	ObjectAttributes GitLabComment      `json:"object_attributes"`
	User             GitLabUser         `json:"user"`
}

type GitLabComment struct {
	Note string `json:"note"`
}

type BitbucketEvent struct {
	Repository  BitbucketRepository  `json:"repository"`
	PullRequest BitbucketPullRequest `json:"pullrequest"`
	Comment     BitbucketComment     `json:"comment"`
}

type BitbucketRepository struct {
	FullName string `json:"full_name"`
}

type BitbucketPullRequest struct {
	ID          int                      `json:"id"`
	Title       string                   `json:"title"`
	Source      BitbucketPullRequestSrc  `json:"source"`
	Destination BitbucketPullRequestDest `json:"destination"`
	LastCommit  BitbucketCommit          `json:"last_commit"`
	Author      BitbucketAuthor          `json:"author"`
	Links       struct {
		Html struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
}

type BitbucketAuthor struct {
	DisplayName string `json:"display_name"`
	Links       struct {
		Avatar struct {
			Href string `json:"href"`
		} `json:"avatar"`
	} `json:"links"`
}

type BitbucketPullRequestSrc struct {
	Branch BitbucketBranch `json:"branch"`
	Commit BitbucketCommit `json:"commit"`
}

type BitbucketPullRequestDest struct {
	Branch BitbucketBranch `json:"branch"`
	Commit BitbucketCommit `json:"commit"`
}

type BitbucketBranch struct {
	Name string `json:"name"`
}

type BitbucketComment struct {
	Content BitbucketCommentContent `json:"content"`
}

type BitbucketCommentContent struct {
	Raw string `json:"raw"`
}

type BitbucketCommit struct {
	Hash string `json:"hash"`
}

const (
	GitProviderGitHub    = "github"
	GitProviderGitLab    = "gitlab"
	GitProviderGerrit    = "gerrit"
	GitProviderBitbucket = "bitbucket"

	GerritEventTypeCommentAdded    = "comment-added"
	GitHubEventTypeCommentAdded    = "issue_comment"
	GitLabEventTypeCommentAdded    = "Note Hook"
	BitbucketEventTypeCommentAdded = "pullrequest:comment_created"

	EventTypeReviewComment = "comment"
	EventTypeMerge         = "merge"

	RecheckComment  = "/recheck"
	OkToTestComment = "/ok-to-test"
)

// EventInfo represents information about an event.
type EventInfo struct {
	GitProvider        string
	RepoPath           string
	TargetBranch       string
	Type               string
	Codebase           *codebaseApi.Codebase
	HasPipelineRecheck bool
	PullRequest        *PullRequest
}

type PullRequest struct {
	HeadRef           string `json:"headRef"`
	HeadSha           string `json:"headSha"`
	Title             string `json:"title"`
	ChangeNumber      int    `json:"changeNumber"`
	LastCommitMessage string `json:"lastCommitMessage"`
	Author            string `json:"author"`
	AuthorAvatarUrl   string `json:"authorAvatarUrl"`
	Url               string `json:"url"`
}

// IsReviewCommentEvent returns true if the event is a review comment event.
func (e *EventInfo) IsReviewCommentEvent() bool {
	return e.Type == EventTypeReviewComment
}

// ContainsPipelineRecheckPrefix checks if the comment starts with the pipeline recheck or ok to test comment.
func ContainsPipelineRecheckPrefix(s string) bool {
	return strings.HasPrefix(s, RecheckComment) || strings.HasPrefix(s, OkToTestComment)
}

// ContainsPipelineRecheck checks if the comment contains the pipeline recheck or ok to test comment.
// It's used for Gerrit because its comments are in the format: "Patch Set 2:\n\n/recheck".
func ContainsPipelineRecheck(s string) bool {
	return strings.Contains(s, RecheckComment) || strings.Contains(s, OkToTestComment)
}
