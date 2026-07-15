package interceptor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/event_processor"
	"github.com/epam/edp-tekton/pkg/event_processor/mocks"
	"github.com/epam/edp-tekton/pkg/reporter/provider/types"
)

type fakeStatusSetter struct {
	err    error
	ref    types.CommitRef
	status types.CommitStatus
	calls  int
}

func (f *fakeStatusSetter) SetCommitStatus(_ context.Context, ref types.CommitRef, status types.CommitStatus) error {
	f.calls++
	f.ref = ref
	f.status = status

	return f.err
}

func TestQueuedStatusReportingEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params map[string]any
		want   bool
	}{
		{name: "bool true", params: map[string]any{queuedStatusReportingParam: true}, want: true},
		{name: "bool false", params: map[string]any{queuedStatusReportingParam: false}, want: false},
		{name: "non-bool value", params: map[string]any{queuedStatusReportingParam: "true"}, want: false},
		{name: "missing", params: map[string]any{}, want: false},
		{name: "nil params", params: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, queuedStatusReportingEnabled(tt.params))
		})
	}
}

func queuedStatusFixtures(gitProvider string) []ctrlClient.Object {
	return []ctrlClient.Object{
		&codebaseApi.Codebase{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "demo"},
			Spec:       codebaseApi.CodebaseSpec{GitServer: "my-git"},
		},
		&codebaseApi.GitServer{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "my-git"},
			Spec: codebaseApi.GitServerSpec{
				GitProvider:      gitProvider,
				GitHost:          "git.example.com",
				NameSshKeySecret: "ci-secret",
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "ci-secret"},
			Data:       map[string][]byte{"token": []byte("api-token")},
		},
	}
}

func newQueuedStatusEvent() *event_processor.EventInfo {
	return &event_processor.EventInfo{
		GitProvider:  event_processor.GitProviderGitLab,
		RepoPath:     "/group/repo",
		TargetBranch: "master",
		Codebase: &codebaseApi.Codebase{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "demo"},
			Spec:       codebaseApi.CodebaseSpec{GitServer: "my-git"},
		},
		PullRequest: &event_processor.PullRequest{
			HeadSha:      "abc123",
			ChangeNumber: 1,
			Url:          "https://git.example.com/group/repo/-/merge_requests/1",
		},
	}
}

func TestEDPInterceptor_PostQueuedCommitStatus(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	tests := []struct {
		name          string
		gitProvider   string
		factoryErr    error
		setterErr     error
		wantErr       string
		wantCalls     int
		wantFactoryOK bool
	}{
		{
			name:          "posts queued status",
			gitProvider:   codebaseApi.GitProviderGitlab,
			wantCalls:     1,
			wantFactoryOK: true,
		},
		{
			name:        "gerrit is skipped",
			gitProvider: codebaseApi.GitProviderGerrit,
			wantCalls:   0,
		},
		{
			name:          "factory error is returned",
			gitProvider:   codebaseApi.GitProviderGitlab,
			factoryErr:    errors.New("bad provider"),
			wantErr:       "failed to create commit status setter",
			wantFactoryOK: true,
		},
		{
			name:          "setter error is returned",
			gitProvider:   codebaseApi.GitProviderGitlab,
			setterErr:     errors.New("api down"),
			wantErr:       "failed to set queued commit status",
			wantCalls:     1,
			wantFactoryOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(queuedStatusFixtures(tt.gitProvider)...).
				Build()

			setter := &fakeStatusSetter{err: tt.setterErr}

			var gotProvider, gotHost, gotToken string

			i := &EDPInterceptor{
				client: fakeClient,
				logger: zap.NewNop().Sugar(),
				statusSetterFactory: func(gitProvider, host, token string) (types.CommitStatusSetter, error) {
					gotProvider, gotHost, gotToken = gitProvider, host, token

					return setter, tt.factoryErr
				},
			}

			err := i.postQueuedCommitStatus(context.Background(), zap.NewNop().Sugar(), "default", newQueuedStatusEvent())

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantCalls, setter.calls)

			if tt.wantFactoryOK {
				assert.Equal(t, tt.gitProvider, gotProvider)
				assert.Equal(t, "git.example.com", gotHost)
				assert.Equal(t, "api-token", gotToken)
			}

			if tt.wantCalls > 0 {
				assert.Equal(t, types.CommitRef{RepoFullName: "group/repo", Sha: "abc123"}, setter.ref)
				assert.Equal(t, types.CommitStatus{
					State:       types.CommitStatePending,
					Context:     "Review Pipeline",
					Key:         "review",
					Name:        "Pipeline (QUEUED)",
					Description: "QUEUED",
					TargetURL:   "https://git.example.com/group/repo/-/merge_requests/1",
				}, setter.status)
			}
		})
	}
}

func TestEDPInterceptor_PostQueuedCommitStatus_ResolveError(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	// No fixtures: the GitServer lookup fails.
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	i := &EDPInterceptor{
		client: fakeClient,
		logger: zap.NewNop().Sugar(),
		statusSetterFactory: func(string, string, string) (types.CommitStatusSetter, error) {
			t.Fatal("factory must not be called when git server resolution fails")

			return nil, nil
		},
	}

	err := i.postQueuedCommitStatus(context.Background(), zap.NewNop().Sugar(), "default", newQueuedStatusEvent())
	assert.ErrorContains(t, err, "failed to resolve git server")
}

func TestEDPInterceptor_PostQueuedCommitStatus_MissingPullRequestData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(e *event_processor.EventInfo)
	}{
		{
			name:   "nil pull request",
			mutate: func(e *event_processor.EventInfo) { e.PullRequest = nil },
		},
		{
			name:   "empty head sha",
			mutate: func(e *event_processor.EventInfo) { e.PullRequest.HeadSha = "" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// No fixtures and a failing factory: the guard must return before
			// any lookup or setter construction happens.
			i := &EDPInterceptor{
				logger: zap.NewNop().Sugar(),
				statusSetterFactory: func(string, string, string) (types.CommitStatusSetter, error) {
					t.Fatal("factory must not be called without pull request data")

					return nil, nil
				},
			}

			event := newQueuedStatusEvent()
			tt.mutate(event)

			assert.NoError(t, i.postQueuedCommitStatus(context.Background(), zap.NewNop().Sugar(), "default", event))
		})
	}
}

func TestEDPInterceptor_PostQueuedCommitStatus_PortalTargetURL(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(queuedStatusFixtures(codebaseApi.GitProviderGitlab)...).
		Build()

	setter := &fakeStatusSetter{}

	i := &EDPInterceptor{
		client: fakeClient,
		logger: zap.NewNop().Sugar(),
		statusSetterFactory: func(string, string, string) (types.CommitStatusSetter, error) {
			return setter, nil
		},
		portalBaseURL: "https://portal.example.com/c/cluster/cicd/pipelineruns",
	}

	require.NoError(t,
		i.postQueuedCommitStatus(context.Background(), zap.NewNop().Sugar(), "default", newQueuedStatusEvent()))
	assert.Equal(t, "https://portal.example.com/c/cluster/cicd/pipelineruns", setter.status.TargetURL)
}

func TestEDPInterceptor_Process_QueuedStatusReporting(t *testing.T) {
	t.Parallel()

	newGitLabProcessor := func() event_processor.Processor {
		m := &mocks.Processor{}
		m.On("Process", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(newQueuedStatusEvent(), nil)

		return m
	}

	codebaseBranch := &codebaseApi.CodebaseBranch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "demo-master",
			Labels: map[string]string{
				codebaseApi.CodebaseLabel: "demo",
			},
		},
		Spec: codebaseApi.CodebaseBranchSpec{BranchName: "master"},
	}

	request := &triggersv1.InterceptorRequest{
		Header: map[string][]string{
			"X-Gitlab-Event": {"Merge Request Hook"},
		},
		Context: &triggersv1.TriggerContext{
			TriggerID: "namespace/default/triggers/name",
		},
		InterceptorParams: map[string]any{
			queuedStatusReportingParam: true,
		},
	}

	tests := []struct {
		name      string
		setterErr error
	}{
		{name: "status is posted"},
		{name: "status failure does not block triggering", setterErr: errors.New("api down")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			scheme := runtime.NewScheme()
			require.NoError(t, codebaseApi.AddToScheme(scheme))
			require.NoError(t, corev1.AddToScheme(scheme))

			fixtures := append(
				queuedStatusFixtures(codebaseApi.GitProviderGitlab),
				codebaseBranch.DeepCopy(),
			)

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(fixtures...).
				Build()

			setter := &fakeStatusSetter{err: tt.setterErr}

			i := NewEDPInterceptor(
				fakeClient,
				&mocks.Processor{},
				newGitLabProcessor(),
				&mocks.Processor{},
				&mocks.Processor{},
				zap.NewNop().Sugar(),
			)
			i.statusSetterFactory = func(string, string, string) (types.CommitStatusSetter, error) {
				return setter, nil
			}

			resp := i.Process(context.Background(), request)

			assert.True(t, resp.Continue)
			assert.Equal(t, 1, setter.calls)
		})
	}
}
