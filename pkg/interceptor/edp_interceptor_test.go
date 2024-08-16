package interceptor

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v31/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/event_processor"
	"github.com/epam/edp-tekton/pkg/event_processor/mocks"
)

func TestEDPInterceptor_Execute(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))

	codebaseBranch := &codebaseApi.CodebaseBranch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "demo-master",
		},
		Spec: codebaseApi.CodebaseBranchSpec{
			Pipelines: map[string]string{
				"review": "review-pipeline",
				"build":  "build-pipeline",
			},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(codebaseBranch).Build()

	tests := []struct {
		name            string
		reqBody         string
		gerritProcessor func(t *testing.T) event_processor.Processor
		wantResp        *triggersv1.InterceptorResponse
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name:    "success",
			reqBody: `{"body": "{\"project\": {\"name\": \"demo\"}, \"change\": {\"branch\": \"feature1\"}}", "context": {"trigger_id": "namespace/default/triggers/name"}}`,
			gerritProcessor: func(t *testing.T) event_processor.Processor {
				m := &mocks.Processor{}
				m.On("Process", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&event_processor.EventInfo{
					GitProvider:  event_processor.GitProviderGerrit,
					RepoPath:     "/o/r",
					TargetBranch: "master",
					Type:         event_processor.EventTypeMerge,
					Codebase: &codebaseApi.Codebase{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "demo",
						},
						Spec: codebaseApi.CodebaseSpec{
							BuildTool:  "maven",
							Framework:  "spring",
							GitUrlPath: "/o/r",
						},
					},
				}, nil)

				return m
			},
			wantResp: &triggersv1.InterceptorResponse{
				Extensions: map[string]interface{}{
					"spec": codebaseApi.CodebaseSpec{
						Framework:            "spring",
						BuildTool:            "maven",
						GitUrlPath:           "/o/r",
						CommitMessagePattern: ptr.To(""),
						JiraServer:           ptr.To(""),
					},
					"codebase":       "demo",
					"codebasebranch": "demo-master",
					"targetBranch":   "master",
					"pullRequest":    nil,
					"pipelines": map[string]string{
						"review": "review-pipeline",
						"build":  "build-pipeline",
					},
				},
				Continue: true,
			},
			wantErr: assert.NoError,
		},
		{
			name:    "failed to parse body",
			reqBody: `{"body": invalid data`,
			gerritProcessor: func(t *testing.T) event_processor.Processor {
				return &mocks.Processor{}
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodPost, "https://www.tektoncd.com", strings.NewReader(tt.reqBody))
			require.NoError(t, err)

			interceptor := NewEDPInterceptor(
				fakeClient,
				mocks.NewProcessor(t),
				mocks.NewProcessor(t),
				tt.gerritProcessor(t),
				zap.NewNop().Sugar(),
			)

			got, err := interceptor.Execute(req)
			if !tt.wantErr(t, err) {
				return
			}

			if tt.wantResp != nil {
				want, err := json.Marshal(tt.wantResp)
				require.NoError(t, err)

				assert.JSONEq(t, string(want), string(got))
			}
		})
	}
}

func TestEDPInterceptor_Process(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))

	type fields struct {
		gitHubProcessor func(t *testing.T) event_processor.Processor
		gitLabProcessor func(t *testing.T) event_processor.Processor
		gerritProcessor func(t *testing.T) event_processor.Processor
		kubeObjects     []client.Object
	}

	type args struct {
		r       *triggersv1.InterceptorRequest
		reqBody any
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *triggersv1.InterceptorResponse
	}{
		{
			name: "github",
			fields: fields{
				gitHubProcessor: func(t *testing.T) event_processor.Processor {
					m := &mocks.Processor{}
					m.On("Process", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&event_processor.EventInfo{
						GitProvider:  event_processor.GitProviderGitHub,
						RepoPath:     "/o/r",
						TargetBranch: "master",
						Type:         event_processor.EventTypeMerge,
						Codebase: &codebaseApi.Codebase{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "default",
								Name:      "demo",
							},
							Spec: codebaseApi.CodebaseSpec{
								Framework:            "spring",
								BuildTool:            "maven",
								GitUrlPath:           "/o/r",
								CommitMessagePattern: ptr.To(""),
								JiraServer:           ptr.To(""),
							},
						},
						PullRequest: &event_processor.PullRequest{
							HeadRef:           "feature",
							HeadSha:           "sha",
							Title:             "fix",
							ChangeNumber:      1,
							LastCommitMessage: "commit message",
						},
					}, nil)

					return m
				},
				gitLabProcessor: func(t *testing.T) event_processor.Processor {
					return &mocks.Processor{}
				},
				gerritProcessor: func(t *testing.T) event_processor.Processor {
					return &mocks.Processor{}
				},
				kubeObjects: []client.Object{
					&codebaseApi.CodebaseBranch{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "demo-master",
						},
						Spec: codebaseApi.CodebaseBranchSpec{
							Pipelines: map[string]string{
								"review": "review-pipeline",
								"build":  "build-pipeline",
							},
						},
					},
				},
			},
			args: args{
				r: &triggersv1.InterceptorRequest{
					Header: map[string][]string{
						"X-Github-Event": {event_processor.GitHubEventTypeCommentAdded},
					},
					Context: &triggersv1.TriggerContext{
						TriggerID: "namespace/default/triggers/name",
					},
				},
				reqBody: github.PullRequestEvent{
					Repo: &github.Repository{
						FullName: ptr.To("o/r"),
					},
					PullRequest: &github.PullRequest{
						Base: &github.PullRequestBranch{
							Ref: ptr.To("master"),
						},
					},
				},
			},
			want: &triggersv1.InterceptorResponse{
				Extensions: map[string]interface{}{
					"spec": codebaseApi.CodebaseSpec{
						Framework:            "spring",
						BuildTool:            "maven",
						GitUrlPath:           "/o/r",
						CommitMessagePattern: ptr.To(""),
						JiraServer:           ptr.To(""),
					},
					"codebase":       "demo",
					"codebasebranch": "demo-master",
					"targetBranch":   "master",
					"pullRequest": &event_processor.PullRequest{
						HeadRef:           "feature",
						HeadSha:           "sha",
						Title:             "fix",
						ChangeNumber:      1,
						LastCommitMessage: "commit message",
					},
					"pipelines": map[string]string{
						"review": "review-pipeline",
						"build":  "build-pipeline",
					},
				},
				Continue: true,
			},
		},
		{
			name: "gitlab",
			fields: fields{
				gitHubProcessor: func(t *testing.T) event_processor.Processor {
					return &mocks.Processor{}
				},
				gitLabProcessor: func(t *testing.T) event_processor.Processor {
					m := &mocks.Processor{}

					m.On("Process", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&event_processor.EventInfo{
						GitProvider:  event_processor.GitProviderGitLab,
						RepoPath:     "/o/r",
						TargetBranch: "master",
						Type:         event_processor.EventTypeMerge,
						Codebase: &codebaseApi.Codebase{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "default",
								Name:      "demo",
							},
							Spec: codebaseApi.CodebaseSpec{
								Framework:            "spring",
								BuildTool:            "maven",
								GitUrlPath:           "/o/r",
								CommitMessagePattern: ptr.To(""),
								JiraServer:           ptr.To(""),
							},
						},
						PullRequest: &event_processor.PullRequest{
							HeadRef:      "feature",
							HeadSha:      "sha",
							Title:        "fix",
							ChangeNumber: 1,
						},
					}, nil)

					return m
				},
				gerritProcessor: func(t *testing.T) event_processor.Processor {
					return &mocks.Processor{}
				},
				kubeObjects: []client.Object{
					&codebaseApi.CodebaseBranch{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "demo-master",
						},
						Spec: codebaseApi.CodebaseBranchSpec{
							Pipelines: map[string]string{
								"review": "review-pipeline",
								"build":  "build-pipeline",
							},
						},
					},
				},
			},
			args: args{
				r: &triggersv1.InterceptorRequest{
					Header: map[string][]string{
						"X-Gitlab-Event": {event_processor.GitLabEventTypeCommentAdded},
					},
					Context: &triggersv1.TriggerContext{
						TriggerID: "namespace/default/triggers/name",
					},
				},
				reqBody: event_processor.GitLabMergeRequestsEvent{
					Project: event_processor.GitLabProject{
						PathWithNamespace: "o/r",
					},
					ObjectAttributes: event_processor.GitLabMergeRequest{
						TargetBranch: "master",
						Title:        "title",
						LastCommit: event_processor.GitLabCommit{
							ID: "123",
						},
					},
				},
			},
			want: &triggersv1.InterceptorResponse{
				Extensions: map[string]interface{}{
					"spec": codebaseApi.CodebaseSpec{
						Framework:            "spring",
						BuildTool:            "maven",
						GitUrlPath:           "/o/r",
						CommitMessagePattern: ptr.To(""),
						JiraServer:           ptr.To(""),
					},
					"codebase":       "demo",
					"codebasebranch": "demo-master",
					"targetBranch":   "master",
					"pullRequest": &event_processor.PullRequest{
						HeadRef:      "feature",
						HeadSha:      "sha",
						Title:        "fix",
						ChangeNumber: 1,
					},
					"pipelines": map[string]string{
						"review": "review-pipeline",
						"build":  "build-pipeline",
					},
				},
				Continue: true,
			},
		},
		{
			name: "gerrit",
			fields: fields{
				gitHubProcessor: func(t *testing.T) event_processor.Processor {
					return &mocks.Processor{}
				},
				gitLabProcessor: func(t *testing.T) event_processor.Processor {
					return &mocks.Processor{}
				},
				gerritProcessor: func(t *testing.T) event_processor.Processor {
					m := &mocks.Processor{}

					m.On("Process", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&event_processor.EventInfo{
						GitProvider:  event_processor.GitProviderGerrit,
						RepoPath:     "/o/r",
						TargetBranch: "master",
						Type:         event_processor.EventTypeMerge,
						Codebase: &codebaseApi.Codebase{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "default",
								Name:      "demo",
							},
							Spec: codebaseApi.CodebaseSpec{
								Framework:            "spring",
								BuildTool:            "maven",
								GitUrlPath:           "/o/r",
								CommitMessagePattern: ptr.To(""),
								JiraServer:           ptr.To(""),
							},
						},
					}, nil)

					return m
				},
				kubeObjects: []client.Object{
					&codebaseApi.CodebaseBranch{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "demo-master",
						},
						Spec: codebaseApi.CodebaseBranchSpec{
							Pipelines: map[string]string{
								"review": "review-pipeline",
								"build":  "build-pipeline",
							},
						},
					},
				},
			},
			args: args{
				r: &triggersv1.InterceptorRequest{
					Context: &triggersv1.TriggerContext{
						TriggerID: "namespace/default/triggers/name",
					},
				},
				reqBody: event_processor.GerritEvent{
					Project: struct {
						Name string `json:"name"`
					}{
						Name: "o/r",
					},
					Change: struct {
						Branch string `json:"branch"`
					}{
						Branch: "master",
					},
					Type: "patch-created",
				},
			},
			want: &triggersv1.InterceptorResponse{
				Extensions: map[string]interface{}{
					"spec": codebaseApi.CodebaseSpec{
						Framework:            "spring",
						BuildTool:            "maven",
						GitUrlPath:           "/o/r",
						CommitMessagePattern: ptr.To(""),
						JiraServer:           ptr.To(""),
					},
					"codebase":       "demo",
					"codebasebranch": "demo-master",
					"targetBranch":   "master",
					"pullRequest":    event_processor.EventInfo{}.PullRequest,
					"pipelines": map[string]string{
						"review": "review-pipeline",
						"build":  "build-pipeline",
					},
				},
				Continue: true,
			},
		},
		{
			name: "codebasebranch not found",
			fields: fields{
				gitHubProcessor: func(t *testing.T) event_processor.Processor {
					return &mocks.Processor{}
				},
				gitLabProcessor: func(t *testing.T) event_processor.Processor {
					return &mocks.Processor{}
				},
				gerritProcessor: func(t *testing.T) event_processor.Processor {
					m := &mocks.Processor{}

					m.On("Process", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&event_processor.EventInfo{
						GitProvider:  event_processor.GitProviderGerrit,
						RepoPath:     "/o/r",
						TargetBranch: "master",
						Type:         event_processor.EventTypeMerge,
						Codebase: &codebaseApi.Codebase{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "default",
								Name:      "demo",
							},
							Spec: codebaseApi.CodebaseSpec{
								Framework:            "spring",
								BuildTool:            "maven",
								GitUrlPath:           "/o/r",
								CommitMessagePattern: ptr.To(""),
								JiraServer:           ptr.To(""),
							},
						},
					}, nil)

					return m
				},
			},
			args: args{
				r: &triggersv1.InterceptorRequest{
					Context: &triggersv1.TriggerContext{
						TriggerID: "namespace/default/triggers/name",
					},
				},
				reqBody: event_processor.GerritEvent{
					Project: struct {
						Name string `json:"name"`
					}{
						Name: "o/r",
					},
					Change: struct {
						Branch string `json:"branch"`
					}{
						Branch: "master",
					},
					Type: "patch-created",
				},
			},
			want: &triggersv1.InterceptorResponse{
				Extensions: map[string]interface{}{
					"spec": codebaseApi.CodebaseSpec{
						Framework:            "spring",
						BuildTool:            "maven",
						GitUrlPath:           "/o/r",
						CommitMessagePattern: ptr.To(""),
						JiraServer:           ptr.To(""),
					},
					"codebase":       "demo",
					"codebasebranch": "demo-master",
					"targetBranch":   "master",
					"pullRequest":    event_processor.EventInfo{}.PullRequest,
					"pipelines":      codebaseApi.CodebaseBranchSpec{}.Pipelines,
				},
				Continue: false,
			},
		},
		{
			name: "comment event with no recheck",
			fields: fields{
				gitHubProcessor: func(t *testing.T) event_processor.Processor {
					return &mocks.Processor{}
				},
				gitLabProcessor: func(t *testing.T) event_processor.Processor {
					return &mocks.Processor{}
				},
				gerritProcessor: func(t *testing.T) event_processor.Processor {
					m := &mocks.Processor{}

					m.On("Process", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&event_processor.EventInfo{
						GitProvider:        event_processor.GitProviderGerrit,
						Type:               event_processor.EventTypeReviewComment,
						HasPipelineRecheck: false,
					}, nil)

					return m
				},
			},
			args: args{
				r: &triggersv1.InterceptorRequest{
					Context: &triggersv1.TriggerContext{
						TriggerID: "namespace/default/triggers/name",
					},
				},
				reqBody: event_processor.GerritEvent{
					Project: struct {
						Name string `json:"name"`
					}{
						Name: "o/r",
					},
					Change: struct {
						Branch string `json:"branch"`
					}{
						Branch: "master",
					},
					Type: "patch-created",
				},
			},
			want: &triggersv1.InterceptorResponse{
				Continue: false,
			},
		},
		{
			name: "failed to process event",
			fields: fields{
				gitHubProcessor: func(t *testing.T) event_processor.Processor {
					return &mocks.Processor{}
				},
				gitLabProcessor: func(t *testing.T) event_processor.Processor {
					return &mocks.Processor{}
				},
				gerritProcessor: func(t *testing.T) event_processor.Processor {
					m := &mocks.Processor{}

					m.On("Process", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
						nil,
						errors.New("failed to process event"),
					)

					return m
				},
			},
			args: args{
				r: &triggersv1.InterceptorRequest{
					Context: &triggersv1.TriggerContext{
						TriggerID: "namespace/default/triggers/name",
					},
				},
				reqBody: event_processor.GerritEvent{
					Project: struct {
						Name string `json:"name"`
					}{
						Name: "o/r",
					},
					Change: struct {
						Branch string `json:"branch"`
					}{
						Branch: "master",
					},
					Type: "patch-created",
				},
			},
			want: &triggersv1.InterceptorResponse{
				Continue: false,
				Status: triggersv1.Status{
					Code:    codes.InvalidArgument,
					Message: "failed to process Gerrit event: failed to process event",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tt.args.reqBody)
			require.NoError(t, err)

			tt.args.r.Body = string(body)

			i := NewEDPInterceptor(
				fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.fields.kubeObjects...).Build(),
				tt.fields.gitHubProcessor(t),
				tt.fields.gitLabProcessor(t),
				tt.fields.gerritProcessor(t),
				zap.NewNop().Sugar(),
			)

			got := i.Process(context.Background(), tt.args.r)
			assert.Equal(t, tt.want, got)
		})
	}
}
