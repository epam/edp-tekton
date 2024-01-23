package event_processor

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/pkg/apis/edp/v1"
)

func TestGitLabEventProcessor_processMergeEvent(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))

	type args struct {
		body any
	}

	tests := []struct {
		name        string
		args        args
		kubeObjects []client.Object
		wantErr     require.ErrorAssertionFunc
		want        *EventInfo
	}{
		{
			name: "merge event process successfully",
			args: args{
				body: GitLabMergeRequestsEvent{
					Project: GitLabProject{
						PathWithNamespace: "/o/r",
					},
					ObjectAttributes: GitLabMergeRequest{
						TargetBranch: "master",
						Title:        "fix",
						LastCommit: GitLabCommit{
							ID: "123",
						},
						SourceBranch: "feature1",
						ChangeNumber: 1,
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
						GitUrlPath: pointer.String("/o/r"),
					},
				},
			},
			wantErr: require.NoError,
			want: &EventInfo{
				GitProvider:  GitProviderGitLab,
				RepoPath:     "/o/r",
				TargetBranch: "master",
				Type:         EventTypeMerge,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-codebase",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				PullRequest: &PullRequest{
					HeadSha:      "123",
					Title:        "fix",
					HeadRef:      "feature1",
					ChangeNumber: 1,
				},
			},
		},
		{
			name: "failed to get codebase",
			args: args{
				body: GitLabMergeRequestsEvent{
					Project: GitLabProject{
						PathWithNamespace: "/o/r",
					},
					ObjectAttributes: GitLabMergeRequest{
						TargetBranch: "master",
					},
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get codebase")
			},
		},
		{
			name: "failed to get branch",
			args: args{
				body: GitLabMergeRequestsEvent{
					Project: GitLabProject{
						PathWithNamespace: "/o/r",
					},
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "gitlab target branch empty")
			},
		},
		{
			name: "failed to get repository path",
			args: args{
				body: GitLabMergeRequestsEvent{},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "gitlab repository path empty")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tt.args.body)
			require.NoError(t, err)

			p := NewGitLabEventProcessor(
				fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.kubeObjects...).Build(),
				zap.NewNop().Sugar(),
			)

			got, err := p.processMergeEvent(context.Background(), body, "default")

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGitLabEventProcessor_processCommentEvent(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))

	type args struct {
		body any
	}

	tests := []struct {
		name        string
		args        args
		kubeObjects []client.Object
		wantErr     require.ErrorAssertionFunc
		want        *EventInfo
	}{
		{
			name: "comment event process successfully",
			args: args{
				body: GitLabCommentEvent{
					Project: GitLabProject{
						PathWithNamespace: "/o/r",
					},
					MergeRequest: GitLabMergeRequest{
						TargetBranch: "master",
						Title:        "fix",
						LastCommit: GitLabCommit{
							ID: "123",
						},
						SourceBranch: "feature1",
						ChangeNumber: 1,
					},
					ObjectAttributes: GitLabComment{
						Note: "/recheck",
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
						GitUrlPath: pointer.String("/o/r"),
					},
				},
			},
			wantErr: require.NoError,
			want: &EventInfo{
				GitProvider:        GitProviderGitLab,
				RepoPath:           "/o/r",
				TargetBranch:       "master",
				Type:               EventTypeReviewComment,
				HasPipelineRecheck: true,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-codebase",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				PullRequest: &PullRequest{
					HeadRef:      "feature1",
					HeadSha:      "123",
					Title:        "fix",
					ChangeNumber: 1,
				},
			},
		},
		{
			name: "comment event with no recheck",
			args: args{
				body: GitLabCommentEvent{
					Project: GitLabProject{
						PathWithNamespace: "/o/r",
					},
					MergeRequest: GitLabMergeRequest{
						TargetBranch: "master",
						Title:        "fix",
						LastCommit: GitLabCommit{
							ID: "123",
						},
						SourceBranch: "feature1",
						ChangeNumber: 1,
					},
					ObjectAttributes: GitLabComment{
						Note: "no recheck",
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
						GitUrlPath: pointer.String("/o/r"),
					},
				},
			},
			wantErr: require.NoError,
			want: &EventInfo{
				GitProvider:  GitProviderGitLab,
				RepoPath:     "/o/r",
				TargetBranch: "master",
				Type:         EventTypeReviewComment,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-codebase",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				PullRequest: &PullRequest{
					HeadRef:      "feature1",
					HeadSha:      "123",
					Title:        "fix",
					ChangeNumber: 1,
				},
			},
		},
		{
			name: "comment event with no target branch",
			args: args{
				body: GitLabCommentEvent{
					Project: GitLabProject{
						PathWithNamespace: "/o/r",
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
						GitUrlPath: pointer.String("/o/r"),
					},
				},
			},
			wantErr: require.NoError,
			want: &EventInfo{
				GitProvider: GitProviderGitLab,
				Type:        EventTypeReviewComment,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-codebase",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: pointer.String("/o/r"),
					},
				},
			},
		},
		{
			name: "failed to get codebase",
			args: args{
				body: GitLabCommentEvent{
					Project: GitLabProject{
						PathWithNamespace: "/o/r",
					},
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get codebase")
			},
		},
		{
			name: "repository path empty",
			args: args{
				body: GitLabCommentEvent{},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "repository path empty")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tt.args.body)
			require.NoError(t, err)

			p := NewGitLabEventProcessor(
				fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.kubeObjects...).Build(),
				zap.NewNop().Sugar(),
			)

			got, err := p.processCommentEvent(context.Background(), body, "default")

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGitLabEventProcessor_Process(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))

	type args struct {
		body      any
		eventType string
	}

	tests := []struct {
		name        string
		kubeObjects []client.Object
		args        args
		wantErr     require.ErrorAssertionFunc
		want        *EventInfo
	}{
		{
			name: "merge event",
			args: args{
				body: GitLabMergeRequestsEvent{
					Project: GitLabProject{
						PathWithNamespace: "/o/r",
					},
					ObjectAttributes: GitLabMergeRequest{
						TargetBranch: "master",
						Title:        "fix",
						LastCommit: GitLabCommit{
							ID: "123",
						},
						SourceBranch: "feature1",
						ChangeNumber: 1,
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
						GitUrlPath: pointer.String("/o/r"),
					},
				},
			},
			wantErr: require.NoError,
			want: &EventInfo{
				GitProvider:  GitProviderGitLab,
				RepoPath:     "/o/r",
				TargetBranch: "master",
				Type:         EventTypeMerge,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-codebase",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				PullRequest: &PullRequest{
					HeadRef:      "feature1",
					HeadSha:      "123",
					Title:        "fix",
					ChangeNumber: 1,
				},
			},
		},
		{
			name: "comment event",
			args: args{
				eventType: GitLabEventTypeCommentAdded,
				body: GitLabCommentEvent{
					Project: GitLabProject{
						PathWithNamespace: "/o/r",
					},
					MergeRequest: GitLabMergeRequest{
						TargetBranch: "master",
						Title:        "fix",
						LastCommit: GitLabCommit{
							ID: "123",
						},
						SourceBranch: "feature1",
						ChangeNumber: 1,
					},
					ObjectAttributes: GitLabComment{
						Note: "/recheck",
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
						GitUrlPath: pointer.String("/o/r"),
					},
				},
			},
			wantErr: require.NoError,
			want: &EventInfo{
				GitProvider:        GitProviderGitLab,
				RepoPath:           "/o/r",
				TargetBranch:       "master",
				Type:               EventTypeReviewComment,
				HasPipelineRecheck: true,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-codebase",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				PullRequest: &PullRequest{
					HeadRef:      "feature1",
					HeadSha:      "123",
					Title:        "fix",
					ChangeNumber: 1,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tt.args.body)
			require.NoError(t, err)

			p := NewGitLabEventProcessor(
				fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.kubeObjects...).Build(),
				zap.NewNop().Sugar(),
			)

			got, err := p.Process(context.Background(), body, "default", tt.args.eventType)

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
