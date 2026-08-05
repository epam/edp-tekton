// Package types holds the contracts shared by all git provider implementations.
package types

import "context"

type PullRequestRef struct {
	// RepoFullName is the full repository path, e.g. org/repo or group/subgroup/repo.
	RepoFullName string
	// Number is the pull request number (merge request IID for GitLab).
	Number int
}

// CommentStrategy selects how the report comment is published to the pull
// request thread. It is defined here rather than in the reporter config so
// provider implementations can branch on it without depending on
// configuration loading.
type CommentStrategy string

const (
	// CommentStrategyUpdate finds the previous report comment by its hidden
	// marker and edits it in place.
	CommentStrategyUpdate CommentStrategy = "update"

	// CommentStrategyNew always creates a new comment.
	CommentStrategyNew CommentStrategy = "new"

	// CommentStrategyRecreate creates a new comment and then deletes every
	// previous comment carrying the same marker, so exactly one report stays
	// visible at the bottom of the thread. Create-first ordering guarantees
	// the pull request is never left without a report when cleanup fails.
	CommentStrategyRecreate CommentStrategy = "recreate"
)

type Comment struct {
	// Marker is a hidden HTML comment identifying report comments, used to
	// find, update or delete a previously published report.
	Marker string
	// Body is the full, already-truncated markdown body (including the marker).
	Body string

	Strategy CommentStrategy
}

// CleanupError reports that the new comment was published but deleting stale
// report comments failed. Publishing succeeded, so callers must treat the
// report as delivered — retrying UpsertComment would duplicate it — and rely
// on the next run's recreate pass to sweep the leftovers.
type CleanupError struct {
	Err error
}

func (e *CleanupError) Error() string { return e.Err.Error() }

func (e *CleanupError) Unwrap() error { return e.Err }

// Provider posts, updates or recreates a pull request comment.
type Provider interface {
	UpsertComment(ctx context.Context, ref PullRequestRef, comment Comment) error
}

// CollapsibleSectionsSupport is an optional Provider capability (like
// CommitStatusSetter) reporting whether the provider's comment renderer
// executes embedded HTML (<details>/<summary>) instead of escaping it as
// literal text. Providers that do not implement it get failed-step logs
// rendered as plain markdown, which every renderer displays identically.
type CollapsibleSectionsSupport interface {
	SupportsCollapsibleSections() bool
}

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
