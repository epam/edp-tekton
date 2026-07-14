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
}
