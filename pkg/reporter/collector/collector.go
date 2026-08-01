// Package collector gathers per-task and per-step results of a finished
// PipelineRun, including trailing log lines of failed steps.
package collector

import (
	"context"
	"fmt"
	"strings"
	"time"

	tektonpipelineApi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// cascadingSkipSuffix is the message Tekton writes for steps that never ran
// because an earlier step in the same task failed. Their logs carry no
// root-cause information and are not worth publishing.
const cascadingSkipSuffix = "Skipping step because a previous step failed"

type StepResult struct {
	Name      string
	Container string
	Succeeded bool
	ExitCode  int32
	// LogTail holds the trailing log lines of a failed step; empty for
	// successful steps and cascading skips.
	LogTail string
}

type TaskResult struct {
	Name      string
	Succeeded bool
	Duration  time.Duration
	StartTime time.Time
	Steps     []StepResult
}

type Report struct {
	PipelineRunName      string
	PipelineRunNamespace string
	Succeeded            bool
	Tasks                []TaskResult
}

// LogFetcher reads the trailing log lines of a container. It is an interface
// so tests can stub pod log streaming.
type LogFetcher interface {
	GetLogs(ctx context.Context, namespace, podName, container string, tailLines int64) (string, error)
}

// Options controls per-call Collect behavior. FetchLogs varies per
// PipelineRun (log publishing can be overridden by annotation); TailLines is
// the operator-configured tail size, kept separate from FetchLogs so a zero
// tail can never double as an off switch.
type Options struct {
	// FetchLogs controls whether failed steps get a log tail at all.
	FetchLogs bool
	// TailLines is the number of trailing lines per failed step.
	TailLines int64
}

type Collector struct {
	reader     ctrlClient.Reader
	logFetcher LogFetcher
}

// The reader must be an uncached reader: TaskRuns are read directly, never watched.
func New(reader ctrlClient.Reader, logFetcher LogFetcher) *Collector {
	return &Collector{
		reader:     reader,
		logFetcher: logFetcher,
	}
}

// Collect walks the PipelineRun child TaskRuns and returns the per-task,
// per-step outcome, with trailing logs for failed steps when opts.FetchLogs.
func (c *Collector) Collect(
	ctx context.Context,
	pipelineRun *tektonpipelineApi.PipelineRun,
	opts Options,
) (*Report, error) {
	report := &Report{
		PipelineRunName:      pipelineRun.Name,
		PipelineRunNamespace: pipelineRun.Namespace,
		Succeeded:            pipelineRun.Status.GetCondition(apis.ConditionSucceeded).IsTrue(),
	}

	for _, child := range pipelineRun.Status.ChildReferences {
		if child.Kind != "" && child.Kind != "TaskRun" {
			continue
		}

		taskRun := &tektonpipelineApi.TaskRun{}
		if err := c.reader.Get(
			ctx,
			types.NamespacedName{Namespace: pipelineRun.Namespace, Name: child.Name},
			taskRun,
		); err != nil {
			// TaskRuns can be pruned independently of the PipelineRun; report
			// what is still available instead of failing the whole run.
			if errors.IsNotFound(err) {
				continue
			}

			return nil, fmt.Errorf("failed to get TaskRun %s: %w", child.Name, err)
		}

		report.Tasks = append(report.Tasks, c.collectTask(ctx, child.PipelineTaskName, taskRun, opts))
	}

	return report, nil
}

func (c *Collector) collectTask(
	ctx context.Context,
	pipelineTaskName string,
	taskRun *tektonpipelineApi.TaskRun,
	opts Options,
) TaskResult {
	task := TaskResult{
		Name:      pipelineTaskName,
		Succeeded: taskRun.Status.GetCondition(apis.ConditionSucceeded).IsTrue(),
	}

	if taskRun.Status.StartTime != nil {
		task.StartTime = taskRun.Status.StartTime.Time

		if taskRun.Status.CompletionTime != nil {
			task.Duration = taskRun.Status.CompletionTime.Sub(taskRun.Status.StartTime.Time)
		}
	}

	for _, step := range taskRun.Status.Steps {
		stepResult := StepResult{
			Name:      step.Name,
			Container: step.Container,
			Succeeded: step.Terminated == nil || step.Terminated.ExitCode == 0,
		}

		if step.Terminated != nil {
			stepResult.ExitCode = step.Terminated.ExitCode
		}

		if !stepResult.Succeeded && opts.FetchLogs && opts.TailLines > 0 {
			stepResult.LogTail = c.fetchStepLog(ctx, taskRun, step.Container, opts.TailLines)
		}

		task.Steps = append(task.Steps, stepResult)
	}

	return task
}

// fetchStepLog returns the trailing log of a failed step, or empty when the
// log is unavailable (pod pruned) or the step is a cascading skip.
func (c *Collector) fetchStepLog(
	ctx context.Context,
	taskRun *tektonpipelineApi.TaskRun,
	container string,
	tailLines int64,
) string {
	if taskRun.Status.PodName == "" {
		return ""
	}

	logTail, err := c.logFetcher.GetLogs(ctx, taskRun.Namespace, taskRun.Status.PodName, container, tailLines)
	if err != nil {
		// The comment is still valuable without this snippet (e.g. the pod is
		// already garbage collected), so degrade gracefully.
		return ""
	}

	logTail = strings.TrimSpace(logTail)
	if strings.HasSuffix(logTail, cascadingSkipSuffix) {
		return ""
	}

	return logTail
}
