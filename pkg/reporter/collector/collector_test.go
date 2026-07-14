package collector

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
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type stubLogFetcher struct {
	logs map[string]string
	err  error
}

func (s *stubLogFetcher) GetLogs(_ context.Context, _, _, container string, _ int64) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	return s.logs[container], nil
}

func newScheme(t *testing.T) *runtime.Scheme {
	t.Helper()

	scheme := runtime.NewScheme()
	require.NoError(t, tektonpipelineApi.AddToScheme(scheme))

	return scheme
}

func succeededCondition(status corev1.ConditionStatus) duckv1.Conditions {
	return duckv1.Conditions{{Type: apis.ConditionSucceeded, Status: status}}
}

func newTaskRun(name string, succeeded bool, steps []tektonpipelineApi.StepState) *tektonpipelineApi.TaskRun {
	status := corev1.ConditionTrue
	if !succeeded {
		status = corev1.ConditionFalse
	}

	start := metav1.NewTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	end := metav1.NewTime(start.Add(90 * time.Second))

	return &tektonpipelineApi.TaskRun{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "krci"},
		Status: tektonpipelineApi.TaskRunStatus{
			Status: duckv1.Status{Conditions: succeededCondition(status)},
			TaskRunStatusFields: tektonpipelineApi.TaskRunStatusFields{
				PodName:        name + "-pod",
				StartTime:      &start,
				CompletionTime: &end,
				Steps:          steps,
			},
		},
	}
}

func newPipelineRun(children ...string) *tektonpipelineApi.PipelineRun {
	refs := make([]tektonpipelineApi.ChildStatusReference, 0, len(children))
	for _, name := range children {
		refs = append(refs, tektonpipelineApi.ChildStatusReference{
			Name:             name,
			PipelineTaskName: "task-" + name,
			TypeMeta:         runtime.TypeMeta{Kind: "TaskRun"},
		})
	}

	return &tektonpipelineApi.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{Name: "review-run", Namespace: "krci"},
		Status: tektonpipelineApi.PipelineRunStatus{
			Status: duckv1.Status{Conditions: succeededCondition(corev1.ConditionFalse)},
			PipelineRunStatusFields: tektonpipelineApi.PipelineRunStatusFields{
				ChildReferences: refs,
			},
		},
	}
}

func terminatedStep(name string, exitCode int32) tektonpipelineApi.StepState {
	return tektonpipelineApi.StepState{
		Name:      name,
		Container: "step-" + name,
		ContainerState: corev1.ContainerState{
			Terminated: &corev1.ContainerStateTerminated{ExitCode: exitCode},
		},
	}
}

func TestCollect(t *testing.T) {
	t.Parallel()

	failedTaskRun := newTaskRun("build", false, []tektonpipelineApi.StepState{
		terminatedStep("npm-build", 1),
		terminatedStep("cascading", 1),
		terminatedStep("lint", 0),
	})
	okTaskRun := newTaskRun("fetch", true, []tektonpipelineApi.StepState{
		terminatedStep("clone", 0),
	})

	reader := fake.NewClientBuilder().
		WithScheme(newScheme(t)).
		WithObjects(failedTaskRun, okTaskRun).
		Build()

	fetcher := &stubLogFetcher{logs: map[string]string{
		"step-npm-build": "npm ERR! Exit status 1\n",
		"step-cascading": "2026/01/01 error: Skipping step because a previous step failed",
	}}

	// The PipelineRun references a pruned TaskRun as well; it must be skipped.
	report, err := New(reader, fetcher, 100).Collect(context.Background(), newPipelineRun("build", "fetch", "pruned"))
	require.NoError(t, err)

	assert.False(t, report.Succeeded)
	require.Len(t, report.Tasks, 2)
	require.Len(t, report.TaskRuns, 2)

	build := report.Tasks[0]
	assert.Equal(t, "task-build", build.Name)
	assert.False(t, build.Succeeded)
	assert.Equal(t, 90*time.Second, build.Duration)
	require.Len(t, build.Steps, 3)

	assert.Equal(t, "npm ERR! Exit status 1", build.Steps[0].LogTail, "failed step keeps its trimmed log tail")
	assert.False(t, build.Steps[0].Succeeded)
	assert.Equal(t, int32(1), build.Steps[0].ExitCode)

	assert.Empty(t, build.Steps[1].LogTail, "cascading skip logs are dropped")
	assert.True(t, build.Steps[2].Succeeded)
	assert.Empty(t, build.Steps[2].LogTail, "successful steps fetch no logs")

	fetch := report.Tasks[1]
	assert.True(t, fetch.Succeeded)
}

func TestCollectLogFetchErrorDegradesGracefully(t *testing.T) {
	t.Parallel()

	failedTaskRun := newTaskRun("build", false, []tektonpipelineApi.StepState{terminatedStep("npm-build", 1)})

	reader := fake.NewClientBuilder().WithScheme(newScheme(t)).WithObjects(failedTaskRun).Build()

	report, err := New(reader, &stubLogFetcher{err: errors.New("pod is gone")}, 100).
		Collect(context.Background(), newPipelineRun("build"))
	require.NoError(t, err)

	require.Len(t, report.Tasks, 1)
	assert.Empty(t, report.Tasks[0].Steps[0].LogTail)
	assert.False(t, report.Tasks[0].Steps[0].Succeeded)
}
