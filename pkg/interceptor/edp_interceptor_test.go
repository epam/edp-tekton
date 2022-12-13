package interceptor

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"github.com/tektoncd/triggers/pkg/interceptors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/pkg/apis/edp/v1"
)

func TestEDPInterceptor_Process(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(codebaseApi.AddToScheme(scheme))

	framework := "Java11"
	codebaseMeta := metav1.ObjectMeta{
		Namespace: "test-ns",
		Name:      "demo",
	}

	triggersContext := &triggersv1.TriggerContext{
		TriggerID: "namespace/test-ns/triggers/name",
	}

	tests := []struct {
		name        string
		objects     []runtime.Object
		request     *triggersv1.InterceptorRequest
		want        *triggersv1.InterceptorResponse
		containsErr string
	}{
		{
			name: "success gerrit payload",
			objects: []runtime.Object{
				&codebaseApi.Codebase{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-ns",
						Name:      "demo",
					},
					Spec: codebaseApi.CodebaseSpec{
						BuildTool: "Maven",
						Framework: &framework,
					},
				},
			},
			request: &triggersv1.InterceptorRequest{
				Body:    `{"project": {"name": "demo"}, "change": {"branch": "master"}}`,
				Context: triggersContext,
			},
			want: &triggersv1.InterceptorResponse{
				Extensions: map[string]interface{}{
					"spec": codebaseApi.CodebaseSpec{
						Framework:            stringP("java11"),
						BuildTool:            "maven",
						CommitMessagePattern: stringP(""),
						JiraServer:           stringP(""),
					},
					"codebase":       "demo",
					"codebasebranch": "demo-master",
				},
				Continue: true,
			},
		},
		{
			name: "success github payload",
			objects: []runtime.Object{
				&codebaseApi.Codebase{
					ObjectMeta: codebaseMeta,
					Spec: codebaseApi.CodebaseSpec{
						BuildTool:  "Maven",
						Framework:  &framework,
						GitUrlPath: stringP("/demo/Repo1"),
					},
				},
			},
			request: &triggersv1.InterceptorRequest{
				Body:    `{"repository": {"full_name": "demo/repo1"}, "pull_request": {"base": {"ref": "feature/1"}}}`,
				Header:  map[string][]string{"X-Github-Event": {"data"}},
				Context: triggersContext,
			},
			want: &triggersv1.InterceptorResponse{
				Extensions: map[string]interface{}{
					"spec": codebaseApi.CodebaseSpec{
						Framework:            stringP("java11"),
						BuildTool:            "maven",
						GitUrlPath:           stringP("/demo/Repo1"),
						CommitMessagePattern: stringP(""),
						JiraServer:           stringP(""),
					},
					"codebase":       "demo",
					"codebasebranch": "demo-feature-1",
				},
				Continue: true,
			},
		},
		{
			name: "success gitlab payload",
			objects: []runtime.Object{
				&codebaseApi.Codebase{
					ObjectMeta: codebaseMeta,
					Spec: codebaseApi.CodebaseSpec{
						BuildTool:            "Maven",
						Framework:            &framework,
						GitUrlPath:           stringP("/demo/repo2"),
						CommitMessagePattern: stringP("pattern"),
						JiraServer:           stringP("jira-server"),
					},
				},
			},
			request: &triggersv1.InterceptorRequest{
				Body:    `{"project": {"path_with_namespace": "demo/Repo2"}, "object_attributes": {"target_branch": "feature.1"}}`,
				Header:  map[string][]string{"X-Gitlab-Event": {"data"}},
				Context: triggersContext,
			},
			want: &triggersv1.InterceptorResponse{
				Extensions: map[string]interface{}{
					"spec": codebaseApi.CodebaseSpec{
						BuildTool:            "maven",
						Framework:            stringP("java11"),
						GitUrlPath:           stringP("/demo/repo2"),
						CommitMessagePattern: stringP("pattern"),
						JiraServer:           stringP("jira-server"),
					},
					"codebase":       "demo",
					"codebasebranch": "demo-feature.1",
				},
				Continue: true,
			},
		},
		{
			name: "success with empty framework",
			objects: []runtime.Object{
				&codebaseApi.Codebase{
					ObjectMeta: codebaseMeta,
					Spec: codebaseApi.CodebaseSpec{
						BuildTool:  "Maven",
						GitUrlPath: stringP("/demo/repo2"),
					},
				},
			},
			request: &triggersv1.InterceptorRequest{
				Body:    `{"project": {"path_with_namespace": "demo/repo2"}, "object_attributes": {"target_branch": "master"}}`,
				Header:  map[string][]string{"X-Gitlab-Event": {"data"}},
				Context: triggersContext,
			},
			want: &triggersv1.InterceptorResponse{
				Extensions: map[string]interface{}{
					"spec": codebaseApi.CodebaseSpec{
						BuildTool:            "maven",
						GitUrlPath:           stringP("/demo/repo2"),
						CommitMessagePattern: stringP(""),
						JiraServer:           stringP(""),
					},
					"codebase":       "demo",
					"codebasebranch": "demo-master",
				},
				Continue: true,
			},
		},
		{
			name: "failed to unmarshal gerrit payload",
			request: &triggersv1.InterceptorRequest{
				Body:    `{"repository": `,
				Context: triggersContext,
			},
			want:        interceptors.Failf(codes.InvalidArgument, "error"),
			containsErr: "failed to unmarshal Gerrit event",
		},
		{
			name: "no project name in gerrit payload",
			request: &triggersv1.InterceptorRequest{
				Body:    `{"project": {"field": "demo"}}`,
				Context: triggersContext,
			},
			want:        interceptors.Failf(codes.InvalidArgument, "error"),
			containsErr: "gerrit repository path empty",
		},
		{
			name: "no branch name in gerrit payload",
			request: &triggersv1.InterceptorRequest{
				Body:    `{"project": {"name": "demo"}}`,
				Context: triggersContext,
			},
			want:        interceptors.Failf(codes.InvalidArgument, "error"),
			containsErr: "gerrit target branch empty",
		},
		{
			name: "failed to unmarshal github payload",
			request: &triggersv1.InterceptorRequest{
				Body:    `{"repository": `,
				Header:  map[string][]string{"X-Github-Event": {"data"}},
				Context: triggersContext,
			},
			want:        interceptors.Failf(codes.InvalidArgument, "error"),
			containsErr: "failed to unmarshal GitHub event",
		},
		{
			name: "no repository name in github payload",
			request: &triggersv1.InterceptorRequest{
				Body:    `{"repository": {"field": "demo"}}`,
				Header:  map[string][]string{"X-Github-Event": {"data"}},
				Context: triggersContext,
			},
			want:        interceptors.Failf(codes.InvalidArgument, "error"),
			containsErr: "github repository path empty",
		},
		{
			name: "no branch name in github payload",
			request: &triggersv1.InterceptorRequest{
				Body:    `{"repository": {"full_name": "demo"}}`,
				Header:  map[string][]string{"X-Github-Event": {"data"}},
				Context: triggersContext,
			},
			want:        interceptors.Failf(codes.InvalidArgument, "error"),
			containsErr: "github target branch empty",
		},
		{
			name: "failed to unmarshal gitlab payload",
			request: &triggersv1.InterceptorRequest{
				Body:    `{"repository": `,
				Header:  map[string][]string{"X-Gitlab-Event": {"data"}},
				Context: triggersContext,
			},
			want:        interceptors.Failf(codes.InvalidArgument, "error"),
			containsErr: "failed to unmarshal GitLab event",
		},
		{
			name: "no repository name in gitlab payload",
			request: &triggersv1.InterceptorRequest{
				Body:    `{"repository": {"field": "demo"}}`,
				Header:  map[string][]string{"X-Gitlab-Event": {"data"}},
				Context: triggersContext,
			},
			want:        interceptors.Failf(codes.InvalidArgument, "error"),
			containsErr: "gitlab repository path empty",
		},
		{
			name: "no branch name in gitlab payload",
			request: &triggersv1.InterceptorRequest{
				Body:    `{"project": {"path_with_namespace": "demo"}}`,
				Header:  map[string][]string{"X-Gitlab-Event": {"data"}},
				Context: triggersContext,
			},
			want:        interceptors.Failf(codes.InvalidArgument, "error"),
			containsErr: "gitlab target branch empty",
		},
		{
			name: "codebase not found for gerrit flow",
			request: &triggersv1.InterceptorRequest{
				Body:    `{"project": {"name": "demo2"}, "change": {"branch": "master"}}`,
				Context: triggersContext,
			},
			want:        interceptors.Failf(codes.InvalidArgument, "error"),
			containsErr: "not found",
		},
		{
			name: "codebase not found in the list in gitlab flow",
			request: &triggersv1.InterceptorRequest{
				Body:    `{"project": {"path_with_namespace": "demo/Repo2"}, "object_attributes": {"target_branch": "master"}}`,
				Header:  map[string][]string{"X-Gitlab-Event": {"data"}},
				Context: triggersContext,
			},
			want:        interceptors.Failf(codes.InvalidArgument, "error"),
			containsErr: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(tt.objects...).Build()
			interceptor := NewEDPInterceptor(fakeClient, zap.NewNop().Sugar())

			got := interceptor.Process(context.Background(), tt.request)

			if tt.containsErr != "" {
				require.Contains(t, got.Status.Message, tt.containsErr)
			}

			// Disable checking equality of status message, equality of status code is enough.
			got.Status.Message = ""
			tt.want.Status.Message = ""

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEDPInterceptor_Execute(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(codebaseApi.AddToScheme(scheme))

	framework := "Java11"
	frameworkTransformed := "java11"
	codebase := &codebaseApi.Codebase{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
			Name:      "demo",
		},
		Spec: codebaseApi.CodebaseSpec{
			Framework:  &framework,
			BuildTool:  "Maven",
			GitUrlPath: stringP("/demo"),
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(codebase).Build()
	interceptor := NewEDPInterceptor(fakeClient, zap.NewNop().Sugar())

	tests := []struct {
		name     string
		reqBody  string
		wantResp *triggersv1.InterceptorResponse
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name:    "success",
			reqBody: `{"body": "{\"project\": {\"name\": \"demo\"}, \"change\": {\"branch\": \"feature1\"}}", "context": {"trigger_id": "namespace/test-ns/triggers/name"}}`,
			wantResp: &triggersv1.InterceptorResponse{
				Extensions: map[string]interface{}{
					"spec": codebaseApi.CodebaseSpec{
						Framework:            &frameworkTransformed,
						BuildTool:            "maven",
						GitUrlPath:           stringP("/demo"),
						CommitMessagePattern: stringP(""),
						JiraServer:           stringP(""),
					},
					"codebase":       "demo",
					"codebasebranch": "demo-feature1",
				},
				Continue: true,
			},
			wantErr: assert.NoError,
		},
		{
			name:    "failed to parse body",
			reqBody: `{"body": invalid data`,
			wantErr: assert.Error,
		},
		{
			name:    "failed to get codebase",
			reqBody: `{"body": "{\"project\": {\"name\": \"demo2\"}, \"change\": {\"branch\": \"feature1\"}}", "context": {"trigger_id": "namespace/test-ns/triggers/name"}}`,
			wantResp: &triggersv1.InterceptorResponse{
				Continue: false,
				Status: triggersv1.Status{
					Code:    codes.InvalidArgument,
					Message: "failed to get codebase: codebases.v2.edp.epam.com \"demo2\" not found",
				},
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "https://www.tektoncd.com", strings.NewReader(tt.reqBody))
			require.NoError(t, err)

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
