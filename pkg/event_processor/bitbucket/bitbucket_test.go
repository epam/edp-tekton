package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/event_processor"
)

// createTestKubeObjects creates common Kubernetes objects for testing.
func createTestKubeObjects() []client.Object {
	return []client.Object{
		&codebaseApi.Codebase{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-codebase",
				Namespace: "default",
			},
			Spec: codebaseApi.CodebaseSpec{
				GitUrlPath: "/o/r",
				GitServer:  "test-git-server",
			},
		},
		&codebaseApi.GitServer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-git-server",
				Namespace: "default",
			},
			Spec: codebaseApi.GitServerSpec{
				NameSshKeySecret: "test-secret",
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-secret",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"token": []byte("test-token"),
			},
		},
	}
}

func TestBitbucketEventProcessor_processCommentEvent(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "pullrequests/1") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"values": [{"message": "commit message"}]}`))

			return
		}

		if strings.Contains(r.URL.Path, "pullrequests/2") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))

			return
		}
	}))

	t.Cleanup(server.Close)

	type args struct {
		body any
	}

	tests := []struct {
		name        string
		args        args
		kubeObjects []client.Object
		wantErr     require.ErrorAssertionFunc
		want        *event_processor.EventInfo
	}{
		{
			name: "comment event process successfully",
			args: args{
				body: event_processor.BitbucketEvent{
					Repository: event_processor.BitbucketRepository{
						FullName: "o/r",
					},
					PullRequest: event_processor.BitbucketPullRequest{
						ID:    1,
						Title: "fix",
						Source: event_processor.BitbucketPullRequestSrc{
							Branch: event_processor.BitbucketBranch{
								Name: "feature1",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "123",
							},
						},
						Destination: event_processor.BitbucketPullRequestDest{
							Branch: event_processor.BitbucketBranch{
								Name: "master",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "456",
							},
						},
						LastCommit: event_processor.BitbucketCommit{
							Hash: "123",
						},
						Author: event_processor.BitbucketAuthor{
							DisplayName: "bbuser",
							Links: struct {
								Avatar struct {
									Href string `json:"href"`
								} `json:"avatar"`
							}{
								Avatar: struct {
									Href string `json:"href"`
								}{
									Href: "https://bitbucket.org/avatar/bbuser",
								},
							},
						},
						Links: struct {
							Html struct {
								Href string `json:"href"`
							} `json:"html"`
						}{
							Html: struct {
								Href string `json:"href"`
							}{
								Href: "https://bitbucket.org/o/r/pull-requests/1",
							},
						},
					},
					Comment: event_processor.BitbucketComment{
						Content: event_processor.BitbucketCommentContent{
							Raw: "/recheck",
						},
					},
				},
			},
			kubeObjects: createTestKubeObjects(),
			wantErr:     require.NoError,
			want: &event_processor.EventInfo{
				GitProvider:        event_processor.GitProviderBitbucket,
				RepoPath:           "/o/r",
				TargetBranch:       "master",
				Type:               event_processor.EventTypeReviewComment,
				HasPipelineRecheck: true,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-codebase",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: "/o/r",
						GitServer:  "test-git-server",
					},
				},
				PullRequest: &event_processor.PullRequest{
					HeadRef:           "feature1",
					HeadSha:           "123",
					Title:             "fix",
					ChangeNumber:      1,
					LastCommitMessage: "commit message",
					Author:            "bbuser",
					AuthorAvatarUrl:   "https://bitbucket.org/avatar/bbuser",
					Url:               "https://bitbucket.org/o/r/pull-requests/1",
				},
			},
		},
		{
			name: "pr doesn't contain commits",
			args: args{
				body: event_processor.BitbucketEvent{
					Repository: event_processor.BitbucketRepository{
						FullName: "o/r",
					},
					PullRequest: event_processor.BitbucketPullRequest{
						ID:    2,
						Title: "fix",
						Source: event_processor.BitbucketPullRequestSrc{
							Branch: event_processor.BitbucketBranch{
								Name: "feature1",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "123",
							},
						},
						Destination: event_processor.BitbucketPullRequestDest{
							Branch: event_processor.BitbucketBranch{
								Name: "master",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "456",
							},
						},
						LastCommit: event_processor.BitbucketCommit{
							Hash: "123",
						},
					},
					Comment: event_processor.BitbucketComment{
						Content: event_processor.BitbucketCommentContent{
							Raw: "/recheck",
						},
					},
				},
			},
			kubeObjects: createTestKubeObjects(),
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "pull request doesn't have commits")
			},
			want: nil,
		},
		{
			name: "repository path empty",
			args: args{
				body: event_processor.BitbucketEvent{},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "repository path empty")
			},
		},
		{
			name: "failed to get codebase",
			args: args{
				body: event_processor.BitbucketEvent{
					Repository: event_processor.BitbucketRepository{
						FullName: "o/r",
					},
					PullRequest: event_processor.BitbucketPullRequest{
						ID:    1,
						Title: "fix",
						Source: event_processor.BitbucketPullRequestSrc{
							Branch: event_processor.BitbucketBranch{
								Name: "feature1",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "123",
							},
						},
						Destination: event_processor.BitbucketPullRequestDest{
							Branch: event_processor.BitbucketBranch{
								Name: "master",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "456",
							},
						},
						LastCommit: event_processor.BitbucketCommit{
							Hash: "123",
						},
					},
					Comment: event_processor.BitbucketComment{
						Content: event_processor.BitbucketCommentContent{
							Raw: "/recheck",
						},
					},
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get codebase by repo path")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tt.args.body)

			require.NoError(t, err)

			p := NewEventProcessor(
				fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.kubeObjects...).Build(),
				&EventProcessorOptions{
					Logger:      zap.NewNop().Sugar(),
					RestyClient: resty.New().SetBaseURL(server.URL),
				},
			)
			got, err := p.Process(context.Background(), body, "default", event_processor.BitbucketEventTypeCommentAdded)

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEventProcessor_processMergeEvent(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "pullrequests/1") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"values": [{"message": "commit message"}]}`))

			return
		}

		if strings.Contains(r.URL.Path, "pullrequests/2") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))

			return
		}

		if strings.Contains(r.URL.Path, "pullrequests/3") {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
	}))

	t.Cleanup(server.Close)

	type args struct {
		body any
	}

	tests := []struct {
		name        string
		args        args
		kubeObjects []client.Object
		wantErr     require.ErrorAssertionFunc
		want        *event_processor.EventInfo
	}{
		{
			name: "merge event process successfully",
			args: args{
				body: event_processor.BitbucketEvent{
					Repository: event_processor.BitbucketRepository{
						FullName: "o/r",
					},
					PullRequest: event_processor.BitbucketPullRequest{
						ID:    1,
						Title: "fix",
						Source: event_processor.BitbucketPullRequestSrc{
							Branch: event_processor.BitbucketBranch{
								Name: "feature1",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "123",
							},
						},
						Destination: event_processor.BitbucketPullRequestDest{
							Branch: event_processor.BitbucketBranch{
								Name: "master",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "456",
							},
						},
						LastCommit: event_processor.BitbucketCommit{
							Hash: "123",
						},
						Author: event_processor.BitbucketAuthor{
							DisplayName: "bbmergeuser",
							Links: struct {
								Avatar struct {
									Href string `json:"href"`
								} `json:"avatar"`
							}{
								Avatar: struct {
									Href string `json:"href"`
								}{
									Href: "https://bitbucket.org/avatar/bbmergeuser",
								},
							},
						},
						Links: struct {
							Html struct {
								Href string `json:"href"`
							} `json:"html"`
						}{
							Html: struct {
								Href string `json:"href"`
							}{
								Href: "https://bitbucket.org/o/r/pull-requests/1",
							},
						},
					},
				},
			},
			kubeObjects: createTestKubeObjects(),
			wantErr:     require.NoError,
			want: &event_processor.EventInfo{
				GitProvider:  event_processor.GitProviderBitbucket,
				RepoPath:     "/o/r",
				TargetBranch: "master",
				Type:         event_processor.EventTypeMerge,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-codebase",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: "/o/r",
						GitServer:  "test-git-server",
					},
				},
				PullRequest: &event_processor.PullRequest{
					HeadRef:           "feature1",
					HeadSha:           "123",
					Title:             "fix",
					ChangeNumber:      1,
					LastCommitMessage: "commit message",
					Author:            "bbmergeuser",
					AuthorAvatarUrl:   "https://bitbucket.org/avatar/bbmergeuser",
					Url:               "https://bitbucket.org/o/r/pull-requests/1",
				},
			},
		},
		{
			name: "pr doesn't contain commits",
			args: args{
				body: event_processor.BitbucketEvent{
					Repository: event_processor.BitbucketRepository{
						FullName: "o/r",
					},
					PullRequest: event_processor.BitbucketPullRequest{
						ID:    2,
						Title: "fix",
						Source: event_processor.BitbucketPullRequestSrc{
							Branch: event_processor.BitbucketBranch{
								Name: "feature1",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "123",
							},
						},
						Destination: event_processor.BitbucketPullRequestDest{
							Branch: event_processor.BitbucketBranch{
								Name: "master",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "456",
							},
						},
						LastCommit: event_processor.BitbucketCommit{
							Hash: "123",
						},
					},
				},
			},
			kubeObjects: createTestKubeObjects(),
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "pull request doesn't have commits")
			},
			want: nil,
		},
		{
			name: "repository path empty",
			args: args{
				body: event_processor.BitbucketEvent{},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "repository path empty")
			},
		},
		{
			name: "failed to get codebase",
			args: args{
				body: event_processor.BitbucketEvent{
					Repository: event_processor.BitbucketRepository{
						FullName: "o/r",
					},
					PullRequest: event_processor.BitbucketPullRequest{
						ID:    1,
						Title: "fix",
						Source: event_processor.BitbucketPullRequestSrc{
							Branch: event_processor.BitbucketBranch{
								Name: "feature1",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "123",
							},
						},
						Destination: event_processor.BitbucketPullRequestDest{
							Branch: event_processor.BitbucketBranch{
								Name: "master",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "456",
							},
						},
						LastCommit: event_processor.BitbucketCommit{
							Hash: "123",
						},
					},
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get codebase by repo path")
			},
		},
		{
			name: "failed to get pr commits",
			args: args{
				body: event_processor.BitbucketEvent{
					Repository: event_processor.BitbucketRepository{
						FullName: "o/r",
					},
					PullRequest: event_processor.BitbucketPullRequest{
						ID:    3,
						Title: "fix",
						Source: event_processor.BitbucketPullRequestSrc{
							Branch: event_processor.BitbucketBranch{
								Name: "feature1",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "123",
							},
						},
						Destination: event_processor.BitbucketPullRequestDest{
							Branch: event_processor.BitbucketBranch{
								Name: "master",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "456",
							},
						},
						LastCommit: event_processor.BitbucketCommit{
							Hash: "123",
						},
					},
				},
			},
			kubeObjects: createTestKubeObjects(),
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get PR latest commit message")
			},
			want: nil,
		},
		{
			name: "failed to get git server token",
			args: args{
				body: event_processor.BitbucketEvent{
					Repository: event_processor.BitbucketRepository{
						FullName: "o/r",
					},
					PullRequest: event_processor.BitbucketPullRequest{
						ID:    3,
						Title: "fix",
						Source: event_processor.BitbucketPullRequestSrc{
							Branch: event_processor.BitbucketBranch{
								Name: "feature1",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "123",
							},
						},
						Destination: event_processor.BitbucketPullRequestDest{
							Branch: event_processor.BitbucketBranch{
								Name: "master",
							},
							Commit: event_processor.BitbucketCommit{
								Hash: "456",
							},
						},
						LastCommit: event_processor.BitbucketCommit{
							Hash: "123",
						},
					},
				},
			},
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-codebase",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: "/o/r",
						GitServer:  "test-git-server",
					},
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get git server token for Bitbucket")
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tt.args.body)

			require.NoError(t, err)

			p := NewEventProcessor(
				fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.kubeObjects...).Build(),
				&EventProcessorOptions{
					Logger:      zap.NewNop().Sugar(),
					RestyClient: resty.New().SetBaseURL(server.URL),
				},
			)
			got, err := p.Process(context.Background(), body, "default", "")

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
