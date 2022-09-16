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

const executeTimeOut = 3 * time.Second

type EDPInterceptorInterface interface {
	Execute(r *http.Request) ([]byte, error)
}

type EDPInterceptor struct {
	Client ctrlClient.Reader
	Logger *zap.SugaredLogger
}

func NewEDPInterceptor(c ctrlClient.Reader, l *zap.SugaredLogger) *EDPInterceptor {
	return &EDPInterceptor{
		Client: c,
		Logger: l,
	}
}

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

	i.Logger.Infof("Interceptor request is: %+v", ireq)

	iresp := i.Process(ctx, &ireq)

	respBytes, err := json.Marshal(iresp)
	if err != nil {
		return nil, internal(err)
	}

	i.Logger.Infof("Interceptor response is: %s", respBytes)

	return respBytes, nil
}

func (i *EDPInterceptor) Process(ctx context.Context, r *triggersv1.InterceptorRequest) *triggersv1.InterceptorResponse {
	ns, _ := triggersv1.ParseTriggerID(r.Context.TriggerID)

	repoName, err := getRepoName(r)
	if err != nil {
		return interceptors.Failf(codes.InvalidArgument, "failed to get repository name: %v", err)
	}

	objectKey := ctrlClient.ObjectKey{
		Namespace: ns,
		Name:      repoName,
	}
	codebase := &codebaseApiV1.Codebase{}

	if err := i.Client.Get(ctx, objectKey, codebase); err != nil {
		return interceptors.Failf(codes.InvalidArgument, "failed to get codebase %s: %v", objectKey, err)
	}

	framework := strings.ToLower(*codebase.Spec.Framework)
	codebase.Spec.Framework = &framework
	codebase.Spec.BuildTool = strings.ToLower(codebase.Spec.BuildTool)

	return &triggersv1.InterceptorResponse{
		Continue:   true,
		Extensions: map[string]interface{}{"spec": codebase.Spec},
	}
}

func getRepoName(r *triggersv1.InterceptorRequest) (string, error) {
	_, isGitHubEvent := r.Header["X-GitHub-Event"]
	_, isGitLabEvent := r.Header["X-Gitlab-Event"]

	if isGitHubEvent || isGitLabEvent {
		gitEventBody := &GitEventBody{}
		if err := json.Unmarshal([]byte(r.Body), gitEventBody); err != nil {
			return "", err
		}

		if gitEventBody.Repository.Name == "" {
			return "", errors.New("repository name is empty")
		}

		return strings.ToLower(gitEventBody.Repository.Name), nil
	}

	gerritEventBody := &GerritEventBody{}
	if err := json.Unmarshal([]byte(r.Body), gerritEventBody); err != nil {
		return "", err
	}

	if gerritEventBody.Project.Name == "" {
		return "", errors.New("project name is empty")
	}

	return strings.ToLower(gerritEventBody.Project.Name), nil
}
