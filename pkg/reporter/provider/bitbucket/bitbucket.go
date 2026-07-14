// Package bitbucket publishes pipeline report comments to Bitbucket Cloud pull requests.
package bitbucket

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"

	"github.com/epam/edp-tekton/pkg/reporter/provider/types"
)

const cloudAPIBaseURL = "https://api.bitbucket.org/2.0"

// Provider posts report comments via the Bitbucket Cloud pull request comments API.
type Provider struct {
	client *resty.Client
	token  string
}

// New creates a Bitbucket Cloud provider. The token is the base64-encoded
// credentials stored in the GitServer secret, sent as Basic authorization
// (same convention as the interceptor's Bitbucket integration).
func New(token string) *Provider {
	return &Provider{
		client: resty.New().SetBaseURL(cloudAPIBaseURL),
		token:  token,
	}
}

// NewWithClient creates a provider with a pre-configured resty client (used in tests).
func NewWithClient(client *resty.Client, token string) *Provider {
	return &Provider{client: client, token: token}
}

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

// UpsertComment creates the report comment or, when c.Update is set, edits the
// previous report comment identified by the marker.
func (p *Provider) UpsertComment(ctx context.Context, ref types.PullRequestRef, c types.Comment) error {
	body := map[string]any{"content": map[string]string{"raw": c.Body}}

	if c.Update {
		existingID, err := p.findComment(ctx, ref, c.Marker)
		if err != nil {
			return err
		}

		if existingID != 0 {
			resp, err := p.request(ctx).
				SetBody(body).
				Put(fmt.Sprintf("/repositories/%s/pullrequests/%d/comments/%d", ref.RepoFullName, ref.Number, existingID))
			if err != nil {
				return fmt.Errorf("failed to update Bitbucket comment %d: %w", existingID, err)
			}

			if resp.IsError() {
				return fmt.Errorf("failed to update Bitbucket comment %d: status %s", existingID, resp.Status())
			}

			return nil
		}
	}

	resp, err := p.request(ctx).
		SetBody(body).
		Post(fmt.Sprintf("/repositories/%s/pullrequests/%d/comments", ref.RepoFullName, ref.Number))
	if err != nil {
		return fmt.Errorf("failed to create Bitbucket comment: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("failed to create Bitbucket comment: status %s", resp.Status())
	}

	return nil
}

func (p *Provider) findComment(ctx context.Context, ref types.PullRequestRef, marker string) (int, error) {
	// Most-recently-updated first: the report comment is edited on every run,
	// so it is found on the first page even in long comment threads.
	path := fmt.Sprintf("/repositories/%s/pullrequests/%d/comments?pagelen=100&sort=-updated_on",
		ref.RepoFullName, ref.Number)

	for path != "" {
		var page commentsPage

		resp, err := p.request(ctx).
			SetResult(&page).
			Get(path)
		if err != nil {
			return 0, fmt.Errorf("failed to list Bitbucket comments: %w", err)
		}

		if resp.IsError() {
			return 0, fmt.Errorf("failed to list Bitbucket comments: status %s", resp.Status())
		}

		for _, c := range page.Values {
			if strings.Contains(c.Content.Raw, marker) {
				return c.ID, nil
			}
		}

		// The API returns an absolute URL for the next page; strip the base so
		// the shared client (with its test-overridable base URL) can follow it.
		// resty also accepts absolute URLs, so a non-matching prefix still works.
		path = strings.TrimPrefix(page.Next, p.client.BaseURL)
	}

	return 0, nil
}

func (p *Provider) request(ctx context.Context) *resty.Request {
	return p.client.R().
		SetContext(ctx).
		ForceContentType("application/json").
		SetHeader("Authorization", fmt.Sprintf("Basic %s", p.token))
}
