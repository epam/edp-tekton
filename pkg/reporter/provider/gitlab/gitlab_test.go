package gitlab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/epam/edp-tekton/pkg/reporter/provider/retry"
	"github.com/epam/edp-tekton/pkg/reporter/provider/types"
)

// route matches on method plus the raw (escaped) path, because GitLab project
// paths are URL-encoded (group%2Frepo) and must stay encoded on the wire.
type route struct {
	method string
	path   string
}

func newTestProvider(t *testing.T, routes map[route]http.HandlerFunc) *Provider {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler, ok := routes[route{method: r.Method, path: r.URL.EscapedPath()}]
		if !ok {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.EscapedPath())
			w.WriteHeader(http.StatusNotFound)

			return
		}

		handler(w, r)
	}))
	t.Cleanup(server.Close)

	// Configure retries exactly as New does, so tests exercise the production
	// retry policy (TestSetCommitStatusRetriesTransientError depends on it).
	return NewWithClient(retry.ConfigureResty(resty.New().SetBaseURL(server.URL)))
}

func TestUpsertCommentUpdatesExistingNote(t *testing.T) {
	t.Parallel()

	var updated string

	p := newTestProvider(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/projects/group%2Frepo/merge_requests/5/notes"}: func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode([]note{
				{ID: 1, Body: "unrelated"},
				{ID: 33, Body: "<!-- m -->\nold"},
			})
		},
		{http.MethodPut, "/projects/group%2Frepo/merge_requests/5/notes/33"}: func(w http.ResponseWriter, r *http.Request) {
			var body map[string]string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			updated = body["body"]
			_ = json.NewEncoder(w).Encode(note{ID: 33, Body: updated})
		},
	})

	err := p.UpsertComment(context.Background(), types.PullRequestRef{RepoFullName: "group/repo", Number: 5},
		types.Comment{Marker: "<!-- m -->", Body: "<!-- m -->\nnew", Update: true})
	require.NoError(t, err)
	assert.Equal(t, "<!-- m -->\nnew", updated)
}

func TestUpsertCommentCreatesNote(t *testing.T) {
	t.Parallel()

	var created string

	p := newTestProvider(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/projects/group%2Frepo/merge_requests/5/notes"}: func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode([]note{})
		},
		{http.MethodPost, "/projects/group%2Frepo/merge_requests/5/notes"}: func(w http.ResponseWriter, r *http.Request) {
			var body map[string]string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			created = body["body"]

			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(note{ID: 1, Body: created})
		},
	})

	err := p.UpsertComment(context.Background(), types.PullRequestRef{RepoFullName: "group/repo", Number: 5},
		types.Comment{Marker: "<!-- m -->", Body: "<!-- m -->\nreport", Update: true})
	require.NoError(t, err)
	assert.Equal(t, "<!-- m -->\nreport", created)
}

// statusesRoute is the commit statuses read the state-machine pre-check performs.
var statusesRoute = route{http.MethodGet, "/projects/group%2Frepo/repository/commits/abc123/statuses"}

// currentStatuses responds to the pre-check read with the given states,
// newest first (descending ids), matching the requested server ordering.
func currentStatuses(states ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		statuses := make([]commitStatus, 0, len(states))
		for i, s := range states {
			statuses = append(statuses, commitStatus{ID: int64(len(states) - i), Status: s})
		}

		_ = json.NewEncoder(w).Encode(statuses)
	}
}

func TestSetCommitStatusPostsPending(t *testing.T) {
	t.Parallel()

	var posted map[string]string

	p := newTestProvider(t, map[route]http.HandlerFunc{
		statusesRoute: currentStatuses(),
		{http.MethodPost, "/projects/group%2Frepo/statuses/abc123"}: func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, json.NewDecoder(r.Body).Decode(&posted))
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		},
	})

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "group/repo", Sha: "abc123"},
		types.CommitStatus{
			State:       types.CommitStatePending,
			Context:     "Review Pipeline",
			Description: "QUEUED",
			TargetURL:   "https://example.com/mr/1",
		})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"state":       "pending",
		"context":     "Review Pipeline",
		"description": "QUEUED",
		"target_url":  "https://example.com/mr/1",
	}, posted)
}

func TestSetCommitStatusIgnoresServerOrdering(t *testing.T) {
	t.Parallel()

	// Oldest first, as if the server ignored order_by/sort: the newest (highest
	// id) status is running, so posting pending must be skipped. Trusting the
	// server ordering would read the stale success and issue the POST (the
	// routes map fails the test on one).
	p := newTestProvider(t, map[route]http.HandlerFunc{
		statusesRoute: func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode([]commitStatus{
				{ID: 1, Status: "success"},
				{ID: 2, Status: "running"},
			})
		},
	})

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "group/repo", Sha: "abc123"},
		types.CommitStatus{State: types.CommitStatePending})
	require.NoError(t, err)
}

func TestSetCommitStatusPostsWhenStatusReadFails(t *testing.T) {
	t.Parallel()

	// The pre-check read is best-effort: a failing statuses lookup must fall
	// through to the POST, where the invalid-transition guard still applies.
	calls := 0

	p := newTestProvider(t, map[route]http.HandlerFunc{
		statusesRoute: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
		{http.MethodPost, "/projects/group%2Frepo/statuses/abc123"}: func(w http.ResponseWriter, _ *http.Request) {
			calls++

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		},
	})

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "group/repo", Sha: "abc123"},
		types.CommitStatus{State: types.CommitStatePending})
	require.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestSetCommitStatusOmitsEmptyTargetURL(t *testing.T) {
	t.Parallel()

	var posted map[string]string

	p := newTestProvider(t, map[route]http.HandlerFunc{
		statusesRoute: currentStatuses(),
		{http.MethodPost, "/projects/group%2Frepo/statuses/abc123"}: func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, json.NewDecoder(r.Body).Decode(&posted))
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		},
	})

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "group/repo", Sha: "abc123"},
		types.CommitStatus{State: types.CommitStatePending, Context: "Review Pipeline", Description: "QUEUED"})
	require.NoError(t, err)
	assert.NotContains(t, posted, "target_url")
}

func TestSetCommitStatusPropagatesAPIError(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t, map[route]http.HandlerFunc{
		statusesRoute: currentStatuses(),
		{http.MethodPost, "/projects/group%2Frepo/statuses/abc123"}: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		},
	})

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "group/repo", Sha: "abc123"},
		types.CommitStatus{State: types.CommitStatePending})
	assert.Error(t, err)
}

func TestSetCommitStatusSkipsInvalidTransition(t *testing.T) {
	t.Parallel()

	// A stuck run left the check running; pending is not a valid transition,
	// so no POST must be issued (the routes map fails the test on one).
	p := newTestProvider(t, map[route]http.HandlerFunc{
		statusesRoute: currentStatuses("running"),
	})

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "group/repo", Sha: "abc123"},
		types.CommitStatus{State: types.CommitStatePending})
	require.NoError(t, err)
}

func TestSetCommitStatusRecreatesFinishedStatus(t *testing.T) {
	t.Parallel()

	posted := false

	// A finished status is recreated by GitLab, so the post must proceed.
	p := newTestProvider(t, map[route]http.HandlerFunc{
		statusesRoute: currentStatuses("failed"),
		{http.MethodPost, "/projects/group%2Frepo/statuses/abc123"}: func(w http.ResponseWriter, _ *http.Request) {
			posted = true

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		},
	})

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "group/repo", Sha: "abc123"},
		types.CommitStatus{State: types.CommitStatePending})
	require.NoError(t, err)
	assert.True(t, posted)
}

func TestSetCommitStatusTreatsInvalidTransitionAsReported(t *testing.T) {
	t.Parallel()

	calls := 0

	// The pre-check sees no conflict, but the status changes before the post
	// lands (concurrent run): the race guard must swallow the conflict.
	p := newTestProvider(t, map[route]http.HandlerFunc{
		statusesRoute: currentStatuses(),
		{http.MethodPost, "/projects/group%2Frepo/statuses/abc123"}: func(w http.ResponseWriter, _ *http.Request) {
			calls++

			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message": "Cannot transition status via :enqueue from :pending"}`))
		},
	})

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "group/repo", Sha: "abc123"},
		types.CommitStatus{State: types.CommitStatePending})
	require.NoError(t, err)
	// A transition conflict is permanent: it must not be retried either.
	assert.Equal(t, 1, calls)
}

func TestSetCommitStatusRetriesTransientError(t *testing.T) {
	t.Parallel()

	calls := 0

	p := newTestProvider(t, map[route]http.HandlerFunc{
		statusesRoute: currentStatuses(),
		{http.MethodPost, "/projects/group%2Frepo/statuses/abc123"}: func(w http.ResponseWriter, _ *http.Request) {
			calls++
			if calls == 1 {
				w.WriteHeader(http.StatusBadGateway)

				return
			}

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		},
	})

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "group/repo", Sha: "abc123"},
		types.CommitStatus{State: types.CommitStatePending})
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
}

func TestValidTransition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		current string
		desired string
		want    bool
	}{
		{"no status yet accepts anything", "", "pending", true},
		{"pending rejects another pending", "pending", "pending", false},
		{"pending accepts running", "pending", "running", true},
		{"pending accepts success", "pending", "success", true},
		{"running rejects pending", "running", "pending", false},
		{"running rejects running", "running", "running", false},
		{"running accepts success", "running", "success", true},
		{"running accepts failed", "running", "failed", true},
		{"running accepts canceled", "running", "canceled", true},
		{"finished accepts pending", "failed", "pending", true},
		{"finished accepts running", "success", "running", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, validTransition(tt.current, tt.desired))
		})
	}
}

func TestSetCommitStatusRejectsUnsupportedState(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t, map[route]http.HandlerFunc{})

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "group/repo", Sha: "abc123"},
		types.CommitStatus{State: "success"})
	assert.ErrorContains(t, err, "unsupported GitLab commit state")
}

func TestUpsertCommentPropagatesAPIError(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/projects/group%2Frepo/merge_requests/5/notes"}: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		},
	})

	err := p.UpsertComment(context.Background(), types.PullRequestRef{RepoFullName: "group/repo", Number: 5},
		types.Comment{Marker: "<!-- m -->", Body: "b", Update: false})
	assert.Error(t, err)
}

func TestSupportsCollapsibleSections(t *testing.T) {
	t.Parallel()

	p := New("git.example.com", "token")

	assert.True(t, p.SupportsCollapsibleSections(), "GitLab renders embedded HTML in markdown notes")
}
