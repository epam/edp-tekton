package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v81/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/event_processor"
)

// createTestKubeObjects creates common Kubernetes objects for GitHub testing.
func createTestKubeObjects() []client.Object {
	return []client.Object{
		&codebaseApi.Codebase{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "codebase1",
				Namespace: "default",
			},
			Spec: codebaseApi.CodebaseSpec{
				GitServer:  "github",
				GitUrlPath: "/o/r",
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
				event_processor.GitServerTokenField: []byte("ssh-privatekey"),
			},
		},
	}
}

// mockGitHubPullRequestAPI creates a mock HTTP server for GitHub pull request API calls.
func mockGitHubPullRequestAPI(
	t *testing.T,
	owner string,
	repo string,
	prNumber int,
	title string,
	baseBranch string,
	headBranch string,
	headSHA string,
	commitMessage string,
) (URL string, teardown func()) {
	apiHandler := http.NewServeMux()
	apiHandler.HandleFunc("/repos/"+owner+"/"+repo+"/pulls/1", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		resp, err := json.Marshal(map[string]interface{}{
			"title":  title,
			"number": prNumber,
			"base": map[string]interface{}{
				"ref": baseBranch,
			},
			"head": map[string]interface{}{
				"ref": headBranch,
				"sha": headSHA,
			},
		})
		require.NoError(t, err)

		_, err = w.Write(resp)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	})

	apiHandler.HandleFunc("/repos/"+owner+"/"+repo+"/pulls/1/commits", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		resp, err := json.Marshal([]map[string]interface{}{
			{
				"commit": map[string]interface{}{
					"message": commitMessage,
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
}

// gitHubTestCase defines the structure for GitHub event processor test cases.
type gitHubTestCase struct {
	name        string
	args        gitHubTestArgs
	mockhttp    func(t *testing.T) (URL string, teardown func())
	kubeObjects []client.Object
	want        *event_processor.EventInfo
	wantErr     require.ErrorAssertionFunc
}

type gitHubTestArgs struct {
	body any
}

// createGitHubCommentTestCase creates a test case for GitHub comment event processing.
func createGitHubCommentTestCase(
	name string,
	commentBody string,
	hasPipelineRecheck bool,
) gitHubTestCase {
	return gitHubTestCase{
		name: name,
		args: gitHubTestArgs{
			body: map[string]interface{}{
				"action": "created",
				"issue": map[string]interface{}{
					"number": 1,
				},
				"comment": map[string]interface{}{
					"body": commentBody,
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
			return mockGitHubPullRequestAPI(
				t,
				"o",
				"p",
				1,
				"feature 1",
				"master",
				"feature",
				"6dcb09b5b57875f334f61aebed695e2e4193db5e",
				"commit message",
			)
		},
		kubeObjects: createTestKubeObjects(),
		wantErr:     require.NoError,
		want: &event_processor.EventInfo{
			GitProvider:        event_processor.GitProviderGitHub,
			RepoPath:           "/o/r",
			TargetBranch:       "master",
			Type:               event_processor.EventTypeReviewComment,
			HasPipelineRecheck: hasPipelineRecheck,
			Codebase: &codebaseApi.Codebase{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "codebase1",
					Namespace:       "default",
					ResourceVersion: "999",
				},
				Spec: codebaseApi.CodebaseSpec{
					GitServer:  "github",
					GitUrlPath: "/o/r",
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
	}
}

// runGitHubEventProcessorTest is a helper function to run GitHub event processor tests.
func runGitHubEventProcessorTest(
	t *testing.T,
	scheme *runtime.Scheme,
	tt gitHubTestCase,
	processFunc func(
		p *EventProcessor,
		ctx context.Context,
		body []byte,
		namespace string,
	) (*event_processor.EventInfo, error),
) {
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

		got, err := processFunc(p, context.Background(), body, "default")
		tt.wantErr(t, err)
		assert.Equal(t, tt.want, got)
	})
}

func TestGitHubEventProcessor_processCommentEvent(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	tests := []gitHubTestCase{
		createGitHubCommentTestCase(
			"comment event - should process pull request",
			"/recheck",
			true,
		),
		{
			name: "comment event OkToTestComment - should process pull request",
			args: gitHubTestArgs{
				body: map[string]interface{}{
					"action": "created",
					"issue": map[string]interface{}{
						"number": 1,
					},
					"comment": map[string]interface{}{
						"body": event_processor.OkToTestComment,
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
			kubeObjects: createTestKubeObjects(),
			wantErr:     require.NoError,
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
						GitUrlPath: "/o/r",
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
		createGitHubCommentTestCase(
			"comment event - should process with no pipeline recheck",
			"fix it",
			false,
		),
		{
			name: "comment event - should skip none pull request event",
			args: gitHubTestArgs{
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
			kubeObjects: createTestKubeObjects(),
			wantErr:     require.NoError,
			want: &event_processor.EventInfo{
				GitProvider:        event_processor.GitProviderGitHub,
				Type:               event_processor.EventTypeReviewComment,
				HasPipelineRecheck: false,
			},
		},
		{
			name: "comment event - pull request commits empty",
			args: gitHubTestArgs{
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
			kubeObjects: createTestKubeObjects(),
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "github pull request commits empty")
			},
		},
		{
			name: "comment event - failed to get pull request commits",
			args: gitHubTestArgs{
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
			kubeObjects: createTestKubeObjects(),
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get GitHub pull request commits")
			},
		},
		{
			name: "comment event - failed to get pull request",
			args: gitHubTestArgs{
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
			kubeObjects: createTestKubeObjects(),
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get GitHub pull request")
			},
		},
		{
			name: "comment event - failed to get GitServer token",
			args: gitHubTestArgs{
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
						GitUrlPath: "/o/r",
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
			args: gitHubTestArgs{
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
						GitUrlPath: "/o/r",
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
			args: gitHubTestArgs{
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
						GitUrlPath: "/o/r",
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
			args: gitHubTestArgs{
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
			args: gitHubTestArgs{
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
		runGitHubEventProcessorTest(
			t,
			scheme,
			tt,
			func(
				p *EventProcessor,
				ctx context.Context,
				body []byte,
				namespace string,
			) (*event_processor.EventInfo, error) {
				return p.processCommentEvent(ctx, body, namespace)
			},
		)
	}
}

func TestGitHubEventProcessor_processMergeEvent(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	tests := []gitHubTestCase{
		{
			name: "merge event - success",
			args: gitHubTestArgs{
				body: github.PullRequestEvent{
					Number: ptr.To(1),
					Repo: &github.Repository{
						FullName: ptr.To("o/r"),
						Name:     ptr.To("p"),
						Owner: &github.User{
							Login: ptr.To("o"),
						},
					},
					PullRequest: &github.PullRequest{
						Title: ptr.To("title"),
						Base: &github.PullRequestBranch{
							Ref: ptr.To("master"),
						},
						Head: &github.PullRequestBranch{
							Ref: ptr.To("branch"),
							SHA: ptr.To("sha"),
						},
						User: &github.User{
							Login:     ptr.To("prauthor"),
							AvatarURL: ptr.To("https://avatars.githubusercontent.com/u/456"),
						},
						HTMLURL: ptr.To("https://github.com/o/r/pull/1"),
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
						GitUrlPath: "/o/r",
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
						event_processor.GitServerTokenField: []byte("ssh-privatekey"),
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
						GitUrlPath: "/o/r",
						GitServer:  "github",
					},
				},
				PullRequest: &event_processor.PullRequest{
					HeadRef:           "branch",
					HeadSha:           "sha",
					Title:             "title",
					ChangeNumber:      1,
					LastCommitMessage: "commit message",
					Author:            "prauthor",
					AuthorAvatarUrl:   "https://avatars.githubusercontent.com/u/456",
					Url:               "https://github.com/o/r/pull/1",
				},
			},
		},
		{
			name: "merge event - failed to get pull request commits",
			args: gitHubTestArgs{
				body: github.PullRequestEvent{
					Number: ptr.To(1),
					Repo: &github.Repository{
						FullName: ptr.To("o/r"),
						Name:     ptr.To("p"),
						Owner: &github.User{
							Login: ptr.To("o"),
						},
					},
					PullRequest: &github.PullRequest{
						Title: ptr.To("title"),
						Base: &github.PullRequestBranch{
							Ref: ptr.To("master"),
						},
						Head: &github.PullRequestBranch{
							Ref: ptr.To("branch"),
							SHA: ptr.To("sha"),
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
						GitUrlPath: "/o/r",
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
						event_processor.GitServerTokenField: []byte("ssh-privatekey"),
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
			args: gitHubTestArgs{
				body: github.PullRequestEvent{
					Repo: &github.Repository{
						FullName: ptr.To("o/r"),
					},
					PullRequest: &github.PullRequest{
						Base: &github.PullRequestBranch{
							Ref: ptr.To("master"),
						},
						Head: &github.PullRequestBranch{
							Ref: ptr.To("branch"),
							SHA: ptr.To("sha"),
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
			args: gitHubTestArgs{
				body: github.PullRequestEvent{
					Repo: &github.Repository{
						FullName: ptr.To("o/r"),
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
			args: gitHubTestArgs{
				body: github.PullRequestEvent{
					Repo: &github.Repository{
						FullName: ptr.To("o/r"),
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
			args: gitHubTestArgs{
				body: github.PullRequestEvent{
					Repo: &github.Repository{
						FullName: ptr.To("o/r"),
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
			args: gitHubTestArgs{
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
			args: gitHubTestArgs{
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
		runGitHubEventProcessorTest(
			t,
			scheme,
			tt,
			func(
				p *EventProcessor,
				ctx context.Context,
				body []byte,
				namespace string,
			) (*event_processor.EventInfo, error) {
				return p.processMergeEvent(ctx, body, namespace)
			},
		)
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
					Number: ptr.To(1),
					Repo: &github.Repository{
						FullName: ptr.To("o/r"),
						Name:     ptr.To("r"),
						Owner: &github.User{
							Login: ptr.To("o"),
						},
					},
					PullRequest: &github.PullRequest{
						Title: ptr.To("title"),
						Base: &github.PullRequestBranch{
							Ref: ptr.To("master"),
						},
						Head: &github.PullRequestBranch{
							Ref: ptr.To("branch"),
							SHA: ptr.To("sha"),
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
						GitUrlPath: "/o/r",
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
						event_processor.GitServerTokenField: []byte("ssh-privatekey"),
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
						GitUrlPath: "/o/r",
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewEventProcessor(tt.args.ksClient, tt.args.options)
			assert.NotNil(t, got)
		})
	}
}
