package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tektonpipelineApi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/reporter"
	"github.com/epam/edp-tekton/pkg/reporter/collector"
	"github.com/epam/edp-tekton/pkg/reporter/formatter"
	providerTypes "github.com/epam/edp-tekton/pkg/reporter/provider/types"
)

type fakeProvider struct {
	comments []providerTypes.Comment
	refs     []providerTypes.PullRequestRef
	err      error
}

func (f *fakeProvider) UpsertComment(
	_ context.Context,
	ref providerTypes.PullRequestRef,
	comment providerTypes.Comment,
) error {
	if f.err != nil {
		return f.err
	}

	f.refs = append(f.refs, ref)
	f.comments = append(f.comments, comment)

	return nil
}

type stubLogFetcher struct{}

func (stubLogFetcher) GetLogs(_ context.Context, _, _, container string, _ int64) (string, error) {
	return "log tail of " + container + " secret=gh-token", nil
}

func newScheme(t *testing.T) *runtime.Scheme {
	t.Helper()

	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, codebaseApi.AddToScheme(scheme))
	require.NoError(t, tektonpipelineApi.AddToScheme(scheme))

	return scheme
}

func newPipelineRun() *tektonpipelineApi.PipelineRun {
	return &tektonpipelineApi.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "review-my-app-abc",
			Namespace: "krci",
			Labels: map[string]string{
				reporter.PipelineTypeLabel: reporter.PipelineTypeReview,
				codebaseApi.CodebaseLabel:  "my-app",
				reporter.ChangeNumberLabel: "7",
			},
			Annotations: map[string]string{
				reporter.ResultAnnotationsKey: `{
					"app.edp.epam.com/git-repository": "org/my-app",
					"app.edp.epam.com/git-change-number": "7"
				}`,
			},
		},
		Status: tektonpipelineApi.PipelineRunStatus{
			Status: duckv1.Status{
				Conditions: duckv1.Conditions{{Type: apis.ConditionSucceeded, Status: corev1.ConditionFalse}},
			},
			PipelineRunStatusFields: tektonpipelineApi.PipelineRunStatusFields{
				ChildReferences: []tektonpipelineApi.ChildStatusReference{
					{Name: "review-my-app-abc-build", PipelineTaskName: "build", TypeMeta: runtime.TypeMeta{Kind: "TaskRun"}},
				},
			},
		},
	}
}

func newTaskRun() *tektonpipelineApi.TaskRun {
	start := metav1.NewTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	end := metav1.NewTime(start.Add(30 * time.Second))

	return &tektonpipelineApi.TaskRun{
		ObjectMeta: metav1.ObjectMeta{Name: "review-my-app-abc-build", Namespace: "krci"},
		Status: tektonpipelineApi.TaskRunStatus{
			Status: duckv1.Status{
				Conditions: duckv1.Conditions{{Type: apis.ConditionSucceeded, Status: corev1.ConditionFalse}},
			},
			TaskRunStatusFields: tektonpipelineApi.TaskRunStatusFields{
				PodName:        "build-pod",
				StartTime:      &start,
				CompletionTime: &end,
				Steps: []tektonpipelineApi.StepState{
					{
						Name:      "npm-build",
						Container: "step-npm-build",
						ContainerState: corev1.ContainerState{
							Terminated: &corev1.ContainerStateTerminated{ExitCode: 1},
						},
					},
				},
			},
		},
	}
}

func gitObjects() []ctrlClient.Object {
	return []ctrlClient.Object{
		&codebaseApi.Codebase{
			ObjectMeta: metav1.ObjectMeta{Name: "my-app", Namespace: "krci"},
			Spec:       codebaseApi.CodebaseSpec{GitServer: "github"},
		},
		&codebaseApi.GitServer{
			ObjectMeta: metav1.ObjectMeta{Name: "github", Namespace: "krci"},
			Spec: codebaseApi.GitServerSpec{
				GitHost:          "github.com",
				GitProvider:      codebaseApi.GitProviderGithub,
				NameSshKeySecret: "ci-github",
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "ci-github", Namespace: "krci"},
			Data:       map[string][]byte{"token": []byte("gh-token")},
		},
	}
}

func newReconciler(
	t *testing.T,
	client ctrlClient.Client,
	gitProvider providerTypes.Provider,
	config *reporter.Config,
) *PipelineRunReconciler {
	t.Helper()

	if config == nil {
		config = &reporter.Config{TailLines: 100, CommentStrategy: reporter.CommentStrategyUpdate}
	}

	return NewPipelineRunReconciler(
		client,
		client,
		collector.New(client, stubLogFetcher{}, config.TailLines),
		formatter.New(formatter.PortalLinkBuilder{}),
		func(_, _, _ string) (providerTypes.Provider, error) {
			return gitProvider, nil
		},
		config,
	)
}

func reconcileRequest() ctrl.Request {
	return ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "krci", Name: "review-my-app-abc"}}
}

func TestReconcilePublishesReportAndMarksPipelineRun(t *testing.T) {
	t.Parallel()

	pipelineRun := newPipelineRun()
	objects := append(gitObjects(), pipelineRun, newTaskRun())

	client := fake.NewClientBuilder().WithScheme(newScheme(t)).WithObjects(objects...).Build()
	gitProvider := &fakeProvider{}

	r := newReconciler(t, client, gitProvider, nil)

	_, err := r.Reconcile(context.Background(), reconcileRequest())
	require.NoError(t, err)

	require.Len(t, gitProvider.comments, 1)
	assert.Equal(t, providerTypes.PullRequestRef{RepoFullName: "org/my-app", Number: 7}, gitProvider.refs[0])

	comment := gitProvider.comments[0]
	assert.True(t, comment.Update)
	assert.Contains(t, comment.Body, "<!-- krci-pipeline-report codebase=my-app -->")
	assert.Contains(t, comment.Body, "| ❌ | build | 30s |")
	assert.Contains(t, comment.Body, "log tail of step-npm-build")
	assert.Contains(t, comment.Body, "secret=*****", "git token must be masked in logs")
	assert.NotContains(t, comment.Body, "gh-token")

	updated := &tektonpipelineApi.PipelineRun{}
	require.NoError(t, client.Get(context.Background(), reconcileRequest().NamespacedName, updated))
	assert.Equal(t, "true", updated.Annotations[reporter.ReportedAnnotation])
}

func TestReconcileSkipsAlreadyReported(t *testing.T) {
	t.Parallel()

	pipelineRun := newPipelineRun()
	pipelineRun.Annotations[reporter.ReportedAnnotation] = "true"

	client := fake.NewClientBuilder().WithScheme(newScheme(t)).WithObjects(pipelineRun).Build()
	gitProvider := &fakeProvider{}

	r := newReconciler(t, client, gitProvider, nil)

	_, err := r.Reconcile(context.Background(), reconcileRequest())
	require.NoError(t, err)
	assert.Empty(t, gitProvider.comments)
}

func TestReconcileSkipsRunningAndNonReviewRuns(t *testing.T) {
	t.Parallel()

	running := newPipelineRun()
	running.Status.Conditions = duckv1.Conditions{{Type: apis.ConditionSucceeded, Status: corev1.ConditionUnknown}}

	build := newPipelineRun()
	build.Name = "build-run"
	build.Labels[reporter.PipelineTypeLabel] = "build"

	client := fake.NewClientBuilder().WithScheme(newScheme(t)).WithObjects(running, build).Build()
	gitProvider := &fakeProvider{}

	r := newReconciler(t, client, gitProvider, nil)

	_, err := r.Reconcile(context.Background(), reconcileRequest())
	require.NoError(t, err)

	_, err = r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "krci", Name: "build-run"},
	})
	require.NoError(t, err)

	assert.Empty(t, gitProvider.comments)
}

func TestReconcilePermanentErrorsDoNotRequeue(t *testing.T) {
	t.Parallel()

	t.Run("missing codebase resources are marked handled", func(t *testing.T) {
		t.Parallel()

		// Codebase/GitServer/secret are absent: the report can never succeed, so
		// the run must be marked handled to avoid re-erroring on every resync
		// (the same path an unsupported Gerrit review run takes).
		client := fake.NewClientBuilder().WithScheme(newScheme(t)).WithObjects(newPipelineRun()).Build()
		gitProvider := &fakeProvider{}

		r := newReconciler(t, client, gitProvider, nil)

		result, err := r.Reconcile(context.Background(), reconcileRequest())
		require.NoError(t, err)
		assert.Zero(t, result.RequeueAfter)
		assert.Empty(t, gitProvider.comments)

		current := &tektonpipelineApi.PipelineRun{}
		require.NoError(t, client.Get(context.Background(), reconcileRequest().NamespacedName, current))
		assert.Equal(t, reportSkipped, current.Annotations[reporter.ReportedAnnotation])
	})

	t.Run("missing repository metadata", func(t *testing.T) {
		t.Parallel()

		pipelineRun := newPipelineRun()
		pipelineRun.Annotations = nil

		objects := append(gitObjects(), pipelineRun, newTaskRun())
		client := fake.NewClientBuilder().WithScheme(newScheme(t)).WithObjects(objects...).Build()

		r := newReconciler(t, client, &fakeProvider{}, nil)

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		require.NoError(t, err)
	})
}

func TestReconcileProviderErrorRequeues(t *testing.T) {
	t.Parallel()

	objects := append(gitObjects(), newPipelineRun(), newTaskRun())
	client := fake.NewClientBuilder().WithScheme(newScheme(t)).WithObjects(objects...).Build()

	r := newReconciler(t, client, &fakeProvider{err: errors.New("api is down")}, nil)

	_, err := r.Reconcile(context.Background(), reconcileRequest())
	require.Error(t, err)

	// The annotation is not set, so the retry publishes the comment.
	current := &tektonpipelineApi.PipelineRun{}
	require.NoError(t, client.Get(context.Background(), reconcileRequest().NamespacedName, current))
	assert.NotContains(t, current.Annotations, reporter.ReportedAnnotation)
}

func TestReconcileNewCommentStrategy(t *testing.T) {
	t.Parallel()

	objects := append(gitObjects(), newPipelineRun(), newTaskRun())
	client := fake.NewClientBuilder().WithScheme(newScheme(t)).WithObjects(objects...).Build()
	gitProvider := &fakeProvider{}

	config := &reporter.Config{TailLines: 100, CommentStrategy: reporter.CommentStrategyNew}
	r := newReconciler(t, client, gitProvider, config)

	_, err := r.Reconcile(context.Background(), reconcileRequest())
	require.NoError(t, err)

	require.Len(t, gitProvider.comments, 1)
	assert.False(t, gitProvider.comments[0].Update)
}
