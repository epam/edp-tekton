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

	return NewWithClient(resty.New().SetBaseURL(server.URL))
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
