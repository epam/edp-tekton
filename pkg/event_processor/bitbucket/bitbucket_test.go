package bitbucket

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

func TestBitbucketEventProcessor_processCommentEvent(t *testing.T) {
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
				body: event_processor.BitbucketCommentEvent{
					BitbucketEvent: event_processor.BitbucketEvent{
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
					Comment: event_processor.BitbucketComment{
						Content: event_processor.BitbucketCommentContent{
							Raw: "/recheck",
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
					},
				},
			},
			wantErr: require.NoError,
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
					},
				},
				PullRequest: &event_processor.PullRequest{
					HeadRef:           "feature1",
					HeadSha:           "123",
					Title:             "fix",
					ChangeNumber:      1,
					LastCommitMessage: "123",
				},
			},
		},
		{
			name: "repository path empty",
			args: args{
				body: event_processor.BitbucketCommentEvent{},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "repository path empty")
			},
		},
		{
			name: "failed to get codebase",
			args: args{
				body: event_processor.BitbucketCommentEvent{
					BitbucketEvent: event_processor.BitbucketEvent{
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
