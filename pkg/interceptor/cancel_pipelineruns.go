package interceptor

import (
	"context"
	"fmt"
	"strconv"
	"time"

	tektonpipelineApi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"go.uber.org/zap"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/event_processor"
)

const (
	pipelineTypeLabel = "app.edp.epam.com/pipelinetype"
	changeNumberLabel = "app.edp.epam.com/git-change-number"

	pipelineTypeReview = "review"

	// cancelInProgressParam is the interceptor parameter that enables cancellation of
	// in-progress review PipelineRuns superseded by a new event for the same change.
	cancelInProgressParam = "cancelInProgress"

	// cancelReasonAnnotation mirrors the annotation the tekton-pipeline-queue operator
	// stamps on runs it cancels, so consumers (e.g. the finally status tasks) treat
	// interceptor- and queue-cancelled runs uniformly.
	cancelReasonAnnotation = "app.edp.epam.com/queue-cancel-reason"
	cancelReasonSuperseded = "superseded"

	// cancelInProgressTimeout bounds the whole best-effort cancellation so a slow
	// API server cannot burn the interceptor's request budget: the webhook budget
	// (executeTimeOut) is shared with postQueuedCommitStatus, which runs after
	// this call.
	cancelInProgressTimeout = time.Second
)

// cancelInProgressEnabled checks if the cancelInProgress interceptor parameter is set to true.
func cancelInProgressEnabled(params map[string]any) bool {
	v, _ := params[cancelInProgressParam].(bool)

	return v
}

// cancelInProgressPipelineRuns gracefully cancels review PipelineRuns that are still running
// for the same codebase and pull request. It is called before the new PipelineRun is created,
// so the run triggered by the current event is never affected.
func (i *EDPInterceptor) cancelInProgressPipelineRuns(
	ctx context.Context,
	log *zap.SugaredLogger,
	ns string,
	event *event_processor.EventInfo,
) error {
	ctx, cancel := context.WithTimeout(ctx, cancelInProgressTimeout)
	defer cancel()

	pipelineRuns := &tektonpipelineApi.PipelineRunList{}
	if err := i.client.List(
		ctx,
		pipelineRuns,
		ctrlClient.InNamespace(ns),
		ctrlClient.MatchingLabels{
			codebaseApi.CodebaseLabel: event.Codebase.Name,
			pipelineTypeLabel:         pipelineTypeReview,
			changeNumberLabel:         strconv.Itoa(event.PullRequest.ChangeNumber),
		},
	); err != nil {
		return fmt.Errorf("failed to list PipelineRuns for codebase %s change %d: %w",
			event.Codebase.Name, event.PullRequest.ChangeNumber, err)
	}

	for idx := range pipelineRuns.Items {
		pipelineRun := &pipelineRuns.Items[idx]

		if pipelineRun.IsDone() ||
			pipelineRun.IsCancelled() ||
			pipelineRun.IsGracefullyCancelled() ||
			pipelineRun.IsGracefullyStopped() {
			continue
		}

		patch := ctrlClient.MergeFrom(pipelineRun.DeepCopy())
		pipelineRun.Spec.Status = tektonpipelineApi.PipelineRunSpecStatusCancelledRunFinally

		if pipelineRun.Annotations == nil {
			pipelineRun.Annotations = map[string]string{}
		}

		pipelineRun.Annotations[cancelReasonAnnotation] = cancelReasonSuperseded

		if err := i.client.Patch(ctx, pipelineRun, patch); err != nil {
			log.Errorf("Failed to cancel PipelineRun %s superseded by a new event for codebase %s change %d: %s",
				pipelineRun.Name, event.Codebase.Name, event.PullRequest.ChangeNumber, err)

			continue
		}

		log.Infof("Canceled in-progress PipelineRun %s superseded by a new event for codebase %s change %d",
			pipelineRun.Name, event.Codebase.Name, event.PullRequest.ChangeNumber)
	}

	return nil
}
