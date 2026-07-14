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

// StepResult describes the outcome of a single step within a task.
type StepResult struct {
	Name      string
	Container string
	Succeeded bool
	ExitCode  int32
	// LogTail holds the trailing log lines of a failed step; empty for
	// successful steps and cascading skips.
	LogTail string
}

// TaskResult describes the outcome of a single pipeline task.
type TaskResult struct {
	Name      string
	Succeeded bool
	Duration  time.Duration
	StartTime time.Time
	Steps     []StepResult
}

// Report is the aggregated result of a PipelineRun.
type Report struct {
	PipelineRunName      string
	PipelineRunNamespace string
	Succeeded            bool
	Tasks                []TaskResult
	// TaskRuns are the child TaskRuns backing Tasks, kept so callers can
	// inspect step specs (e.g. secret env references) without re-fetching.
	TaskRuns []*tektonpipelineApi.TaskRun
}

// LogFetcher reads the trailing log lines of a container. It is an interface
// so tests can stub pod log streaming.
type LogFetcher interface {
	GetLogs(ctx context.Context, namespace, podName, container string, tailLines int64) (string, error)
}

// Collector builds a Report from a finished PipelineRun.
type Collector struct {
	reader     ctrlClient.Reader
	logFetcher LogFetcher
	tailLines  int64
}

// New creates a Collector. The reader is used for direct (uncached) TaskRun reads.
func New(reader ctrlClient.Reader, logFetcher LogFetcher, tailLines int64) *Collector {
	return &Collector{
		reader:     reader,
		logFetcher: logFetcher,
		tailLines:  tailLines,
	}
}

// Collect walks the PipelineRun child TaskRuns and returns the per-task,
// per-step outcome with trailing logs for failed steps.
func (c *Collector) Collect(ctx context.Context, pipelineRun *tektonpipelineApi.PipelineRun) (*Report, error) {
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

		report.Tasks = append(report.Tasks, c.collectTask(ctx, child.PipelineTaskName, taskRun))
		report.TaskRuns = append(report.TaskRuns, taskRun)
	}

	return report, nil
}

func (c *Collector) collectTask(
	ctx context.Context,
	pipelineTaskName string,
	taskRun *tektonpipelineApi.TaskRun,
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

		if !stepResult.Succeeded {
			stepResult.LogTail = c.fetchStepLog(ctx, taskRun, step.Container)
		}

		task.Steps = append(task.Steps, stepResult)
	}

	return task
}

// fetchStepLog returns the trailing log of a failed step, or empty when the
// log is unavailable (pod pruned) or the step is a cascading skip.
func (c *Collector) fetchStepLog(ctx context.Context, taskRun *tektonpipelineApi.TaskRun, container string) string {
	if taskRun.Status.PodName == "" {
		return ""
	}

	logTail, err := c.logFetcher.GetLogs(ctx, taskRun.Namespace, taskRun.Status.PodName, container, c.tailLines)
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
