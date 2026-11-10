package interceptor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc/codes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	clientinterceptor "sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/event_processor"
	"github.com/epam/edp-tekton/pkg/event_processor/mocks"
)

const (
	testEventID   = "event-1"
	testTriggerID = "namespace/default/triggers/github-review"
)

func TestEDPInterceptor_Process_LogsFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		processor   func() event_processor.Processor
		clientFuncs *clientinterceptor.Funcs
		wantCode    codes.Code
		check       func(t *testing.T, logs *observer.ObservedLogs)
	}{
		{
			name: "processor error is logged with the event ID",
			processor: func() event_processor.Processor {
				m := &mocks.Processor{}
				m.On("Process", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("GET https://api.github.com/repos/o/r/pulls/1: 404 Not Found []"))

				return m
			},
			wantCode: codes.InvalidArgument,
			check: func(t *testing.T, logs *observer.ObservedLogs) {
				entries := logs.FilterLevelExact(zapcore.ErrorLevel).All()
				require.Len(t, entries, 1)
				assert.Equal(t, "Failed to process webhook event", entries[0].Message)

				fields := entries[0].ContextMap()
				assert.Equal(t, testEventID, fields["eventID"])
				assert.Equal(t, testTriggerID, fields["trigger"])
				assert.Contains(t, fields["error"], "failed to process GitHub event")
				assert.Contains(t, fields["error"], "404 Not Found")
			},
		},
		{
			name: "CodebaseBranch lookup error is logged with the repository",
			processor: func() event_processor.Processor {
				return processorReturning(&event_processor.EventInfo{
					GitProvider:  event_processor.GitProviderGitHub,
					RepoPath:     "/o/r",
					TargetBranch: "master",
					Type:         event_processor.EventTypeMerge,
					Codebase: &codebaseApi.Codebase{
						ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "demo"},
					},
				})
			},
			clientFuncs: &clientinterceptor.Funcs{
				List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
					return errors.New("forbidden")
				},
			},
			wantCode: codes.Internal,
			check: func(t *testing.T, logs *observer.ObservedLogs) {
				entries := logs.FilterLevelExact(zapcore.ErrorLevel).All()
				require.Len(t, entries, 1)
				assert.Equal(t, "Failed to process webhook event", entries[0].Message)

				fields := entries[0].ContextMap()
				assert.Equal(t, testEventID, fields["eventID"])
				assert.Equal(t, "/o/r", fields["repo"])
				assert.Contains(t, fields["error"], "forbidden")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			scheme := runtime.NewScheme()
			require.NoError(t, codebaseApi.AddToScheme(scheme))

			builder := fake.NewClientBuilder().WithScheme(scheme)
			if tt.clientFuncs != nil {
				builder = builder.WithInterceptorFuncs(*tt.clientFuncs)
			}

			core, logs := observer.New(zapcore.InfoLevel)

			i := NewEDPInterceptor(
				builder.Build(),
				tt.processor(), &mocks.Processor{}, &mocks.Processor{}, &mocks.Processor{},
				zap.New(core).Sugar(),
			)

			resp := i.Process(context.Background(), &triggersv1.InterceptorRequest{
				Body:   "{}",
				Header: map[string][]string{"X-Github-Event": {"pull_request"}},
				Context: &triggersv1.TriggerContext{
					EventID:   testEventID,
					TriggerID: testTriggerID,
				},
			})

			require.NotNil(t, resp)
			assert.False(t, resp.Continue)
			require.NotNil(t, resp.Status)
			assert.Equal(t, tt.wantCode, resp.Status.Code)

			tt.check(t, logs)
		})
	}
}

func processorReturning(event *event_processor.EventInfo) event_processor.Processor {
	m := &mocks.Processor{}
	m.On("Process", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(event, nil)

	return m
}
