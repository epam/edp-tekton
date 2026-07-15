// Package gitlab publishes pipeline report comments as GitLab merge request notes.
package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"

	"github.com/epam/edp-tekton/pkg/reporter/provider/retry"
	"github.com/epam/edp-tekton/pkg/reporter/provider/types"
)

// GitLab commit status API states.
const (
	statePending  = "pending"
	stateRunning  = "running"
	stateSuccess  = "success"
	stateFailed   = "failed"
	stateCanceled = "canceled"
)

// Provider posts report comments via the GitLab merge request notes REST API.
type Provider struct {
	client *resty.Client
}

// New creates a GitLab provider for the given host (e.g. gitlab.com).
// Transient API failures are retried per the shared retry policy.
func New(host, token string) *Provider {
	client := retry.ConfigureResty(resty.New().
		SetBaseURL(fmt.Sprintf("https://%s/api/v4", host)).
		SetAuthToken(token))

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

// SetCommitStatus posts a commit status via the GitLab commit statuses API.
//
// GitLab enforces a state machine on commit statuses: an unfinished status
// (pending/running) in the same context is transitioned in place and only
// valid transitions are accepted, while a finished or absent status is
// recreated, accepting any state. The current status is therefore read first
// and the post is skipped when it would be an invalid transition (e.g. a
// stuck run left the check pending/running); the check already shows a live
// state for this SHA and the next terminal post resets it.
func (p *Provider) SetCommitStatus(ctx context.Context, ref types.CommitRef, status types.CommitStatus) error {
	state, err := apiState(status.State)
	if err != nil {
		return err
	}

	// Best-effort read: on lookup failure fall through to the post, where the
	// invalid-transition guard still covers the conflict.
	if current, err := p.currentStatus(ctx, ref, status.Context); err == nil && !validTransition(current, state) {
		return nil
	}

	body := map[string]string{
		"state":       state,
		"context":     status.Context,
		"description": status.Description,
	}
	if status.TargetURL != "" {
		body["target_url"] = status.TargetURL
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(body).
		Post(fmt.Sprintf("/projects/%s/statuses/%s", url.PathEscape(ref.RepoFullName), url.PathEscape(ref.Sha)))
	if err != nil {
		return fmt.Errorf("failed to set GitLab commit status: %w", err)
	}

	if resp.IsError() {
		// The status may change between the read above and this post (e.g. a
		// concurrent run transitions it); treat the resulting conflict as
		// already reported, same as the pre-check.
		if isInvalidTransition(resp) {
			return nil
		}

		return fmt.Errorf("failed to set GitLab commit status: status %s", resp.Status())
	}

	return nil
}

// commitStatus is the subset of the commit statuses API response we consume.
type commitStatus struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

// currentStatus returns the newest status for the commit in the given context
// (GitLab calls it name), or "" when the context has no status yet.
//
// The read is deliberately not scoped by ref: the subsequent post attaches to
// the newest pipeline for the SHA, which is the same scope the newest-first
// read observes. If the same SHA carries a same-context status on another ref,
// the worst case is a skipped best-effort post — the next terminal status
// resets the context.
func (p *Provider) currentStatus(ctx context.Context, ref types.CommitRef, statusContext string) (string, error) {
	var statuses []commitStatus

	resp, err := p.client.R().
		SetContext(ctx).
		ForceContentType("application/json").
		SetQueryParams(map[string]string{
			"name":     statusContext,
			"order_by": "id",
			"sort":     "desc",
		}).
		SetResult(&statuses).
		Get(fmt.Sprintf("/projects/%s/repository/commits/%s/statuses",
			url.PathEscape(ref.RepoFullName), url.PathEscape(ref.Sha)))
	if err != nil {
		return "", fmt.Errorf("failed to list GitLab commit statuses: %w", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("failed to list GitLab commit statuses: status %s", resp.Status())
	}

	if len(statuses) == 0 {
		return "", nil
	}

	// order_by/sort are not part of this endpoint's documented contract and
	// self-managed GitLab was observed to ignore them (returning oldest first),
	// so never trust the server ordering: pick the newest (highest id) status
	// ourselves.
	newest := statuses[0]
	for _, s := range statuses[1:] {
		if s.ID > newest.ID {
			newest = s
		}
	}

	return newest.Status, nil
}

// validTransition mirrors GitLab's CommitStatus state machine. Unfinished
// statuses are transitioned in place: pending accepts every event except
// another enqueue (pending), running accepts only terminal events. A finished
// or absent status is recreated, so any state is valid.
//
// The default branch is only correct for the states apiState can produce
// today (pending: enqueue is valid from created/skipped/manual/scheduled, and
// finished statuses are recreated). Re-validate this table against GitLab's
// commit_status.rb state machine before wiring any new state through apiState.
func validTransition(current, desired string) bool {
	switch current {
	case statePending:
		return desired != statePending
	case stateRunning:
		return desired == stateSuccess || desired == stateFailed || desired == stateCanceled
	default:
		return true
	}
}

func isInvalidTransition(resp *resty.Response) bool {
	return resp.StatusCode() == http.StatusBadRequest &&
		strings.Contains(resp.String(), "Cannot transition status")
}

func apiState(state types.CommitState) (string, error) {
	switch state {
	case types.CommitStatePending:
		return statePending, nil
	default:
		return "", fmt.Errorf("unsupported GitLab commit state %q", state)
	}
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
