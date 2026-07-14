// Package gitlab publishes pipeline report comments as GitLab merge request notes.
package gitlab

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"

	"github.com/epam/edp-tekton/pkg/reporter/provider/types"
)

// Provider posts report comments via the GitLab merge request notes REST API.
type Provider struct {
	client *resty.Client
}

// New creates a GitLab provider for the given host (e.g. gitlab.com).
func New(host, token string) *Provider {
	client := resty.New().
		SetBaseURL(fmt.Sprintf("https://%s/api/v4", host)).
		SetAuthToken(token)

	return &Provider{client: client}
}

// NewWithClient creates a provider with a pre-configured resty client (used in tests).
func NewWithClient(client *resty.Client) *Provider {
	return &Provider{client: client}
}

type note struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

// UpsertComment creates the report note or, when comment.Update is set, edits
// the previous report note identified by the marker.
func (p *Provider) UpsertComment(ctx context.Context, ref types.PullRequestRef, comment types.Comment) error {
	project := url.PathEscape(ref.RepoFullName)

	if comment.Update {
		existingID, err := p.findNote(ctx, project, ref.Number, comment.Marker)
		if err != nil {
			return err
		}

		if existingID != 0 {
			resp, err := p.client.R().
				SetContext(ctx).
				SetBody(map[string]string{"body": comment.Body}).
				Put(fmt.Sprintf("/projects/%s/merge_requests/%d/notes/%d", project, ref.Number, existingID))
			if err != nil {
				return fmt.Errorf("failed to update GitLab note %d: %w", existingID, err)
			}

			if resp.IsError() {
				return fmt.Errorf("failed to update GitLab note %d: status %s", existingID, resp.Status())
			}

			return nil
		}
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(map[string]string{"body": comment.Body}).
		Post(fmt.Sprintf("/projects/%s/merge_requests/%d/notes", project, ref.Number))
	if err != nil {
		return fmt.Errorf("failed to create GitLab note: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("failed to create GitLab note: status %s", resp.Status())
	}

	return nil
}

func (p *Provider) findNote(ctx context.Context, project string, mergeRequestIID int, marker string) (int, error) {
	page := 1

	for {
		var notes []note

		resp, err := p.client.R().
			SetContext(ctx).
			ForceContentType("application/json").
			SetQueryParams(map[string]string{
				"per_page": "100",
				"page":     strconv.Itoa(page),
				// Most-recently-updated first: the report note is edited on
				// every run, so it is found on the first page.
				"order_by": "updated_at",
				"sort":     "desc",
			}).
			SetResult(&notes).
			Get(fmt.Sprintf("/projects/%s/merge_requests/%d/notes", project, mergeRequestIID))
		if err != nil {
			return 0, fmt.Errorf("failed to list GitLab notes: %w", err)
		}

		if resp.IsError() {
			return 0, fmt.Errorf("failed to list GitLab notes: status %s", resp.Status())
		}

		for _, n := range notes {
			if strings.Contains(n.Body, marker) {
				return n.ID, nil
			}
		}

		nextPage := resp.Header().Get("X-Next-Page")
		if nextPage == "" || len(notes) == 0 {
			return 0, nil
		}

		page++
	}
}
