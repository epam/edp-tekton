package interceptor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"github.com/tektoncd/triggers/pkg/interceptors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/event_processor"
)

const (
	executeTimeOut = 3 * time.Second
)

// EDPInterceptorInterface is an interface for EDPInterceptor.
type EDPInterceptorInterface interface {
	Execute(r *http.Request) ([]byte, error)
}

// EDPInterceptor is an interceptor for EDP.
type EDPInterceptor struct {
	gitHubProcessor    event_processor.Processor
	gitLabProcessor    event_processor.Processor
	gerritProcessor    event_processor.Processor
	bitbucketProcessor event_processor.Processor
	client             ctrlClient.Reader
	logger             *zap.SugaredLogger
}

// NewEDPInterceptor creates a new EDPInterceptor.
func NewEDPInterceptor(
	c ctrlClient.Reader,
	gitHubProcessor event_processor.Processor,
	gitLabProcessor event_processor.Processor,
	gerritProcessor event_processor.Processor,
	bitbucketProcessor event_processor.Processor,
	l *zap.SugaredLogger,
) *EDPInterceptor {
	return &EDPInterceptor{
		gitHubProcessor:    gitHubProcessor,
		gitLabProcessor:    gitLabProcessor,
		gerritProcessor:    gerritProcessor,
		bitbucketProcessor: bitbucketProcessor,
		client:             c,
		logger:             l,
	}
}

// Execute executes the interceptor.
func (i *EDPInterceptor) Execute(r *http.Request) ([]byte, error) {
	ctx, cancel := context.WithTimeout(r.Context(), executeTimeOut)
	defer cancel()

	var body bytes.Buffer

	defer func() {
		if err := r.Body.Close(); err != nil {
			i.logger.Errorf("Failed to close body: %s", err)
		}
	}()

	if _, err := io.Copy(&body, r.Body); err != nil {
		return nil, internal(fmt.Errorf("failed to read body: %w", err))
	}

	var ireq triggersv1.InterceptorRequest
	if err := json.Unmarshal(body.Bytes(), &ireq); err != nil {
		return nil, badRequest(fmt.Errorf("failed to parse body as InterceptorRequest: %w", err))
	}

	i.logger.Infof("Interceptor request is: %s", body.Bytes())

	iresp := i.Process(ctx, &ireq)

	respBytes, err := json.Marshal(iresp)
	if err != nil {
		return nil, internal(err)
	}

	i.logger.Infof("Interceptor response is: %s", respBytes)

	return respBytes, nil
}

// Process processes the interceptor request.
func (i *EDPInterceptor) Process(ctx context.Context, r *triggersv1.InterceptorRequest) *triggersv1.InterceptorResponse {
	event, err := i.processEvent(ctx, r)
	if err != nil {
		return interceptors.Fail(codes.InvalidArgument, err.Error())
	}

	if event.IsReviewCommentEvent() {
		if !event.HasPipelineRecheck {
			i.logger.Infof("Pipeline recheck comment is not found, skipping pipeline triggering")

			return &triggersv1.InterceptorResponse{
				Continue: false,
			}
		}

		i.logger.Infof("Found comment for recheck, triggering pipeline")
	}

	prepareCodebase(event.Codebase)

	codebaseBranchName := convertBranchToCadebaseBranchName(event.TargetBranch, event.Codebase.Name)
	trigger := true
	ns, _ := triggersv1.ParseTriggerID(r.Context.TriggerID)

	codebaseBranch := codebaseApi.CodebaseBranch{}
	if err = i.client.Get(ctx, ctrlClient.ObjectKey{Namespace: ns, Name: codebaseBranchName}, &codebaseBranch); err != nil {
		if !k8serrors.IsNotFound(err) {
			return interceptors.Fail(codes.Internal, err.Error())
		}

		trigger = false

		i.logger.Infof("Codebasebranch with the name %s is not found, skipping pipeline triggering. "+
			"You can ignore this message otherwise add branch %s to codebase %s for the pipeline triggering",
			codebaseBranchName,
			event.TargetBranch,
			event.Codebase.Name,
		)
	}

	return &triggersv1.InterceptorResponse{
		Continue: trigger,
		Extensions: map[string]interface{}{
			"spec":           event.Codebase.Spec,
			"codebase":       event.Codebase.Name,
			"codebasebranch": codebaseBranchName,
			"targetBranch":   event.TargetBranch,
			"pullRequest":    event.PullRequest,
			"pipelines":      codebaseBranch.Spec.Pipelines,
		},
	}
}

// processEvent returns event info from interceptor request.
func (i *EDPInterceptor) processEvent(ctx context.Context, r *triggersv1.InterceptorRequest) (*event_processor.EventInfo, error) {
	githubEventType, isGitHubEvent := r.Header["X-Github-Event"]
	gitLabEventType, isGitLabEvent := r.Header["X-Gitlab-Event"]
	bitbucketEventType, isBitbucketEvent := r.Header["X-Event-Key"]
	ns, _ := triggersv1.ParseTriggerID(r.Context.TriggerID)

	if isGitLabEvent {
		event, err := i.gitLabProcessor.Process(ctx, []byte(r.Body), ns, getEventTypeFromHeader(gitLabEventType))
		if err != nil {
			return nil, fmt.Errorf("failed to process GitLab event: %w", err)
		}

		return event, nil
	}

	if isGitHubEvent {
		event, err := i.gitHubProcessor.Process(ctx, []byte(r.Body), ns, getEventTypeFromHeader(githubEventType))
		if err != nil {
			return nil, fmt.Errorf("failed to process GitHub event: %w", err)
		}

		return event, nil
	}

	if isBitbucketEvent {
		event, err := i.bitbucketProcessor.Process(ctx, []byte(r.Body), ns, getEventTypeFromHeader(bitbucketEventType))
		if err != nil {
			return nil, fmt.Errorf("failed to process Bitbucket event: %w", err)
		}

		return event, nil
	}

	event, err := i.gerritProcessor.Process(ctx, []byte(r.Body), ns, "")
	if err != nil {
		return nil, fmt.Errorf("failed to process Gerrit event: %w", err)
	}

	return event, nil
}

// getEventTypeFromHeader returns event type from header.
func getEventTypeFromHeader(headerData []string) string {
	if len(headerData) == 0 {
		return ""
	}

	return headerData[0]
}

// prepareCodebase prepares codebase for interceptor response.
func prepareCodebase(codebase *codebaseApi.Codebase) {
	codebase.Spec.Framework = strings.ToLower(codebase.Spec.Framework)

	codebase.Spec.BuildTool = strings.ToLower(codebase.Spec.BuildTool)

	if codebase.Spec.CommitMessagePattern == nil {
		codebase.Spec.CommitMessagePattern = ptr.To("")
	}

	if codebase.Spec.JiraServer == nil {
		codebase.Spec.JiraServer = ptr.To("")
	}
}

// convertBranchToCadebaseBranchName converts branch name to CodebaseBranch CR name.
func convertBranchToCadebaseBranchName(branch, codebaseName string) string {
	r := strings.NewReplacer("/", "-")

	return fmt.Sprintf("%s-%s", codebaseName, r.Replace(branch))
}
