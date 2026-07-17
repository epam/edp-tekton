package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v81/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/epam/edp-tekton/pkg/reporter/provider/types"
)

func newTestProvider(t *testing.T, handler http.Handler) *Provider {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := github.NewClient(server.Client())

	baseURL, err := url.Parse(server.URL + "/")
	require.NoError(t, err)

	client.BaseURL = baseURL

	return NewWithClient(client)
}

func TestUpsertCommentCreatesWhenNoMarkerFound(t *testing.T) {
	t.Parallel()

	var created string

	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/org/repo/issues/7/comments", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode([]*github.IssueComment{
			{ID: github.Ptr(int64(1)), Body: github.Ptr("unrelated comment")},
		})
	})
	mux.HandleFunc("POST /repos/org/repo/issues/7/comments", func(w http.ResponseWriter, r *http.Request) {
		var c github.IssueComment
		require.NoError(t, json.NewDecoder(r.Body).Decode(&c))
		created = c.GetBody()

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(&c)
	})

	p := newTestProvider(t, mux)

	err := p.UpsertComment(context.Background(), types.PullRequestRef{RepoFullName: "org/repo", Number: 7},
		types.Comment{Marker: "<!-- m -->", Body: "<!-- m -->\nreport", Update: true})
	require.NoError(t, err)
	assert.Equal(t, "<!-- m -->\nreport", created)
}

func TestUpsertCommentUpdatesExisting(t *testing.T) {
	t.Parallel()

	var updated string

	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/org/repo/issues/7/comments", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode([]*github.IssueComment{
			{ID: github.Ptr(int64(42)), Body: github.Ptr("<!-- m -->\nold report")},
		})
	})
	mux.HandleFunc("PATCH /repos/org/repo/issues/comments/42", func(w http.ResponseWriter, r *http.Request) {
		var c github.IssueComment
		require.NoError(t, json.NewDecoder(r.Body).Decode(&c))
		updated = c.GetBody()
		_ = json.NewEncoder(w).Encode(&c)
	})
	mux.HandleFunc("POST /repos/org/repo/issues/7/comments", func(w http.ResponseWriter, _ *http.Request) {
		t.Error("must not create a new comment when one with the marker exists")
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprint(w, "{}")
	})

	p := newTestProvider(t, mux)

	err := p.UpsertComment(context.Background(), types.PullRequestRef{RepoFullName: "org/repo", Number: 7},
		types.Comment{Marker: "<!-- m -->", Body: "<!-- m -->\nnew report", Update: true})
	require.NoError(t, err)
	assert.Equal(t, "<!-- m -->\nnew report", updated)
}

func TestUpsertCommentAlwaysCreatesWithoutUpdate(t *testing.T) {
	t.Parallel()

	createCalls := 0

	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/org/repo/issues/7/comments", func(w http.ResponseWriter, _ *http.Request) {
		t.Error("must not list comments when update strategy is off")

		_, _ = fmt.Fprint(w, "[]")
	})
	mux.HandleFunc("POST /repos/org/repo/issues/7/comments", func(w http.ResponseWriter, _ *http.Request) {
		createCalls++

		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprint(w, "{}")
	})

	p := newTestProvider(t, mux)

	err := p.UpsertComment(context.Background(), types.PullRequestRef{RepoFullName: "org/repo", Number: 7},
		types.Comment{Marker: "<!-- m -->", Body: "report", Update: false})
	require.NoError(t, err)
	assert.Equal(t, 1, createCalls)
}

func TestSplitRepoInvalid(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t, http.NewServeMux())

	err := p.UpsertComment(context.Background(), types.PullRequestRef{RepoFullName: "invalid", Number: 1},
		types.Comment{Body: "b"})
	assert.Error(t, err)
}

func TestSetCommitStatusPostsPending(t *testing.T) {
	t.Parallel()

	var posted *github.RepoStatus

	mux := http.NewServeMux()
	mux.HandleFunc("POST /repos/org/repo/statuses/abc123", func(w http.ResponseWriter, r *http.Request) {
		var s github.RepoStatus
		require.NoError(t, json.NewDecoder(r.Body).Decode(&s))
		posted = &s

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(&s)
	})

	p := newTestProvider(t, mux)

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "org/repo", Sha: "abc123"},
		types.CommitStatus{
			State:       types.CommitStatePending,
			Context:     "Review Pipeline",
			Description: "QUEUED",
			TargetURL:   "https://example.com/pr/1",
		})
	require.NoError(t, err)
	require.NotNil(t, posted)
	assert.Equal(t, "pending", posted.GetState())
	assert.Equal(t, "Review Pipeline", posted.GetContext())
	assert.Equal(t, "QUEUED", posted.GetDescription())
	assert.Equal(t, "https://example.com/pr/1", posted.GetTargetURL())
}

func TestSetCommitStatusPropagatesAPIError(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /repos/org/repo/statuses/abc123", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	p := newTestProvider(t, mux)

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "org/repo", Sha: "abc123"},
		types.CommitStatus{State: types.CommitStatePending})
	assert.Error(t, err)
}

func TestSetCommitStatusRetriesTransientError(t *testing.T) {
	t.Parallel()

	calls := 0

	mux := http.NewServeMux()
	mux.HandleFunc("POST /repos/org/repo/statuses/abc123", func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusBadGateway)

			return
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	})

	p := newTestProvider(t, mux)

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "org/repo", Sha: "abc123"},
		types.CommitStatus{State: types.CommitStatePending})
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
}

func TestSetCommitStatusDoesNotRetryPermanentError(t *testing.T) {
	t.Parallel()

	calls := 0

	mux := http.NewServeMux()
	mux.HandleFunc("POST /repos/org/repo/statuses/abc123", func(w http.ResponseWriter, _ *http.Request) {
		calls++

		w.WriteHeader(http.StatusUnprocessableEntity)
	})

	p := newTestProvider(t, mux)

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "org/repo", Sha: "abc123"},
		types.CommitStatus{State: types.CommitStatePending})
	assert.Error(t, err)
	assert.Equal(t, 1, calls)
}

func TestSetCommitStatusRejectsUnsupportedState(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t, http.NewServeMux())

	err := p.SetCommitStatus(context.Background(),
		types.CommitRef{RepoFullName: "org/repo", Sha: "abc123"},
		types.CommitStatus{State: "success"})
	assert.ErrorContains(t, err, "unsupported GitHub commit state")
}

func TestSupportsCollapsibleSections(t *testing.T) {
	t.Parallel()

	p, err := New("github.com", "token")
	require.NoError(t, err)

	assert.True(t, p.SupportsCollapsibleSections(), "GitHub renders embedded HTML in markdown comments")
}
