package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v31/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/pkg/apis/edp/v1"

	"github.com/epam/edp-tekton/pkg/event_processor"
)

func TestGitHubEventProcessor_processCommentEvent(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	type args struct {
		body any
	}

	tests := []struct {
		name        string
		args        args
		mockhttp    func(t *testing.T) (URL string, teardown func())
		kubeObjects []client.Object
		want        *event_processor.EventInfo
		wantErr     require.ErrorAssertionFunc
	}{
		{
			name: "comment event - should process pull request",
			args: args{
				body: map[string]interface{}{
					"action": "created",
					"issue": map[string]interface{}{
						"number": 1,
					},
					"comment": map[string]interface{}{
						"body": "/recheck",
					},
					"repository": map[string]interface{}{
						"full_name": "o/r",
						"name":      "p",
						"owner": map[string]interface{}{
							"login": "o",
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				apiHandler := http.NewServeMux()
				apiHandler.HandleFunc("/repos/o/p/pulls/1", func(w http.ResponseWriter, r *http.Request) {
					resp, err := json.Marshal(map[string]interface{}{
						"title":  "feature 1",
						"number": 1,
						"base": map[string]interface{}{
							"ref": "master",
						},
						"head": map[string]interface{}{
							"ref": "feature",
							"sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
						},
					})
					require.NoError(t, err)

					_, err = w.Write(resp)
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
				})

				apiHandler.HandleFunc("/repos/o/p/pulls/1/commits", func(w http.ResponseWriter, r *http.Request) {
					resp, err := json.Marshal([]map[string]interface{}{
						{
							"commit": map[string]interface{}{
								"message": "commit message",
							},
						},
					})
					require.NoError(t, err)

					_, err = w.Write(resp)
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
				})

				server := httptest.NewServer(apiHandler)
				return server.URL, server.Close
			},
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "codebase1",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitServer:  "github",
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				&codebaseApi.GitServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "github",
						Namespace: "default",
					},
					Spec: codebaseApi.GitServerSpec{
						NameSshKeySecret: "ssh-key-secret",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ssh-key-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						gitServerTokenField: []byte("ssh-privatekey"),
					},
				},
			},
			wantErr: require.NoError,
			want: &event_processor.EventInfo{
				GitProvider:        event_processor.GitProviderGitHub,
				RepoPath:           "/o/r",
				TargetBranch:       "master",
				Type:               event_processor.EventTypeReviewComment,
				HasPipelineRecheck: true,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "codebase1",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitServer:  "github",
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				PullRequest: &event_processor.PullRequest{
					HeadRef:           "feature",
					HeadSha:           "6dcb09b5b57875f334f61aebed695e2e4193db5e",
					Title:             "feature 1",
					ChangeNumber:      1,
					LastCommitMessage: "commit message",
				},
			},
		},
		{
			name: "comment event - should process with no pipeline recheck",
			args: args{
				body: map[string]interface{}{
					"action": "created",
					"issue": map[string]interface{}{
						"number": 1,
					},
					"comment": map[string]interface{}{
						"body": "fix it",
					},
					"repository": map[string]interface{}{
						"full_name": "o/r",
						"name":      "p",
						"owner": map[string]interface{}{
							"login": "o",
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				apiHandler := http.NewServeMux()
				apiHandler.HandleFunc("/repos/o/p/pulls/1", func(w http.ResponseWriter, r *http.Request) {
					resp, err := json.Marshal(map[string]interface{}{
						"title":  "feature 1",
						"number": 1,
						"base": map[string]interface{}{
							"ref": "master",
						},
						"head": map[string]interface{}{
							"ref": "feature",
							"sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
						},
					})
					require.NoError(t, err)

					_, err = w.Write(resp)
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
				})

				apiHandler.HandleFunc("/repos/o/p/pulls/1/commits", func(w http.ResponseWriter, r *http.Request) {
					resp, err := json.Marshal([]map[string]interface{}{
						{
							"commit": map[string]interface{}{
								"message": "commit message",
							},
						},
					})
					require.NoError(t, err)

					_, err = w.Write(resp)
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
				})

				server := httptest.NewServer(apiHandler)
				return server.URL, server.Close
			},
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "codebase1",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitServer:  "github",
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				&codebaseApi.GitServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "github",
						Namespace: "default",
					},
					Spec: codebaseApi.GitServerSpec{
						NameSshKeySecret: "ssh-key-secret",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ssh-key-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						gitServerTokenField: []byte("ssh-privatekey"),
					},
				},
			},
			wantErr: require.NoError,
			want: &event_processor.EventInfo{
				GitProvider:        event_processor.GitProviderGitHub,
				RepoPath:           "/o/r",
				TargetBranch:       "master",
				Type:               event_processor.EventTypeReviewComment,
				HasPipelineRecheck: false,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "codebase1",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitServer:  "github",
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				PullRequest: &event_processor.PullRequest{
					HeadRef:           "feature",
					HeadSha:           "6dcb09b5b57875f334f61aebed695e2e4193db5e",
					Title:             "feature 1",
					ChangeNumber:      1,
					LastCommitMessage: "commit message",
				},
			},
		},
		{
			name: "comment event - should skip none pull request event",
			args: args{
				body: map[string]interface{}{
					"action": "created",
					"issue": map[string]interface{}{
						"number": 1,
					},
					"comment": map[string]interface{}{
						"body": "fix it",
					},
					"repository": map[string]interface{}{
						"full_name": "o/r",
						"name":      "p",
						"owner": map[string]interface{}{
							"login": "o",
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				apiHandler := http.NewServeMux()
				apiHandler.HandleFunc("/repos/o/p/pulls/1", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				})

				server := httptest.NewServer(apiHandler)
				return server.URL, server.Close
			},
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "codebase1",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitServer:  "github",
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				&codebaseApi.GitServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "github",
						Namespace: "default",
					},
					Spec: codebaseApi.GitServerSpec{
						NameSshKeySecret: "ssh-key-secret",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ssh-key-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						gitServerTokenField: []byte("ssh-privatekey"),
					},
				},
			},
			wantErr: require.NoError,
			want: &event_processor.EventInfo{
				GitProvider:        event_processor.GitProviderGitHub,
				Type:               event_processor.EventTypeReviewComment,
				HasPipelineRecheck: false,
			},
		},
		{
			name: "comment event - pull request commits empty",
			args: args{
				body: map[string]interface{}{
					"action": "created",
					"issue": map[string]interface{}{
						"number": 1,
					},
					"comment": map[string]interface{}{
						"body": "/recheck",
					},
					"repository": map[string]interface{}{
						"full_name": "o/r",
						"name":      "p",
						"owner": map[string]interface{}{
							"login": "o",
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				apiHandler := http.NewServeMux()
				apiHandler.HandleFunc("/repos/o/p/pulls/1", func(w http.ResponseWriter, r *http.Request) {
					resp, err := json.Marshal(map[string]interface{}{
						"title":  "feature 1",
						"number": 1,
						"base": map[string]interface{}{
							"ref": "master",
						},
						"head": map[string]interface{}{
							"ref": "feature",
							"sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
						},
					})
					require.NoError(t, err)

					_, err = w.Write(resp)
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
				})

				apiHandler.HandleFunc("/repos/o/p/pulls/1/commits", func(w http.ResponseWriter, r *http.Request) {
					resp, err := json.Marshal([]map[string]interface{}{})
					require.NoError(t, err)

					_, err = w.Write(resp)
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
				})

				server := httptest.NewServer(apiHandler)
				return server.URL, server.Close
			},
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "codebase1",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitServer:  "github",
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				&codebaseApi.GitServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "github",
						Namespace: "default",
					},
					Spec: codebaseApi.GitServerSpec{
						NameSshKeySecret: "ssh-key-secret",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ssh-key-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						gitServerTokenField: []byte("ssh-privatekey"),
					},
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "github pull request commits empty")
			},
		},
		{
			name: "comment event - failed to get pull request commits",
			args: args{
				body: map[string]interface{}{
					"action": "created",
					"issue": map[string]interface{}{
						"number": 1,
					},
					"comment": map[string]interface{}{
						"body": "/recheck",
					},
					"repository": map[string]interface{}{
						"full_name": "o/r",
						"name":      "p",
						"owner": map[string]interface{}{
							"login": "o",
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				apiHandler := http.NewServeMux()
				apiHandler.HandleFunc("/repos/o/p/pulls/1", func(w http.ResponseWriter, r *http.Request) {
					resp, err := json.Marshal(map[string]interface{}{
						"title":  "feature 1",
						"number": 1,
						"base": map[string]interface{}{
							"ref": "master",
						},
						"head": map[string]interface{}{
							"ref": "feature",
							"sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
						},
					})
					require.NoError(t, err)

					_, err = w.Write(resp)
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
				})

				apiHandler.HandleFunc("/repos/o/p/pulls/1/commits", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				})

				server := httptest.NewServer(apiHandler)
				return server.URL, server.Close
			},
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "codebase1",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitServer:  "github",
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				&codebaseApi.GitServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "github",
						Namespace: "default",
					},
					Spec: codebaseApi.GitServerSpec{
						NameSshKeySecret: "ssh-key-secret",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ssh-key-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						gitServerTokenField: []byte("ssh-privatekey"),
					},
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get GitHub pull request commits")
			},
		},
		{
			name: "comment event - failed to get pull request",
			args: args{
				body: map[string]interface{}{
					"action": "created",
					"issue": map[string]interface{}{
						"number": 1,
					},
					"comment": map[string]interface{}{
						"body": "fix it",
					},
					"repository": map[string]interface{}{
						"full_name": "o/r",
						"name":      "p",
						"owner": map[string]interface{}{
							"login": "o",
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				apiHandler := http.NewServeMux()
				apiHandler.HandleFunc("/repos/o/p/pulls/1", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				})

				server := httptest.NewServer(apiHandler)
				return server.URL, server.Close
			},
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "codebase1",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitServer:  "github",
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				&codebaseApi.GitServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "github",
						Namespace: "default",
					},
					Spec: codebaseApi.GitServerSpec{
						NameSshKeySecret: "ssh-key-secret",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ssh-key-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						gitServerTokenField: []byte("ssh-privatekey"),
					},
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get GitHub pull request")
			},
		},
		{
			name: "comment event - failed to get GitServer token",
			args: args{
				body: map[string]interface{}{
					"action": "created",
					"issue": map[string]interface{}{
						"number": 1,
					},
					"comment": map[string]interface{}{
						"body": "fix it",
					},
					"repository": map[string]interface{}{
						"full_name": "o/r",
						"name":      "p",
						"owner": map[string]interface{}{
							"login": "o",
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				apiHandler := http.NewServeMux()
				apiHandler.HandleFunc("/repos/o/p/pulls/1", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				})

				server := httptest.NewServer(apiHandler)
				return server.URL, server.Close
			},
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "codebase1",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitServer:  "github",
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				&codebaseApi.GitServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "github",
						Namespace: "default",
					},
					Spec: codebaseApi.GitServerSpec{
						NameSshKeySecret: "ssh-key-secret",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ssh-key-secret",
						Namespace: "default",
					},
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "token is empty in GitServer secret")
			},
		},
		{
			name: "comment event - failed to get GitServer Secret",
			args: args{
				body: map[string]interface{}{
					"action": "created",
					"issue": map[string]interface{}{
						"number": 1,
					},
					"comment": map[string]interface{}{
						"body": "fix it",
					},
					"repository": map[string]interface{}{
						"full_name": "o/r",
						"name":      "p",
						"owner": map[string]interface{}{
							"login": "o",
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				apiHandler := http.NewServeMux()
				apiHandler.HandleFunc("/repos/o/p/pulls/1", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				})

				server := httptest.NewServer(apiHandler)
				return server.URL, server.Close
			},
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "codebase1",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitServer:  "github",
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				&codebaseApi.GitServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "github",
						Namespace: "default",
					},
					Spec: codebaseApi.GitServerSpec{
						NameSshKeySecret: "ssh-key-secret",
					},
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get GitServer secret")
			},
		},
		{
			name: "comment event - failed to get GitServer",
			args: args{
				body: map[string]interface{}{
					"action": "created",
					"issue": map[string]interface{}{
						"number": 1,
					},
					"comment": map[string]interface{}{
						"body": "fix it",
					},
					"repository": map[string]interface{}{
						"full_name": "o/r",
						"name":      "p",
						"owner": map[string]interface{}{
							"login": "o",
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				apiHandler := http.NewServeMux()
				apiHandler.HandleFunc("/repos/o/p/pulls/1", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				})

				server := httptest.NewServer(apiHandler)
				return server.URL, server.Close
			},
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "codebase1",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitServer:  "github",
						GitUrlPath: pointer.String("/o/r"),
					},
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get GitServer")
			},
		},
		{
			name: "comment event - failed to get Codebase",
			args: args{
				body: map[string]interface{}{
					"action": "created",
					"issue": map[string]interface{}{
						"number": 1,
					},
					"comment": map[string]interface{}{
						"body": "fix it",
					},
					"repository": map[string]interface{}{
						"full_name": "o/r",
						"name":      "p",
						"owner": map[string]interface{}{
							"login": "o",
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				apiHandler := http.NewServeMux()
				apiHandler.HandleFunc("/repos/o/p/pulls/1", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				})

				server := httptest.NewServer(apiHandler)
				return server.URL, server.Close
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "not found")
			},
		},
		{
			name: "comment event - wrong comment action",
			args: args{
				body: map[string]interface{}{
					"action": "deleted",
					"issue": map[string]interface{}{
						"number": 1,
					},
					"comment": map[string]interface{}{
						"body": "fix it",
					},
					"repository": map[string]interface{}{
						"full_name": "o/r",
						"name":      "p",
						"owner": map[string]interface{}{
							"login": "o",
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				apiHandler := http.NewServeMux()
				apiHandler.HandleFunc("/repos/o/p/pulls/1", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				})

				server := httptest.NewServer(apiHandler)
				return server.URL, server.Close
			},
			wantErr: require.NoError,
			want: &event_processor.EventInfo{
				GitProvider:        event_processor.GitProviderGitHub,
				Type:               event_processor.EventTypeReviewComment,
				HasPipelineRecheck: false,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rawURL, teardown := tt.mockhttp(t)
			defer teardown()

			serverURL, err := url.Parse(rawURL)
			require.NoError(t, err)

			p := NewEventProcessor(
				fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.kubeObjects...).Build(),
				&EventProcessorOptions{
					Logger: zap.NewNop().Sugar(),
					GitHubClient: func(ctx context.Context, token string) *github.Client {
						c := github.NewClient(nil)
						c.BaseURL = serverURL.JoinPath("/")

						return c
					},
				},
			)

			body, err := json.Marshal(tt.args.body)
			require.NoError(t, err)

			event, err := p.processCommentEvent(context.Background(), body, "default")
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, event)
		})
	}
}

func TestGitHubEventProcessor_processMergeEvent(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	type args struct {
		body any
	}

	tests := []struct {
		name        string
		args        args
		mockhttp    func(t *testing.T) (URL string, teardown func())
		kubeObjects []client.Object
		want        *event_processor.EventInfo
		wantErr     require.ErrorAssertionFunc
	}{
		{
			name: "merge event - success",
			args: args{
				body: github.PullRequestEvent{
					Number: pointer.Int(1),
					Repo: &github.Repository{
						FullName: pointer.String("o/r"),
						Name:     pointer.String("p"),
						Owner: &github.User{
							Login: pointer.String("o"),
						},
					},
					PullRequest: &github.PullRequest{
						Title: pointer.String("title"),
						Base: &github.PullRequestBranch{
							Ref: pointer.String("master"),
						},
						Head: &github.PullRequestBranch{
							Ref: pointer.String("branch"),
							SHA: pointer.String("sha"),
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				apiHandler := http.NewServeMux()

				apiHandler.HandleFunc("/repos/o/p/pulls/1/commits", func(w http.ResponseWriter, r *http.Request) {
					resp, err := json.Marshal([]map[string]interface{}{
						{
							"commit": map[string]interface{}{
								"message": "commit message",
							},
						},
					})
					require.NoError(t, err)

					_, err = w.Write(resp)
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
				})

				server := httptest.NewServer(apiHandler)
				return server.URL, server.Close
			},
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "codebase1",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: pointer.String("/o/r"),
						GitServer:  "github",
					},
				},
				&codebaseApi.GitServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "github",
						Namespace: "default",
					},
					Spec: codebaseApi.GitServerSpec{
						NameSshKeySecret: "ssh-key-secret",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ssh-key-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						gitServerTokenField: []byte("ssh-privatekey"),
					},
				},
			},
			wantErr: require.NoError,
			want: &event_processor.EventInfo{
				GitProvider:  event_processor.GitProviderGitHub,
				RepoPath:     "/o/r",
				TargetBranch: "master",
				Type:         event_processor.EventTypeMerge,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "codebase1",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: pointer.String("/o/r"),
						GitServer:  "github",
					},
				},
				PullRequest: &event_processor.PullRequest{
					HeadRef:           "branch",
					HeadSha:           "sha",
					Title:             "title",
					ChangeNumber:      1,
					LastCommitMessage: "commit message",
				},
			},
		},
		{
			name: "merge event - failed to get pull request commits",
			args: args{
				body: github.PullRequestEvent{
					Number: pointer.Int(1),
					Repo: &github.Repository{
						FullName: pointer.String("o/r"),
						Name:     pointer.String("p"),
						Owner: &github.User{
							Login: pointer.String("o"),
						},
					},
					PullRequest: &github.PullRequest{
						Title: pointer.String("title"),
						Base: &github.PullRequestBranch{
							Ref: pointer.String("master"),
						},
						Head: &github.PullRequestBranch{
							Ref: pointer.String("branch"),
							SHA: pointer.String("sha"),
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				apiHandler := http.NewServeMux()

				apiHandler.HandleFunc("/repos/o/p/pulls/1/commits", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				})

				server := httptest.NewServer(apiHandler)
				return server.URL, server.Close
			},
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "codebase1",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: pointer.String("/o/r"),
						GitServer:  "github",
					},
				},
				&codebaseApi.GitServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "github",
						Namespace: "default",
					},
					Spec: codebaseApi.GitServerSpec{
						NameSshKeySecret: "ssh-key-secret",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ssh-key-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						gitServerTokenField: []byte("ssh-privatekey"),
					},
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get GitHub pull request commits")
			},
		},
		{
			name: "merge event - no codebase",
			args: args{
				body: github.PullRequestEvent{
					Repo: &github.Repository{
						FullName: pointer.String("o/r"),
					},
					PullRequest: &github.PullRequest{
						Base: &github.PullRequestBranch{
							Ref: pointer.String("master"),
						},
						Head: &github.PullRequestBranch{
							Ref: pointer.String("branch"),
							SHA: pointer.String("sha"),
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				return "", func() {}
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get codebase")
			},
		},
		{
			name: "merge event - no pull request base ref",
			args: args{
				body: github.PullRequestEvent{
					Repo: &github.Repository{
						FullName: pointer.String("o/r"),
					},
					PullRequest: &github.PullRequest{
						Base: &github.PullRequestBranch{},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				return "", func() {}
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "github target branch empty")
			},
		},
		{
			name: "merge event - no pull request base",
			args: args{
				body: github.PullRequestEvent{
					Repo: &github.Repository{
						FullName: pointer.String("o/r"),
					},
					PullRequest: &github.PullRequest{},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				return "", func() {}
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "github target branch empty")
			},
		},
		{
			name: "merge event - no pull request",
			args: args{
				body: github.PullRequestEvent{
					Repo: &github.Repository{
						FullName: pointer.String("o/r"),
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				return "", func() {}
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "github target branch empty")
			},
		},
		{
			name: "merge event - repo full name empty",
			args: args{
				body: github.PullRequestEvent{
					Repo: &github.Repository{},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				return "", func() {}
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "github repository path empty")
			},
		},
		{
			name: "merge event - repo empty",
			args: args{
				body: github.PullRequestEvent{},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				return "", func() {}
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "github repository path empty")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rawURL, teardown := tt.mockhttp(t)
			defer teardown()

			serverURL, err := url.Parse(rawURL)
			require.NoError(t, err)

			p := NewEventProcessor(
				fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.kubeObjects...).Build(),
				&EventProcessorOptions{
					Logger: zap.NewNop().Sugar(),
					GitHubClient: func(ctx context.Context, token string) *github.Client {
						c := github.NewClient(nil)
						c.BaseURL = serverURL.JoinPath("/")

						return c
					},
				},
			)

			body, err := json.Marshal(tt.args.body)
			require.NoError(t, err)

			got, err := p.processMergeEvent(context.Background(), body, "default")
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGitHubEventProcessor_Process(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	type args struct {
		body      any
		eventType string
	}

	tests := []struct {
		name        string
		args        args
		mockhttp    func(t *testing.T) (URL string, teardown func())
		kubeObjects []client.Object
		wantErr     require.ErrorAssertionFunc
		want        *event_processor.EventInfo
	}{
		{
			name: "merge event",
			args: args{
				body: github.PullRequestEvent{
					Number: pointer.Int(1),
					Repo: &github.Repository{
						FullName: pointer.String("o/r"),
						Name:     pointer.String("r"),
						Owner: &github.User{
							Login: pointer.String("o"),
						},
					},
					PullRequest: &github.PullRequest{
						Title: pointer.String("title"),
						Base: &github.PullRequestBranch{
							Ref: pointer.String("master"),
						},
						Head: &github.PullRequestBranch{
							Ref: pointer.String("branch"),
							SHA: pointer.String("sha"),
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				apiHandler := http.NewServeMux()

				apiHandler.HandleFunc("/repos/o/r/pulls/1/commits", func(w http.ResponseWriter, r *http.Request) {
					resp, err := json.Marshal([]map[string]interface{}{
						{
							"commit": map[string]interface{}{
								"message": "commit message",
							},
						},
					})
					require.NoError(t, err)

					_, err = w.Write(resp)
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
				})

				server := httptest.NewServer(apiHandler)
				return server.URL, server.Close
			},
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "codebase1",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: pointer.String("/o/r"),
						GitServer:  "github",
					},
				},
				&codebaseApi.GitServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "github",
						Namespace: "default",
					},
					Spec: codebaseApi.GitServerSpec{
						NameSshKeySecret: "ssh-key-secret",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ssh-key-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						gitServerTokenField: []byte("ssh-privatekey"),
					},
				},
			},
			wantErr: require.NoError,
			want: &event_processor.EventInfo{
				GitProvider:  event_processor.GitProviderGitHub,
				RepoPath:     "/o/r",
				TargetBranch: "master",
				Type:         event_processor.EventTypeMerge,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "codebase1",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: pointer.String("/o/r"),
						GitServer:  "github",
					},
				},
				PullRequest: &event_processor.PullRequest{
					HeadRef:           "branch",
					HeadSha:           "sha",
					Title:             "title",
					ChangeNumber:      1,
					LastCommitMessage: "commit message",
				},
			},
		},
		{
			name: "comment event",
			args: args{
				eventType: event_processor.GitHubEventTypeCommentAdded,
				body: map[string]interface{}{
					"action": "deleted",
					"issue": map[string]interface{}{
						"number": 1,
					},
					"comment": map[string]interface{}{
						"body": "/recheck",
					},
					"repository": map[string]interface{}{
						"full_name": "o/r",
						"name":      "p",
						"owner": map[string]interface{}{
							"login": "o",
						},
					},
				},
			},
			mockhttp: func(t *testing.T) (URL string, teardown func()) {
				return "", func() {}
			},
			wantErr: require.NoError,
			want: &event_processor.EventInfo{
				GitProvider:        event_processor.GitProviderGitHub,
				Type:               event_processor.EventTypeReviewComment,
				HasPipelineRecheck: false,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rawURL, teardown := tt.mockhttp(t)
			defer teardown()

			serverURL, err := url.Parse(rawURL)
			require.NoError(t, err)

			p := &EventProcessor{
				ksClient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.kubeObjects...).Build(),
				logger:   zap.NewNop().Sugar(),
				gitHubClient: func(ctx context.Context, token string) *github.Client {
					c := github.NewClient(nil)
					c.BaseURL = serverURL.JoinPath("/")

					return c
				},
			}

			body, err := json.Marshal(tt.args.body)
			require.NoError(t, err)

			got, err := p.Process(context.Background(), body, "default", tt.args.eventType)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewGitHubEventProcessor(t *testing.T) {
	t.Parallel()

	type args struct {
		ksClient client.Reader
		options  *EventProcessorOptions
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "new github event processor with default options",
			args: args{
				ksClient: fake.NewClientBuilder().Build(),
			},
		},
		{
			name: "new github event processor with options",
			args: args{
				ksClient: fake.NewClientBuilder().Build(),
				options: &EventProcessorOptions{
					Logger: zap.NewNop().Sugar(),
					GitHubClient: func(ctx context.Context, token string) *github.Client {
						return github.NewClient(nil)
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewEventProcessor(tt.args.ksClient, tt.args.options)
			assert.NotNil(t, got)
		})
	}
}
