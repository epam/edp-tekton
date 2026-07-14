package bitbucket

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

func newTestProvider(t *testing.T, handler http.Handler) *Provider {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return NewWithClient(resty.New().SetBaseURL(server.URL), "dGVzdA==")
}

func newComment(id int, raw string) comment {
	c := comment{ID: id}
	c.Content.Raw = raw

	return c
}

func TestUpsertCommentUpdatesExisting(t *testing.T) {
	t.Parallel()

	var updated string

	mux := http.NewServeMux()
	mux.HandleFunc("GET /repositories/ws/repo/pullrequests/3/comments", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Basic dGVzdA==", r.Header.Get("Authorization"))

		_ = json.NewEncoder(w).Encode(commentsPage{Values: []comment{
			newComment(1, "unrelated"),
			newComment(9, "<!-- m -->\nold"),
		}})
	})
	mux.HandleFunc("PUT /repositories/ws/repo/pullrequests/3/comments/9", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Content struct {
				Raw string `json:"raw"`
			} `json:"content"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		updated = body.Content.Raw
		_ = json.NewEncoder(w).Encode(newComment(9, updated))
	})

	p := newTestProvider(t, mux)

	err := p.UpsertComment(context.Background(), types.PullRequestRef{RepoFullName: "ws/repo", Number: 3},
		types.Comment{Marker: "<!-- m -->", Body: "<!-- m -->\nnew", Update: true})
	require.NoError(t, err)
	assert.Equal(t, "<!-- m -->\nnew", updated)
}

func TestUpsertCommentFollowsPaginationThenCreates(t *testing.T) {
	t.Parallel()

	created := false

	var server *httptest.Server

	mux := http.NewServeMux()
	mux.HandleFunc("GET /repositories/ws/repo/pullrequests/3/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			_ = json.NewEncoder(w).Encode(commentsPage{Values: []comment{newComment(2, "also unrelated")}})

			return
		}

		_ = json.NewEncoder(w).Encode(commentsPage{
			Values: []comment{newComment(1, "unrelated")},
			Next:   server.URL + "/repositories/ws/repo/pullrequests/3/comments?page=2",
		})
	})
	mux.HandleFunc("POST /repositories/ws/repo/pullrequests/3/comments", func(w http.ResponseWriter, _ *http.Request) {
		created = true

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(newComment(5, "new"))
	})

	server = httptest.NewServer(mux)
	t.Cleanup(server.Close)

	p := NewWithClient(resty.New().SetBaseURL(server.URL), "token")

	err := p.UpsertComment(context.Background(), types.PullRequestRef{RepoFullName: "ws/repo", Number: 3},
		types.Comment{Marker: "<!-- m -->", Body: "<!-- m -->\nreport", Update: true})
	require.NoError(t, err)
	assert.True(t, created)
}

func TestUpsertCommentPropagatesAPIError(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	p := newTestProvider(t, mux)

	err := p.UpsertComment(context.Background(), types.PullRequestRef{RepoFullName: "ws/repo", Number: 3},
		types.Comment{Marker: "<!-- m -->", Body: "b", Update: false})
	assert.Error(t, err)
}
