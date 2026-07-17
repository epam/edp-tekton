// Package types holds the contracts shared by all git provider implementations.
package types

import "context"

// PullRequestRef identifies a pull/merge request.
type PullRequestRef struct {
	// RepoFullName is the full repository path, e.g. org/repo or group/subgroup/repo.
	RepoFullName string
	// Number is the pull request number (merge request IID for GitLab).
	Number int
}

// Comment is a report comment to publish.
type Comment struct {
	// Marker is a hidden HTML comment identifying report comments, used to
	// find and update a previously published report.
	Marker string
	// Body is the full, already-truncated markdown body (including the marker).
	Body string
	// Update requests editing an existing comment with the same Marker instead
	// of always creating a new one.
	Update bool
}

// Provider posts or updates a pull request comment.
type Provider interface {
	UpsertComment(ctx context.Context, ref PullRequestRef, comment Comment) error
	// SupportsCollapsibleSections reports whether the provider's comment
	// renderer executes embedded HTML (<details>/<summary>) instead of
	// escaping it as literal text. The reporter uses this to decide whether
	// failed-step logs can be rendered as a collapsible section or must fall
	// back to plain markdown.
	SupportsCollapsibleSections() bool
}

// CommitRef identifies a commit.
type CommitRef struct {
	// RepoFullName is the full repository path, e.g. org/repo or group/subgroup/repo.
	RepoFullName string
	Sha          string
}

// CommitState is a provider-agnostic commit status state; each provider maps
// it to its own API value.
type CommitState string

// CommitStatePending marks a commit as awaiting a CI verdict
// (GitLab/GitHub: pending, Bitbucket: INPROGRESS).
const CommitStatePending CommitState = "pending"

// CommitStatus is a provider-agnostic commit status request. Providers pick
// the labeling fields their API supports.
type CommitStatus struct {
	State CommitState
	// Context labels the check on GitLab/GitHub. It must match the context the
	// pipeline's own status tasks use so every stage updates the same check.
	Context string
	// Key is the Bitbucket build-status key: statuses with the same key
	// overwrite each other, so it must match the pipeline's status task KEY.
	Key string
	// Name is the Bitbucket build-status display name.
	Name        string
	Description string
	// TargetURL links the status to details; optional on GitLab/GitHub,
	// required by the Bitbucket build-status API.
	TargetURL string
}

// CommitStatusSetter sets a commit status (build status on Bitbucket).
type CommitStatusSetter interface {
	SetCommitStatus(ctx context.Context, ref CommitRef, status CommitStatus) error
}
