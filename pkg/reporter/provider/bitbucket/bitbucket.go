// Package bitbucket publishes pipeline report comments to Bitbucket Cloud pull requests.
package bitbucket

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"

	"github.com/epam/edp-tekton/pkg/reporter/provider/retry"
	"github.com/epam/edp-tekton/pkg/reporter/provider/types"
)

const cloudAPIBaseURL = "https://api.bitbucket.org/2.0"

type Provider struct {
	client *resty.Client
	token  string
}

// New creates a Bitbucket Cloud provider. The token is the base64-encoded
// credentials stored in the GitServer secret, sent as Basic authorization
// (same convention as the interceptor's Bitbucket integration).
// Transient API failures are retried per the shared retry policy; build
// statuses are keyed upserts, so retrying is always safe.
func New(token string) *Provider {
	return &Provider{
		client: retry.ConfigureResty(resty.New().SetBaseURL(cloudAPIBaseURL)),
		token:  token,
	}
}

// NewWithClient creates a provider with a pre-configured resty client (used in tests).
func NewWithClient(client *resty.Client, token string) *Provider {
	return &Provider{client: client, token: token}
}

// SupportsCollapsibleSections is false because Bitbucket Cloud escapes
// embedded HTML, rendering <details>/<summary> as literal tags.
func (p *Provider) SupportsCollapsibleSections() bool { return false }

type comment struct {
	ID      int `json:"id"`
	Content struct {
		Raw string `json:"raw"`
	} `json:"content"`
}

type commentsPage struct {
	Values []comment `json:"values"`
	Next   string    `json:"next"`
}

// UpsertComment publishes the report comment per c.Strategy: update edits the
// previous report comment identified by the marker, recreate posts a new
// comment and then deletes the stale ones (create-first, so a cleanup failure
// never leaves the pull request without a report), new always posts.
func (p *Provider) UpsertComment(ctx context.Context, ref types.PullRequestRef, c types.Comment) error {
	body := map[string]any{"content": map[string]string{"raw": c.Body}}

	if c.Strategy == types.CommentStrategyUpdate {
		existingIDs, err := p.findComments(ctx, ref, c.Marker, false)
		if err != nil {
			return err
		}

		if len(existingIDs) != 0 {
			resp, err := p.request(ctx).
				SetBody(body).
				Put(fmt.Sprintf("/repositories/%s/pullrequests/%d/comments/%d", ref.RepoFullName, ref.Number, existingIDs[0]))
			if err != nil {
				return fmt.Errorf("failed to update Bitbucket comment %d: %w", existingIDs[0], err)
			}

			if resp.IsError() {
				return fmt.Errorf("failed to update Bitbucket comment %d: status %s", existingIDs[0], resp.Status())
			}

			return nil
		}
	}

	var created comment

	resp, err := p.request(ctx).
		SetBody(body).
		SetResult(&created).
		Post(fmt.Sprintf("/repositories/%s/pullrequests/%d/comments", ref.RepoFullName, ref.Number))
	if err != nil {
		return fmt.Errorf("failed to create Bitbucket comment: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("failed to create Bitbucket comment: status %s", resp.Status())
	}

	if c.Strategy == types.CommentStrategyRecreate {
		if err := p.deleteStaleComments(ctx, ref, c.Marker, created.ID); err != nil {
			return &types.CleanupError{Err: err}
		}
	}

	return nil
}

// deleteStaleComments removes every marker comment except the just-created
// one (keepID). Deletes are idempotent and cleanup failures are non-fatal, so
// leftovers are swept on the next recreate pass.
func (p *Provider) deleteStaleComments(ctx context.Context, ref types.PullRequestRef, marker string, keepID int) error {
	ids, err := p.findComments(ctx, ref, marker, true)
	if err != nil {
		return err
	}

	for _, id := range ids {
		if id == keepID {
			continue
		}

		resp, err := p.request(ctx).
			Delete(fmt.Sprintf("/repositories/%s/pullrequests/%d/comments/%d", ref.RepoFullName, ref.Number, id))
		if err != nil {
			return fmt.Errorf("failed to delete Bitbucket comment %d: %w", id, err)
		}

		if resp.IsError() {
			return fmt.Errorf("failed to delete Bitbucket comment %d: status %s", id, resp.Status())
		}
	}

	return nil
}

func (p *Provider) SetCommitStatus(ctx context.Context, ref types.CommitRef, status types.CommitStatus) error {
	state, err := apiState(status.State)
	if err != nil {
		return err
	}

	body := map[string]string{
		"state":       state,
		"key":         status.Key,
		"name":        status.Name,
		"description": status.Description,
		// The url field is required by the Bitbucket build-status API.
		"url": status.TargetURL,
	}

	resp, err := p.request(ctx).
		SetBody(body).
		Post(fmt.Sprintf("/repositories/%s/commit/%s/statuses/build",
			escapePathSegments(ref.RepoFullName), url.PathEscape(ref.Sha)))
	if err != nil {
		return fmt.Errorf("failed to set Bitbucket build status: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("failed to set Bitbucket build status: status %s", resp.Status())
	}

	return nil
}

// escapePathSegments escapes every segment of a workspace/repo path while
// keeping the segment separators, which the Bitbucket API expects literal.
func escapePathSegments(path string) string {
	segments := strings.Split(path, "/")
	for i := range segments {
		segments[i] = url.PathEscape(segments[i])
	}

	return strings.Join(segments, "/")
}

func apiState(state types.CommitState) (string, error) {
	switch state {
	case types.CommitStatePending:
		// Bitbucket Cloud has no dedicated pending state; INPROGRESS is the
		// pending-equivalent the pipeline's own status task uses as well.
		return "INPROGRESS", nil
	default:
		return "", fmt.Errorf("unsupported Bitbucket commit state %q", state)
	}
}

// findComments returns the IDs of comments carrying the marker, newest
// updated first. With all set it walks every page so the recreate strategy
// sweeps the full backlog (including comments accumulated under the new
// strategy); without it the scan stops at the first match — the report
// comment is touched on every run, so the update strategy finds it on the
// first page even in long comment threads.
func (p *Provider) findComments(ctx context.Context, ref types.PullRequestRef, marker string, all bool) ([]int, error) {
	path := fmt.Sprintf("/repositories/%s/pullrequests/%d/comments?pagelen=100&sort=-updated_on",
		ref.RepoFullName, ref.Number)

	var ids []int

	for path != "" {
		var page commentsPage

		resp, err := p.request(ctx).
			SetResult(&page).
			Get(path)
		if err != nil {
			return nil, fmt.Errorf("failed to list Bitbucket comments: %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("failed to list Bitbucket comments: status %s", resp.Status())
		}

		for _, c := range page.Values {
			if strings.Contains(c.Content.Raw, marker) {
				ids = append(ids, c.ID)

				if !all {
					return ids, nil
				}
			}
		}

		// The API returns an absolute URL for the next page; strip the base so
		// the shared client (with its test-overridable base URL) can follow it.
		// resty also accepts absolute URLs, so a non-matching prefix still works.
		path = strings.TrimPrefix(page.Next, p.client.BaseURL)
	}

	return ids, nil
}

func (p *Provider) request(ctx context.Context) *resty.Request {
	return p.client.R().
		SetContext(ctx).
		ForceContentType("application/json").
		SetHeader("Authorization", fmt.Sprintf("Basic %s", p.token))
}
