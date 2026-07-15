package interceptor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/event_processor"
	"github.com/epam/edp-tekton/pkg/reporter/gitserver"
	"github.com/epam/edp-tekton/pkg/reporter/provider/types"
)

const (
	// queuedStatusReportingParam is the interceptor parameter that enables
	// posting a queued commit status at webhook time; see postQueuedCommitStatus.
	queuedStatusReportingParam = "queuedStatusReporting"

	// reviewPipelineStatusContext must match the CONTEXT the review pipelines'
	// own status tasks use, so QUEUED, IN PROGRESS and PASSED/FAILED all
	// transition one check instead of creating duplicate contexts.
	reviewPipelineStatusContext = "Review Pipeline"

	// bitbucketReviewStatusKey must match the KEY of the review pipelines'
	// bitbucket-set-status tasks: Bitbucket build statuses with the same key
	// overwrite each other.
	bitbucketReviewStatusKey = "review"

	bitbucketQueuedStatusName = "Pipeline (QUEUED)"
	queuedStatusDescription   = "QUEUED"

	// queuedStatusPostTimeout bounds the whole best-effort status post,
	// including provider retries, so a slow git server cannot burn the
	// interceptor's own request budget and fail the trigger.
	queuedStatusPostTimeout = 2 * time.Second

	// portalBaseURLEnv carries the portal PipelineRun list URL (same value the
	// reporter consumes) used as the queued status target: the PipelineRun does
	// not exist yet, so no details page can be linked.
	portalBaseURLEnv = "PORTAL_BASE_URL"
)

// commitStatusSetterFactory builds a commit status setter for a git provider;
// it mirrors provider.NewCommitStatusSetter and exists for testability.
type commitStatusSetterFactory func(gitProvider, host, token string) (types.CommitStatusSetter, error)

// queuedStatusReportingEnabled checks if the queuedStatusReporting interceptor
// parameter is set to true.
func queuedStatusReportingEnabled(params map[string]any) bool {
	v, _ := params[queuedStatusReportingParam].(bool)

	return v
}

// postQueuedCommitStatus reports a pending/QUEUED commit status for the head
// SHA of the pull request, so the commit shows a check while its review
// PipelineRun waits in the queue. It is called before the PipelineRun is
// created and must stay best-effort: callers only log the returned error.
func (i *EDPInterceptor) postQueuedCommitStatus(
	ctx context.Context,
	log *zap.SugaredLogger,
	ns string,
	event *event_processor.EventInfo,
) error {
	if event.PullRequest == nil || event.PullRequest.HeadSha == "" {
		// Guarded by the caller as well; kept here so the dereferences below
		// never rely on caller discipline alone.
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, queuedStatusPostTimeout)
	defer cancel()

	// The event already carries the full Codebase, so resolve the GitServer directly.
	info, err := gitserver.ResolveGitServer(ctx, i.client, ns, event.Codebase.Spec.GitServer)
	if err != nil {
		return fmt.Errorf("failed to resolve git server for codebase %s: %w", event.Codebase.Name, err)
	}

	if info.Provider == codebaseApi.GitProviderGerrit {
		// Gerrit review reporting is vote-based; there is no commit status to seed.
		return nil
	}

	setter, err := i.statusSetterFactory(info.Provider, info.Host, info.Token)
	if err != nil {
		return fmt.Errorf("failed to create commit status setter: %w", err)
	}

	ref := types.CommitRef{
		RepoFullName: strings.TrimPrefix(event.RepoPath, "/"),
		Sha:          event.PullRequest.HeadSha,
	}
	// The PipelineRun does not exist yet, so link to the portal PipelineRun
	// list; the pipeline's own status tasks overwrite the URL with the run's
	// details page at admission. Fall back to the pull request when the portal
	// URL is not configured.
	targetURL := i.portalBaseURL
	if targetURL == "" {
		targetURL = event.PullRequest.Url
	}

	status := types.CommitStatus{
		State:       types.CommitStatePending,
		Context:     reviewPipelineStatusContext,
		Key:         bitbucketReviewStatusKey,
		Name:        bitbucketQueuedStatusName,
		Description: queuedStatusDescription,
		TargetURL:   targetURL,
	}

	if err := setter.SetCommitStatus(ctx, ref, status); err != nil {
		return fmt.Errorf("failed to set queued commit status for %s@%s: %w", ref.RepoFullName, ref.Sha, err)
	}

	log.Infof("Posted queued commit status for %s@%s", ref.RepoFullName, ref.Sha)

	return nil
}
