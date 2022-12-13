package interceptor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"github.com/tektoncd/triggers/pkg/interceptors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	codebaseApiV1 "github.com/epam/edp-codebase-operator/v2/pkg/apis/edp/v1"
)

const (
	executeTimeOut    = 3 * time.Second
	codebaseListLimit = 1000
)

// EDPInterceptorInterface is an interface for EDPInterceptor.
type EDPInterceptorInterface interface {
	Execute(r *http.Request) ([]byte, error)
}

// EDPInterceptor is an interceptor for EDP.
type EDPInterceptor struct {
	Client ctrlClient.Reader
	Logger *zap.SugaredLogger
}

// NewEDPInterceptor creates a new EDPInterceptor.
func NewEDPInterceptor(c ctrlClient.Reader, l *zap.SugaredLogger) *EDPInterceptor {
	return &EDPInterceptor{
		Client: c,
		Logger: l,
	}
}

// Execute executes the interceptor.
func (i *EDPInterceptor) Execute(r *http.Request) ([]byte, error) {
	ctx, cancel := context.WithTimeout(r.Context(), executeTimeOut)
	defer cancel()

	var body bytes.Buffer

	defer func() {
		if err := r.Body.Close(); err != nil {
			i.Logger.Errorf("Failed to close body: %s", err)
		}
	}()

	if _, err := io.Copy(&body, r.Body); err != nil {
		return nil, internal(fmt.Errorf("failed to read body: %w", err))
	}

	var ireq triggersv1.InterceptorRequest
	if err := json.Unmarshal(body.Bytes(), &ireq); err != nil {
		return nil, badRequest(fmt.Errorf("failed to parse body as InterceptorRequest: %w", err))
	}

	i.Logger.Infof("Interceptor request is: %s", body.Bytes())

	iresp := i.Process(ctx, &ireq)

	respBytes, err := json.Marshal(iresp)
	if err != nil {
		return nil, internal(err)
	}

	i.Logger.Infof("Interceptor response is: %s", respBytes)

	return respBytes, nil
}

// Process processes the interceptor request.
func (i *EDPInterceptor) Process(ctx context.Context, r *triggersv1.InterceptorRequest) *triggersv1.InterceptorResponse {
	ns, _ := triggersv1.ParseTriggerID(r.Context.TriggerID)

	event, err := getEventInfo(r)
	if err != nil {
		return interceptors.Fail(codes.InvalidArgument, err.Error())
	}

	codebase, err := i.getCodeBaseFromEventInfo(ctx, event, ns)
	if err != nil {
		return interceptors.Fail(codes.InvalidArgument, err.Error())
	}

	if codebase.Spec.Framework != nil {
		framework := strings.ToLower(*codebase.Spec.Framework)
		codebase.Spec.Framework = &framework
	}

	codebase.Spec.BuildTool = strings.ToLower(codebase.Spec.BuildTool)

	if codebase.Spec.CommitMessagePattern == nil {
		codebase.Spec.CommitMessagePattern = stringP("")
	}

	if codebase.Spec.JiraServer == nil {
		codebase.Spec.JiraServer = stringP("")
	}

	return &triggersv1.InterceptorResponse{
		Continue: true,
		Extensions: map[string]interface{}{
			"spec":           codebase.Spec,
			"codebase":       codebase.Name,
			"codebasebranch": fmt.Sprintf("%s-%s", codebase.Name, event.Branch),
		},
	}
}

// getCodeBaseFromEventInfo returns codebase from the git provider event.
// If the event is from gerrit, we search codebase by name what is equal to the gerrit project name.
// If the event is from GitHub/GitLab, we search codebase by gitUrlPath.
func (i *EDPInterceptor) getCodeBaseFromEventInfo(ctx context.Context, event *eventInfo, namespace string) (*codebaseApiV1.Codebase, error) {
	if event.GitProvider == gitProviderGerrit {
		codebase := &codebaseApiV1.Codebase{}
		if err := i.Client.Get(ctx, ctrlClient.ObjectKey{Namespace: namespace, Name: event.RepoPath}, codebase); err != nil {
			return nil, fmt.Errorf("failed to get codebase: %w", err)
		}

		return codebase, nil
	}

	codebase, err := i.getCodebaseByRepoPath(ctx, namespace, event.RepoPath)
	if err != nil {
		return nil, err
	}

	return codebase, nil
}

// getCodebaseByRepoPath returns codebase by repository path.
func (i *EDPInterceptor) getCodebaseByRepoPath(ctx context.Context, ns, repoPath string) (*codebaseApiV1.Codebase, error) {
	codebaseList := &codebaseApiV1.CodebaseList{}
	if err := i.Client.List(ctx, codebaseList, ctrlClient.InNamespace(ns), ctrlClient.Limit(codebaseListLimit)); err != nil {
		return nil, fmt.Errorf("unable to get codebase list: %w", err)
	}

	for n := range codebaseList.Items {
		if codebaseList.Items[n].Spec.GitUrlPath != nil && strings.EqualFold(*codebaseList.Items[n].Spec.GitUrlPath, repoPath) {
			return &codebaseList.Items[n], nil
		}
	}

	return nil, fmt.Errorf("codebase with repository path %s not found", repoPath)
}

// getEventInfo returns event info from interceptor request.
func getEventInfo(r *triggersv1.InterceptorRequest) (*eventInfo, error) {
	_, isGitHubEvent := r.Header["X-Github-Event"]
	_, isGitLabEvent := r.Header["X-Gitlab-Event"]

	if isGitLabEvent {
		gitLabEvent, err := unmarshalGitLabEvent([]byte(r.Body))
		if err != nil {
			return nil, err
		}

		return &eventInfo{
			GitProvider: gitProviderGitLab,
			RepoPath:    convertRepositoryPath(gitLabEvent.Project.PathWithNamespace),
			Branch:      convertBranchName(gitLabEvent.ObjectAttributes.TargetBranch),
		}, nil
	}

	if isGitHubEvent {
		gitHubEvent, err := unmarshalGitHubEvent([]byte(r.Body))
		if err != nil {
			return nil, err
		}

		return &eventInfo{
			GitProvider: gitProviderGitHub,
			RepoPath:    convertRepositoryPath(gitHubEvent.Repository.FullName),
			Branch:      convertBranchName(gitHubEvent.PullRequest.Base.Ref),
		}, nil
	}

	gerritEvent, err := unmarshalGerritEvent([]byte(r.Body))
	if err != nil {
		return nil, err
	}

	return &eventInfo{
		GitProvider: gitProviderGerrit,
		RepoPath:    strings.ToLower(gerritEvent.Project.Name),
		Branch:      convertBranchName(gerritEvent.Change.Branch),
	}, nil
}

// unmarshalGitLabEvent unmarshal GitLab event from json payload.
func unmarshalGitLabEvent(body []byte) (*GitLabEvent, error) {
	gitLabEvent := &GitLabEvent{}
	if err := json.Unmarshal(body, gitLabEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GitLab event: %w", err)
	}

	if gitLabEvent.Project.PathWithNamespace == "" {
		return nil, errors.New("gitlab repository path empty")
	}

	if gitLabEvent.ObjectAttributes.TargetBranch == "" {
		return nil, errors.New("gitlab target branch empty")
	}

	return gitLabEvent, nil
}

// unmarshalGitHubEvent unmarshal GitHub event from json payload.
func unmarshalGitHubEvent(body []byte) (*GitHubEvent, error) {
	gitHubEvent := &GitHubEvent{}
	if err := json.Unmarshal(body, gitHubEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GitHub event: %w", err)
	}

	if gitHubEvent.Repository.FullName == "" {
		return nil, errors.New("github repository path empty")
	}

	if gitHubEvent.PullRequest.Base.Ref == "" {
		return nil, errors.New("github target branch empty")
	}

	return gitHubEvent, nil
}

// unmarshalGerritEvent unmarshal Gerrit event from json payload.
func unmarshalGerritEvent(body []byte) (*GerritEvent, error) {
	gerritEventBody := &GerritEvent{}
	if err := json.Unmarshal(body, gerritEventBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gerrit event: %w", err)
	}

	if gerritEventBody.Project.Name == "" {
		return nil, errors.New("gerrit repository path empty")
	}

	if gerritEventBody.Change.Branch == "" {
		return nil, errors.New("gerrit target branch empty")
	}

	return gerritEventBody, nil
}

// stringPtr returns a pointer to the string value passed in.
func stringP(value string) *string {
	return &value
}

// convertRepositoryPath converts repository path to the format which is used in codebase.
func convertRepositoryPath(repo string) string {
	if !strings.HasPrefix(repo, "/") {
		repo = "/" + repo
	}

	return strings.ToLower(repo)
}

// convertBranchName converts the branch name to the format used in the codebase for Kubernetes resource naming.
func convertBranchName(branch string) string {
	r := strings.NewReplacer("/", "-")

	return r.Replace(branch)
}
