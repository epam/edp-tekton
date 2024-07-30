package gitlab

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

	"github.com/epam/edp-tekton/pkg/event_processor"
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
		want        *event_processor.EventInfo
	}{
		{
			name: "merge event process successfully",
			args: args{
				body: event_processor.GitLabMergeRequestsEvent{
					Project: event_processor.GitLabProject{
						PathWithNamespace: "/o/r",
					},
					ObjectAttributes: event_processor.GitLabMergeRequest{
						TargetBranch: "master",
						Title:        "fix",
						LastCommit: event_processor.GitLabCommit{
							ID:      "123",
							Message: "commit message",
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
			want: &event_processor.EventInfo{
				GitProvider:  event_processor.GitProviderGitLab,
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
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				PullRequest: &event_processor.PullRequest{
					HeadSha:           "123",
					Title:             "fix",
					HeadRef:           "feature1",
					ChangeNumber:      1,
					LastCommitMessage: "commit message",
				},
			},
		},
		{
			name: "failed to get codebase",
			args: args{
				body: event_processor.GitLabMergeRequestsEvent{
					Project: event_processor.GitLabProject{
						PathWithNamespace: "/o/r",
					},
					ObjectAttributes: event_processor.GitLabMergeRequest{
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
				body: event_processor.GitLabMergeRequestsEvent{
					Project: event_processor.GitLabProject{
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
				body: event_processor.GitLabMergeRequestsEvent{},
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

			p := NewEventProcessor(
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
		want        *event_processor.EventInfo
	}{
		{
			name: "comment event process successfully",
			args: args{
				body: event_processor.GitLabCommentEvent{
					Project: event_processor.GitLabProject{
						PathWithNamespace: "/o/r",
					},
					MergeRequest: event_processor.GitLabMergeRequest{
						TargetBranch: "master",
						Title:        "fix",
						LastCommit: event_processor.GitLabCommit{
							ID:      "123",
							Message: "commit message",
						},
						SourceBranch: "feature1",
						ChangeNumber: 1,
					},
					ObjectAttributes: event_processor.GitLabComment{
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
			want: &event_processor.EventInfo{
				GitProvider:        event_processor.GitProviderGitLab,
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
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				PullRequest: &event_processor.PullRequest{
					HeadRef:           "feature1",
					HeadSha:           "123",
					Title:             "fix",
					ChangeNumber:      1,
					LastCommitMessage: "commit message",
				},
			},
		},
		{
			name: "comment event process successfully - OkToTestComment",
			args: args{
				body: event_processor.GitLabCommentEvent{
					Project: event_processor.GitLabProject{
						PathWithNamespace: "/o/r",
					},
					MergeRequest: event_processor.GitLabMergeRequest{
						TargetBranch: "master",
						Title:        "fix",
						LastCommit: event_processor.GitLabCommit{
							ID:      "123",
							Message: "commit message",
						},
						SourceBranch: "feature1",
						ChangeNumber: 1,
					},
					ObjectAttributes: event_processor.GitLabComment{
						Note: event_processor.OkToTestComment,
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
			want: &event_processor.EventInfo{
				GitProvider:        event_processor.GitProviderGitLab,
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
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				PullRequest: &event_processor.PullRequest{
					HeadRef:           "feature1",
					HeadSha:           "123",
					Title:             "fix",
					ChangeNumber:      1,
					LastCommitMessage: "commit message",
				},
			},
		},
		{
			name: "comment event with no recheck",
			args: args{
				body: event_processor.GitLabCommentEvent{
					Project: event_processor.GitLabProject{
						PathWithNamespace: "/o/r",
					},
					MergeRequest: event_processor.GitLabMergeRequest{
						TargetBranch: "master",
						Title:        "fix",
						LastCommit: event_processor.GitLabCommit{
							ID:      "123",
							Message: "commit message",
						},
						SourceBranch: "feature1",
						ChangeNumber: 1,
					},
					ObjectAttributes: event_processor.GitLabComment{
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
			want: &event_processor.EventInfo{
				GitProvider:  event_processor.GitProviderGitLab,
				RepoPath:     "/o/r",
				TargetBranch: "master",
				Type:         event_processor.EventTypeReviewComment,
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
				PullRequest: &event_processor.PullRequest{
					HeadRef:           "feature1",
					HeadSha:           "123",
					Title:             "fix",
					ChangeNumber:      1,
					LastCommitMessage: "commit message",
				},
			},
		},
		{
			name: "comment event with no target branch",
			args: args{
				body: event_processor.GitLabCommentEvent{
					Project: event_processor.GitLabProject{
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
			want: &event_processor.EventInfo{
				GitProvider: event_processor.GitProviderGitLab,
				Type:        event_processor.EventTypeReviewComment,
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
				body: event_processor.GitLabCommentEvent{
					Project: event_processor.GitLabProject{
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
				body: event_processor.GitLabCommentEvent{},
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

			p := NewEventProcessor(
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
		want        *event_processor.EventInfo
	}{
		{
			name: "merge event",
			args: args{
				body: event_processor.GitLabMergeRequestsEvent{
					Project: event_processor.GitLabProject{
						PathWithNamespace: "/o/r",
					},
					ObjectAttributes: event_processor.GitLabMergeRequest{
						TargetBranch: "master",
						Title:        "fix",
						LastCommit: event_processor.GitLabCommit{
							ID:      "123",
							Message: "commit message",
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
			want: &event_processor.EventInfo{
				GitProvider:  event_processor.GitProviderGitLab,
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
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				PullRequest: &event_processor.PullRequest{
					HeadRef:           "feature1",
					HeadSha:           "123",
					Title:             "fix",
					ChangeNumber:      1,
					LastCommitMessage: "commit message",
				},
			},
		},
		{
			name: "comment event",
			args: args{
				eventType: event_processor.GitLabEventTypeCommentAdded,
				body: event_processor.GitLabCommentEvent{
					Project: event_processor.GitLabProject{
						PathWithNamespace: "/o/r",
					},
					MergeRequest: event_processor.GitLabMergeRequest{
						TargetBranch: "master",
						Title:        "fix",
						LastCommit: event_processor.GitLabCommit{
							ID:      "123",
							Message: "commit message",
						},
						SourceBranch: "feature1",
						ChangeNumber: 1,
					},
					ObjectAttributes: event_processor.GitLabComment{
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
			want: &event_processor.EventInfo{
				GitProvider:        event_processor.GitProviderGitLab,
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
						GitUrlPath: pointer.String("/o/r"),
					},
				},
				PullRequest: &event_processor.PullRequest{
					HeadRef:           "feature1",
					HeadSha:           "123",
					Title:             "fix",
					ChangeNumber:      1,
					LastCommitMessage: "commit message",
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

			p := NewEventProcessor(
				fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.kubeObjects...).Build(),
				zap.NewNop().Sugar(),
			)

			got, err := p.Process(context.Background(), body, "default", tt.args.eventType)

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
