package interceptor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/event_processor"
	"github.com/epam/edp-tekton/pkg/event_processor/mocks"
)

func bitbucketEventInfo(eventType, headSha, targetBranch string) *event_processor.EventInfo {
	e := &event_processor.EventInfo{
		GitProvider:  event_processor.GitProviderBitbucket,
		RepoPath:     "/o/r",
		TargetBranch: targetBranch,
		Type:         event_processor.EventTypeMerge,
		Codebase: &codebaseApi.Codebase{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "demo"},
			Spec:       codebaseApi.CodebaseSpec{GitUrlPath: "/o/r", CommitMessagePattern: ptr.To(""), JiraServer: ptr.To("")},
		},
		PullRequest: &event_processor.PullRequest{HeadSha: headSha, ChangeNumber: 5},
	}

	switch eventType {
	case event_processor.BitbucketEventTypePullRequestUpdated:
		e.Type = event_processor.EventTypePullRequestUpdate
	case event_processor.BitbucketEventTypeCommentAdded:
		e.Type = event_processor.EventTypeReviewComment
		e.HasPipelineRecheck = true
	}

	return e
}

func reviewBranch(name, branchName string) *codebaseApi.CodebaseBranch {
	return &codebaseApi.CodebaseBranch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      name,
			Labels:    map[string]string{codebaseApi.CodebaseLabel: "demo"},
		},
		Spec: codebaseApi.CodebaseBranchSpec{
			BranchName: branchName,
			Pipelines:  map[string]string{pipelineTypeReview: "review-pipeline"},
		},
	}
}

func bitbucketProcessor(build func(eventType string) *event_processor.EventInfo) *mocks.Processor {
	proc := &mocks.Processor{}
	proc.On("Process", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(_ context.Context, _ []byte, _ string, eventType string) (*event_processor.EventInfo, error) {
			return build(eventType), nil
		})

	return proc
}

func newInterceptor(t *testing.T, bitbucketProc event_processor.Processor, objects ...client.Object) *EDPInterceptor {
	t.Helper()

	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))

	return NewEDPInterceptor(
		fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build(),
		&mocks.Processor{}, &mocks.Processor{}, &mocks.Processor{}, bitbucketProc,
		zap.NewNop().Sugar(),
	)
}

// masterInterceptor wires a Bitbucket processor that always reports head/master and a
// matching review CodebaseBranch, so every event triggers unless the deduper skips it.
func masterInterceptor(t *testing.T, headSha string) *EDPInterceptor {
	t.Helper()

	proc := bitbucketProcessor(func(eventType string) *event_processor.EventInfo {
		return bitbucketEventInfo(eventType, headSha, "master")
	})

	return newInterceptor(t, proc, reviewBranch("demo-master", "master"))
}

func bitbucketRequest(eventKey string) *triggersv1.InterceptorRequest {
	return &triggersv1.InterceptorRequest{
		Body:    "{}",
		Header:  map[string][]string{"X-Event-Key": {eventKey}},
		Context: &triggersv1.TriggerContext{TriggerID: "namespace/default/triggers/name"},
	}
}

func TestEDPInterceptor_Process_BitbucketReviewDedup(t *testing.T) {
	ctx := context.Background()
	update := event_processor.BitbucketEventTypePullRequestUpdated

	t.Run("first update triggers, a repeat of the same head is skipped", func(t *testing.T) {
		i := masterInterceptor(t, "sha1")

		first := i.Process(ctx, bitbucketRequest(update))
		second := i.Process(ctx, bitbucketRequest(update))

		assert.True(t, first.Continue)
		assert.False(t, second.Continue)
	})

	t.Run("created seeds the head so a following update is skipped", func(t *testing.T) {
		i := masterInterceptor(t, "sha1")

		created := i.Process(ctx, bitbucketRequest("pullrequest:created"))
		updateResp := i.Process(ctx, bitbucketRequest(update))

		assert.True(t, created.Continue)
		assert.False(t, updateResp.Continue)
	})

	t.Run("retarget to another branch re-triggers even with the same head", func(t *testing.T) {
		target := "master"
		proc := bitbucketProcessor(func(eventType string) *event_processor.EventInfo {
			return bitbucketEventInfo(eventType, "sha1", target)
		})
		i := newInterceptor(t, proc, reviewBranch("demo-master", "master"), reviewBranch("demo-develop", "develop"))

		first := i.Process(ctx, bitbucketRequest(update))
		target = "develop"
		second := i.Process(ctx, bitbucketRequest(update))

		assert.True(t, first.Continue)
		assert.True(t, second.Continue)
	})

	t.Run("recheck comment always triggers and is not recorded", func(t *testing.T) {
		i := masterInterceptor(t, "sha1")

		comment := i.Process(ctx, bitbucketRequest(event_processor.BitbucketEventTypeCommentAdded))
		updateResp := i.Process(ctx, bitbucketRequest(update))

		assert.True(t, comment.Continue)
		assert.True(t, updateResp.Continue)
	})

	t.Run("update with empty head fails open", func(t *testing.T) {
		i := masterInterceptor(t, "")

		resp := i.Process(ctx, bitbucketRequest(update))

		assert.True(t, resp.Continue)
	})

	t.Run("update without a CodebaseBranch does not trigger or seed", func(t *testing.T) {
		proc := bitbucketProcessor(func(eventType string) *event_processor.EventInfo {
			return bitbucketEventInfo(eventType, "sha1", "master")
		})
		i := newInterceptor(t, proc)

		resp := i.Process(ctx, bitbucketRequest(update))

		assert.False(t, resp.Continue)
	})

	t.Run("non-Bitbucket provider is never deduped", func(t *testing.T) {
		githubProc := bitbucketProcessor(func(eventType string) *event_processor.EventInfo {
			e := bitbucketEventInfo(eventType, "sha1", "master")
			e.GitProvider = event_processor.GitProviderGitHub

			return e
		})
		i := newInterceptor(t, &mocks.Processor{}, reviewBranch("demo-master", "master"))
		i.gitHubProcessor = githubProc

		req := &triggersv1.InterceptorRequest{
			Body:    "{}",
			Header:  map[string][]string{"X-Github-Event": {"pull_request"}},
			Context: &triggersv1.TriggerContext{TriggerID: "namespace/default/triggers/name"},
		}

		first := i.Process(ctx, req)
		second := i.Process(ctx, req)

		assert.True(t, first.Continue)
		assert.True(t, second.Continue)
	})
}
