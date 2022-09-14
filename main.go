package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tektoncd/triggers/pkg/interceptors"
	"google.golang.org/grpc/codes"
	"io"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	codebaseApiV1 "github.com/epam/edp-codebase-operator/v2/pkg/apis/edp/v1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"go.uber.org/zap"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/signals"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Port is the port that the port that interceptor service listens on
	Port         = 8080
	readTimeout  = 5 * time.Second
	writeTimeout = 20 * time.Second
	idleTimeout  = 60 * time.Second
)

func main() {
	ctx := signals.NewContext()

	clusterConfig := ctrl.GetConfigOrDie()

	scheme := runtime.NewScheme()
	utilruntime.Must(codebaseApiV1.AddToScheme(scheme))

	client, err := ctrlClient.New(clusterConfig, ctrlClient.Options{Scheme: scheme})
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}

	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %s", err)
	}

	logger := zapLogger.Sugar()
	ctx = logging.WithLogger(ctx, logger)
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Fatalf("failed to sync the logger: %s", err)
		}
	}()

	service := NewEDPInterceptor(client, logger)

	//TODO: We need to move to https server. See: https://tekton.dev/docs/triggers/clusterinterceptors/#running-clusterinterceptor-as-https
	mux := http.NewServeMux()
	mux.Handle("/", service)
	mux.HandleFunc("/ready", readinessHandler)

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", Port),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      mux,
	}

	logger.Infof("Listen and serve on port %d", Port)

	if err := srv.ListenAndServe(); err != nil {
		logger.Fatalf("failed to start interceptors service: %v", err)
	}
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

func (i *EDPInterceptor) Process(ctx context.Context, r *triggersv1.InterceptorRequest) *triggersv1.InterceptorResponse {
	ns, _ := triggersv1.ParseTriggerID(r.Context.TriggerID)

	repoName, err := i.getRepoName(r)
	if err != nil {
		return interceptors.Failf(codes.InvalidArgument, "failed to unmarshal event body: %v", err)
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

func (i *EDPInterceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := i.executeInterceptor(r)
	if err != nil {
		switch e := err.(type) {
		case Error:
			i.Logger.Infof("HTTP %d - %s", e.Status(), e)
			http.Error(w, e.Error(), e.Status())
		default:
			i.Logger.Errorf("Non Status Error: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(b); err != nil {
		i.Logger.Errorf("failed to write response: %s", err)
	}
}

func (i *EDPInterceptor) executeInterceptor(r *http.Request) ([]byte, error) {
	// Create a context
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var body bytes.Buffer
	defer r.Body.Close()

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

func (i *EDPInterceptor) getRepoName(r *triggersv1.InterceptorRequest) (string, error) {
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

type GerritEventBody struct {
	Project struct {
		Name string `json:"name"`
	} `json:"project"`
}

type GitEventBody struct {
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// HTTPError represents an error with an associated HTTP status code.
type HTTPError struct {
	Code int
	Err  error
}

// Allows HTTPError to satisfy the error interface.
func (se HTTPError) Error() string {
	return se.Err.Error()
}

// Returns our HTTP status code.
func (se HTTPError) Status() int {
	return se.Code
}

func badRequest(err error) HTTPError {
	return HTTPError{Code: http.StatusBadRequest, Err: err}
}

func internal(err error) HTTPError {
	return HTTPError{Code: http.StatusInternalServerError, Err: err}
}
