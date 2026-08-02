package interceptor

import (
	"context"
	"os"
	"strings"

	tektonpipelineApi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"
)

const (
	// serviceAccountAnnotation on a Tekton Pipeline names the ServiceAccount its
	// PipelineRuns should run as. Absent/empty falls back to the default.
	serviceAccountAnnotation = "app.edp.epam.com/service-account"

	pipelineTypeBuild = "build"

	defaultSAEnv = "TEKTON_SA_DEFAULT"

	defaultSA = "tekton-unprivileged"
)

func newDefaultServiceAccount() string {
	if v := os.Getenv(defaultSAEnv); v != "" {
		return v
	}

	return defaultSA
}

// resolveServiceAccounts always returns both "build" and "review" keys: TriggerBindings
// dereference this extension unconditionally, so a missing key or empty value hard-fails
// the webhook event. Resolution is strictly best-effort — any lookup failure silently
// falls back to the default rather than blocking the response.
func (i *EDPInterceptor) resolveServiceAccounts(
	ctx context.Context,
	log *zap.SugaredLogger,
	ns string,
	codebaseBranch *codebaseApi.CodebaseBranch,
) map[string]string {
	result := map[string]string{
		pipelineTypeBuild:  i.defaultSA,
		pipelineTypeReview: i.defaultSA,
	}

	if codebaseBranch == nil {
		return result
	}

	for _, pipelineType := range []string{pipelineTypeBuild, pipelineTypeReview} {
		pipelineName := codebaseBranch.Spec.Pipelines[pipelineType]
		if pipelineName == "" {
			continue
		}

		result[pipelineType] = i.serviceAccountForPipeline(ctx, log, ns, pipelineName)
	}

	return result
}

func (i *EDPInterceptor) serviceAccountForPipeline(
	ctx context.Context,
	log *zap.SugaredLogger,
	ns, pipelineName string,
) string {
	// The annotation lives in metadata, so a metadata-only read avoids pulling
	// the full Pipeline spec on every webhook.
	pipeline := &metav1.PartialObjectMetadata{}
	pipeline.SetGroupVersionKind(tektonpipelineApi.SchemeGroupVersion.WithKind("Pipeline"))

	if err := i.client.Get(ctx, ctrlClient.ObjectKey{Namespace: ns, Name: pipelineName}, pipeline); err != nil {
		log.Debugf("Failed to get Pipeline %s to resolve ServiceAccount, using default %q: %s",
			pipelineName, i.defaultSA, err)

		return i.defaultSA
	}

	if sa := strings.TrimSpace(pipeline.Annotations[serviceAccountAnnotation]); sa != "" {
		return sa
	}

	return i.defaultSA
}
