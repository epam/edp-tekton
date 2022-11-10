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

	repoName, err := getRepoPath(r)
	if err != nil {
		return interceptors.Failf(codes.InvalidArgument, "failed to get repository name: %v", err)
	}

	codebase, err := i.getCodebaseByRepoPath(ctx, ns, repoName)
	if err != nil {
		return interceptors.Failf(codes.NotFound, "failed to get codebase: %v", err)
	}

	if codebase.Spec.Framework != nil {
		framework := strings.ToLower(*codebase.Spec.Framework)
		codebase.Spec.Framework = &framework
	}

	codebase.Spec.BuildTool = strings.ToLower(codebase.Spec.BuildTool)

	return &triggersv1.InterceptorResponse{
		Continue:   true,
		Extensions: map[string]interface{}{"spec": codebase.Spec},
	}
}

// getCodebaseByRepoPath returns codebase by repository path.
func (i *EDPInterceptor) getCodebaseByRepoPath(ctx context.Context, ns, repoPath string) (*codebaseApiV1.Codebase, error) {
	codebaseList := &codebaseApiV1.CodebaseList{}
	if err := i.Client.List(ctx, codebaseList, ctrlClient.InNamespace(ns), ctrlClient.Limit(codebaseListLimit)); err != nil {
		return nil, fmt.Errorf("unable to get codebase list: %w", err)
	}

	for i := range codebaseList.Items {
		if codebaseList.Items[i].Spec.GitUrlPath != nil && strings.EqualFold(*codebaseList.Items[i].Spec.GitUrlPath, repoPath) {
			return &codebaseList.Items[i], nil
		}
	}

	return nil, fmt.Errorf("codebase with repository path %s not found", repoPath)
}

// getRepoPath returns repository path from event body.
func getRepoPath(r *triggersv1.InterceptorRequest) (string, error) {
	_, isGitHubEvent := r.Header["X-Github-Event"]
	_, isGitLabEvent := r.Header["X-Gitlab-Event"]

	if isGitLabEvent {
		gitLabEvent := &GitLabEvent{}
		if err := json.Unmarshal([]byte(r.Body), gitLabEvent); err != nil {
			return "", err
		}

		if gitLabEvent.Project.PathWithNamespace == "" {
			return "", errors.New("gitlab repository path empty")
		}

		return strings.ToLower(gitLabEvent.Project.PathWithNamespace), nil
	}

	if isGitHubEvent {
		gitHubEvent := &GitHubEvent{}
		if err := json.Unmarshal([]byte(r.Body), gitHubEvent); err != nil {
			return "", err
		}

		if gitHubEvent.Repository.FullName == "" {
			return "", errors.New("github repository path empty")
		}

		return strings.ToLower(gitHubEvent.Repository.FullName), nil
	}

	gerritEventBody := &GerritEvent{}
	if err := json.Unmarshal([]byte(r.Body), gerritEventBody); err != nil {
		return "", err
	}

	if gerritEventBody.Project.Name == "" {
		return "", errors.New("gerrit repository path empty")
	}

	return strings.ToLower(gerritEventBody.Project.Name), nil
}

// stringPtr returns a pointer to the string value passed in.
func stringP(value string) *string {
	return &value
}
