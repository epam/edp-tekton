package interceptor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	tektonpipelineApi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	clientinterceptor "sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/event_processor"
	"github.com/epam/edp-tekton/pkg/event_processor/mocks"
)

func TestCancelInProgressEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params map[string]any
		want   bool
	}{
		{name: "bool true", params: map[string]any{cancelInProgressParam: true}, want: true},
		{name: "bool false", params: map[string]any{cancelInProgressParam: false}, want: false},
		{name: "non-bool value", params: map[string]any{cancelInProgressParam: "true"}, want: false},
		{name: "missing", params: map[string]any{}, want: false},
		{name: "nil params", params: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, cancelInProgressEnabled(tt.params))
		})
	}
}

func TestEDPInterceptor_CancelInProgressPipelineRuns(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, tektonpipelineApi.AddToScheme(scheme))

	newPipelineRun := func(
		name, codebase, changeNumber string,
		status duckv1.Status,
		specStatus tektonpipelineApi.PipelineRunSpecStatus,
	) *tektonpipelineApi.PipelineRun {
		return &tektonpipelineApi.PipelineRun{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      name,
				Labels: map[string]string{
					codebaseApi.CodebaseLabel: codebase,
					pipelineTypeLabel:         pipelineTypeReview,
					changeNumberLabel:         changeNumber,
				},
			},
			Spec: tektonpipelineApi.PipelineRunSpec{
				Status: specStatus,
			},
			Status: tektonpipelineApi.PipelineRunStatus{
				Status: status,
			},
		}
	}

	runningStatus := duckv1.Status{}
	succeededStatus := duckv1.Status{
		Conditions: duckv1.Conditions{
			{
				Type:   apis.ConditionSucceeded,
				Status: corev1.ConditionTrue,
			},
		},
	}

	running := newPipelineRun("running", "demo", "1", runningStatus, "")
	runningOtherChange := newPipelineRun("running-other-change", "demo", "2", runningStatus, "")
	runningOtherCodebase := newPipelineRun("running-other-codebase", "other", "1", runningStatus, "")
	succeeded := newPipelineRun("succeeded", "demo", "1", succeededStatus, "")
	// Seeded with statuses different from CancelledRunFinally so the assertions below
	// fail if the skip branches don't execute and the runs get patched.
	alreadyCancelled := newPipelineRun(
		"already-cancelled", "demo", "1", runningStatus, tektonpipelineApi.PipelineRunSpecStatusCancelled,
	)
	gracefullyStopped := newPipelineRun(
		"gracefully-stopped", "demo", "1", runningStatus, tektonpipelineApi.PipelineRunSpecStatusStoppedRunFinally,
	)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(running, runningOtherChange, runningOtherCodebase, succeeded, alreadyCancelled, gracefullyStopped).
		Build()

	interceptor := &EDPInterceptor{
		client: fakeClient,
		logger: zap.NewNop().Sugar(),
	}

	event := &event_processor.EventInfo{
		Codebase: &codebaseApi.Codebase{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "demo",
			},
		},
		PullRequest: &event_processor.PullRequest{
			ChangeNumber: 1,
		},
	}

	require.NoError(t, interceptor.cancelInProgressPipelineRuns(context.Background(), "default", event))

	wantSpecStatus := map[string]tektonpipelineApi.PipelineRunSpecStatus{
		"running":                tektonpipelineApi.PipelineRunSpecStatusCancelledRunFinally,
		"running-other-change":   "",
		"running-other-codebase": "",
		"succeeded":              "",
		"already-cancelled":      tektonpipelineApi.PipelineRunSpecStatusCancelled,
		"gracefully-stopped":     tektonpipelineApi.PipelineRunSpecStatusStoppedRunFinally,
	}

	pipelineRuns := &tektonpipelineApi.PipelineRunList{}
	require.NoError(t, fakeClient.List(context.Background(), pipelineRuns))

	for i := range pipelineRuns.Items {
		pipelineRun := &pipelineRuns.Items[i]
		assert.Equal(t, wantSpecStatus[pipelineRun.Name], pipelineRun.Spec.Status,
			"unexpected spec.status for %s", pipelineRun.Name)

		if pipelineRun.Name == "running" {
			assert.Equal(t, cancelReasonSuperseded, pipelineRun.Annotations[cancelReasonAnnotation],
				"cancelled run must carry the superseded cancel-reason annotation")
		} else {
			assert.NotContains(t, pipelineRun.Annotations, cancelReasonAnnotation,
				"untouched run %s must not carry the cancel-reason annotation", pipelineRun.Name)
		}
	}
}

func TestEDPInterceptor_CancelInProgressPipelineRuns_PatchError(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, tektonpipelineApi.AddToScheme(scheme))

	running := &tektonpipelineApi.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "running",
			Labels: map[string]string{
				codebaseApi.CodebaseLabel: "demo",
				pipelineTypeLabel:         pipelineTypeReview,
				changeNumberLabel:         "1",
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(running).
		WithInterceptorFuncs(clientinterceptor.Funcs{
			Patch: func(
				context.Context, ctrlClient.WithWatch, ctrlClient.Object, ctrlClient.Patch, ...ctrlClient.PatchOption,
			) error {
				return errors.New("patch failed")
			},
		}).
		Build()

	i := &EDPInterceptor{
		client: fakeClient,
		logger: zap.NewNop().Sugar(),
	}

	event := &event_processor.EventInfo{
		Codebase: &codebaseApi.Codebase{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "demo"},
		},
		PullRequest: &event_processor.PullRequest{ChangeNumber: 1},
	}

	require.NoError(t, i.cancelInProgressPipelineRuns(context.Background(), "default", event))

	got := &tektonpipelineApi.PipelineRun{}
	require.NoError(t,
		fakeClient.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "running"}, got))
	assert.Empty(t, got.Spec.Status)
}

func TestEDPInterceptor_Process_CancelInProgress(t *testing.T) {
	t.Parallel()

	newGitHubProcessor := func() event_processor.Processor {
		m := &mocks.Processor{}
		m.On("Process", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&event_processor.EventInfo{
			GitProvider:  event_processor.GitProviderGitHub,
			RepoPath:     "/o/r",
			TargetBranch: "master",
			Type:         event_processor.EventTypeMerge,
			Codebase: &codebaseApi.Codebase{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "demo"},
			},
			PullRequest: &event_processor.PullRequest{ChangeNumber: 1},
		}, nil)

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
		Spec: codebaseApi.CodebaseBranchSpec{
			BranchName: "master",
		},
	}

	running := &tektonpipelineApi.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "running",
			Labels: map[string]string{
				codebaseApi.CodebaseLabel: "demo",
				pipelineTypeLabel:         pipelineTypeReview,
				changeNumberLabel:         "1",
			},
		},
	}

	request := &triggersv1.InterceptorRequest{
		Header: map[string][]string{
			"X-Github-Event": {"pull_request"},
		},
		Context: &triggersv1.TriggerContext{
			TriggerID: "namespace/default/triggers/name",
		},
		InterceptorParams: map[string]any{
			cancelInProgressParam: true,
		},
	}

	t.Run("superseded run is cancelled", func(t *testing.T) {
		t.Parallel()

		scheme := runtime.NewScheme()
		require.NoError(t, codebaseApi.AddToScheme(scheme))
		require.NoError(t, tektonpipelineApi.AddToScheme(scheme))

		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(codebaseBranch.DeepCopy(), running.DeepCopy()).
			Build()

		i := NewEDPInterceptor(
			fakeClient,
			newGitHubProcessor(),
			&mocks.Processor{},
			&mocks.Processor{},
			&mocks.Processor{},
			zap.NewNop().Sugar(),
		)

		resp := i.Process(context.Background(), request)
		require.True(t, resp.Continue)

		got := &tektonpipelineApi.PipelineRun{}
		require.NoError(t,
			fakeClient.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "running"}, got))
		assert.EqualValues(t, tektonpipelineApi.PipelineRunSpecStatusCancelledRunFinally, got.Spec.Status)
	})

	t.Run("cancellation failure does not block triggering", func(t *testing.T) {
		t.Parallel()

		// The scheme misses the Tekton API, so listing PipelineRuns fails.
		scheme := runtime.NewScheme()
		require.NoError(t, codebaseApi.AddToScheme(scheme))

		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(codebaseBranch.DeepCopy()).
			Build()

		i := NewEDPInterceptor(
			fakeClient,
			newGitHubProcessor(),
			&mocks.Processor{},
			&mocks.Processor{},
			&mocks.Processor{},
			zap.NewNop().Sugar(),
		)

		resp := i.Process(context.Background(), request)
		assert.True(t, resp.Continue)
	})
}
