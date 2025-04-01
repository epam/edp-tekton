package gerrit

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/event_processor"
)

func TestGerritEventProcessor_Process(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))

	type args struct {
		body any
	}

	tests := []struct {
		name        string
		kubeObjects []client.Object
		args        args
		wantErr     require.ErrorAssertionFunc
		want        *event_processor.EventInfo
	}{
		{
			name: "change event process successfully",
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-codebase",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: "/test-repo",
					},
				},
			},
			args: args{
				event_processor.GerritEvent{
					Project: struct {
						Name string `json:"name"`
					}{
						Name: "test-repo",
					},
					Change: struct {
						Branch string `json:"branch"`
					}{
						Branch: "test-branch",
					},
					Type: "default",
				},
			},
			wantErr: require.NoError,
			want: &event_processor.EventInfo{
				GitProvider:  event_processor.GitProviderGerrit,
				RepoPath:     "test-repo",
				TargetBranch: "test-branch",
				Type:         event_processor.EventTypeMerge,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-codebase",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: "/test-repo",
					},
				},
			},
		},
		{
			name: "comment event process successfully",
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-codebase",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: "/test-repo",
					},
				},
			},
			args: args{
				event_processor.GerritEvent{
					Project: struct {
						Name string `json:"name"`
					}{
						Name: "test-repo",
					},
					Change: struct {
						Branch string `json:"branch"`
					}{
						Branch: "test-branch",
					},
					Type:    event_processor.GerritEventTypeCommentAdded,
					Comment: "/recheck",
				},
			},
			wantErr: require.NoError,
			want: &event_processor.EventInfo{
				GitProvider:        event_processor.GitProviderGerrit,
				RepoPath:           "test-repo",
				TargetBranch:       "test-branch",
				Type:               event_processor.EventTypeReviewComment,
				HasPipelineRecheck: true,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-codebase",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: "/test-repo",
					},
				},
			},
		},
		{
			name: "comment event process successfully - OkToTestComment",
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-codebase",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: "/test-repo",
					},
				},
			},
			args: args{
				event_processor.GerritEvent{
					Project: struct {
						Name string `json:"name"`
					}{
						Name: "test-repo",
					},
					Change: struct {
						Branch string `json:"branch"`
					}{
						Branch: "test-branch",
					},
					Type:    event_processor.GerritEventTypeCommentAdded,
					Comment: event_processor.OkToTestComment,
				},
			},
			wantErr: require.NoError,
			want: &event_processor.EventInfo{
				GitProvider:        event_processor.GitProviderGerrit,
				RepoPath:           "test-repo",
				TargetBranch:       "test-branch",
				Type:               event_processor.EventTypeReviewComment,
				HasPipelineRecheck: true,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-codebase",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: "/test-repo",
					},
				},
			},
		},
		{
			name: "comment event with no recheck",
			kubeObjects: []client.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-codebase",
						Namespace: "default",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: "/test-repo",
					},
				},
			},
			args: args{
				event_processor.GerritEvent{
					Project: struct {
						Name string `json:"name"`
					}{
						Name: "test-repo",
					},
					Change: struct {
						Branch string `json:"branch"`
					}{
						Branch: "test-branch",
					},
					Type:    event_processor.GerritEventTypeCommentAdded,
					Comment: "fix it",
				},
			},
			wantErr: require.NoError,
			want: &event_processor.EventInfo{
				GitProvider:  event_processor.GitProviderGerrit,
				RepoPath:     "test-repo",
				TargetBranch: "test-branch",
				Type:         event_processor.EventTypeReviewComment,
				Codebase: &codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-codebase",
						Namespace:       "default",
						ResourceVersion: "999",
					},
					Spec: codebaseApi.CodebaseSpec{
						GitUrlPath: "/test-repo",
					},
				},
			},
		},
		{
			name: "failed to get codebase",
			args: args{
				event_processor.GerritEvent{
					Project: struct {
						Name string `json:"name"`
					}{
						Name: "test-repo",
					},
					Change: struct {
						Branch string `json:"branch"`
					}{
						Branch: "test-branch",
					},
					Type: "default",
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
				event_processor.GerritEvent{
					Project: struct {
						Name string `json:"name"`
					}{
						Name: "test-repo",
					},
					Type: "default",
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "gerrit target branch empty")
			},
		},
		{
			name: "failed to get repository path",
			args: args{
				event_processor.GerritEvent{
					Type: "default",
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "gerrit repository path empty")
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
				zap.NewNop().Sugar(),
			)
			got, err := p.Process(context.Background(), body, "default", "")

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
