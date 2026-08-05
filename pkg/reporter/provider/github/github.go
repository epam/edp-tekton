// Package github publishes pipeline report comments to GitHub pull requests.
package github

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v81/github"

	"github.com/epam/edp-tekton/pkg/reporter/provider/retry"
	"github.com/epam/edp-tekton/pkg/reporter/provider/types"
)

const publicHost = "github.com"

// Provider posts report comments via the GitHub Issues comments API
// (pull request conversation comments are issue comments).
type Provider struct {
	client *github.Client
}

// New creates a GitHub provider. For hosts other than github.com the
// GitHub Enterprise API prefix (/api/v3) is used.
func New(host, token string) (*Provider, error) {
	client := github.NewClient(nil).WithAuthToken(token)

	if host != "" && host != publicHost {
		baseURL := fmt.Sprintf("https://%s/api/v3/", host)

		enterprise, err := client.WithEnterpriseURLs(baseURL, baseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to configure GitHub Enterprise URL for host %q: %w", host, err)
		}

		client = enterprise
	}

	return &Provider{client: client}, nil
}

// NewWithClient creates a provider with a pre-configured client (used in tests).
func NewWithClient(client *github.Client) *Provider {
	return &Provider{client: client}
}

// SupportsCollapsibleSections reports that GitHub renders embedded HTML
// (<details>/<summary>) inside markdown comments.
func (p *Provider) SupportsCollapsibleSections() bool { return true }

// UpsertComment publishes the report comment per comment.Strategy: update
// edits the previous report comment identified by the marker, recreate posts
// a new comment and then deletes the stale ones (create-first, so a cleanup
// failure never leaves the pull request without a report), new always posts.
func (p *Provider) UpsertComment(ctx context.Context, ref types.PullRequestRef, comment types.Comment) error {
	owner, repo, err := splitRepo(ref.RepoFullName)
	if err != nil {
		return err
	}

	if comment.Strategy == types.CommentStrategyUpdate {
		existingIDs, err := p.findComments(ctx, owner, repo, ref.Number, comment.Marker, false)
		if err != nil {
			return err
		}

		if len(existingIDs) != 0 {
			_, _, err = p.client.Issues.EditComment(ctx, owner, repo, existingIDs[0], &github.IssueComment{
				Body: github.Ptr(comment.Body),
			})
			if err != nil {
				return fmt.Errorf("failed to update GitHub comment %d: %w", existingIDs[0], err)
			}

			return nil
		}
	}

	created, _, err := p.client.Issues.CreateComment(ctx, owner, repo, ref.Number, &github.IssueComment{
		Body: github.Ptr(comment.Body),
	})
	if err != nil {
		return fmt.Errorf("failed to create GitHub comment: %w", err)
	}

	if comment.Strategy == types.CommentStrategyRecreate {
		if err := p.deleteStaleComments(ctx, owner, repo, ref.Number, comment.Marker, created.GetID()); err != nil {
			return &types.CleanupError{Err: err}
		}
	}

	return nil
}

// deleteStaleComments removes every marker comment except the just-created
// one (keepID). Unlike the resty-based providers, GitHub comment calls carry
// no client-level retry; deletes are idempotent and cleanup failures are
// non-fatal, so leftovers are swept on the next recreate pass.
func (p *Provider) deleteStaleComments(
	ctx context.Context,
	owner, repo string,
	number int,
	marker string,
	keepID int64,
) error {
	ids, err := p.findComments(ctx, owner, repo, number, marker, true)
	if err != nil {
		return err
	}

	for _, id := range ids {
		if id == keepID {
			continue
		}

		if _, err := p.client.Issues.DeleteComment(ctx, owner, repo, id); err != nil {
			return fmt.Errorf("failed to delete GitHub comment %d: %w", id, err)
		}
	}

	return nil
}

func (p *Provider) SetCommitStatus(ctx context.Context, ref types.CommitRef, status types.CommitStatus) error {
	owner, repo, err := splitRepo(ref.RepoFullName)
	if err != nil {
		return err
	}

	state, err := apiState(status.State)
	if err != nil {
		return err
	}

	repoStatus := github.RepoStatus{
		State:       github.Ptr(state),
		Context:     github.Ptr(status.Context),
		Description: github.Ptr(status.Description),
	}
	if status.TargetURL != "" {
		repoStatus.TargetURL = github.Ptr(status.TargetURL)
	}

	// Unlike GitLab, GitHub statuses have no state machine: every POST creates
	// a new status and the latest one wins, so retrying is always safe.
	err = retry.Do(ctx, func() error {
		_, _, err := p.client.Repositories.CreateStatus(ctx, owner, repo, ref.Sha, repoStatus)

		return err
	}, transient)
	if err != nil {
		return fmt.Errorf("failed to set GitHub commit status: %w", err)
	}

	return nil
}

// transient classifies go-github errors for the retry policy: primary and
// secondary rate limits, 5xx responses and transport errors are retryable.
func transient(err error) bool {
	var (
		rateErr  *github.RateLimitError
		abuseErr *github.AbuseRateLimitError
	)

	if errors.As(err, &rateErr) || errors.As(err, &abuseErr) {
		return true
	}

	var respErr *github.ErrorResponse
	if errors.As(err, &respErr) && respErr.Response != nil {
		return retry.Transient(respErr.Response.StatusCode, err)
	}

	return retry.Transient(0, err)
}

func apiState(state types.CommitState) (string, error) {
	switch state {
	case types.CommitStatePending:
		return "pending", nil
	default:
		return "", fmt.Errorf("unsupported GitHub commit state %q", state)
	}
}

// findComments returns the IDs of comments carrying the marker, newest
// updated first. With all set it walks every page so the recreate strategy
// sweeps the full backlog (including comments accumulated under the new
// strategy); without it the scan stops at the first match — the report
// comment is touched on every run, so the update strategy finds it on the
// first page even in long comment threads.
func (p *Provider) findComments(
	ctx context.Context,
	owner, repo string,
	number int,
	marker string,
	all bool,
) ([]int64, error) {
	opts := &github.IssueListCommentsOptions{
		Sort:        github.Ptr("updated"),
		Direction:   github.Ptr("desc"),
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var ids []int64

	for {
		comments, resp, err := p.client.Issues.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list GitHub comments: %w", err)
		}

		for _, c := range comments {
			if strings.Contains(c.GetBody(), marker) {
				ids = append(ids, c.GetID())

				if !all {
					return ids, nil
				}
			}
		}

		if resp.NextPage == 0 {
			return ids, nil
		}

		opts.Page = resp.NextPage
	}
}

func splitRepo(fullName string) (owner, repo string, err error) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid GitHub repository name %q, expected owner/repo", fullName)
	}

	return parts[0], parts[1], nil
}
