package interceptor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tektonpipelineApi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	clientinterceptor "sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"
)

func TestNewDefaultServiceAccount(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want string
	}{
		{
			name: "no env override, built-in default",
			env:  "",
			want: defaultSA,
		},
		{
			name: "env override respected",
			env:  "custom-sa",
			want: "custom-sa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(defaultSAEnv, tt.env)

			assert.Equal(t, tt.want, newDefaultServiceAccount())
		})
	}
}

func TestEDPInterceptor_ResolveServiceAccounts(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, tektonpipelineApi.AddToScheme(scheme))

	newPipeline := func(name string, annotations map[string]string) *tektonpipelineApi.Pipeline {
		return &tektonpipelineApi.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   "default",
				Name:        name,
				Annotations: annotations,
			},
		}
	}

	tests := []struct {
		name           string
		kubeObjects    []ctrlClient.Object
		interceptFuncs *clientinterceptor.Funcs
		codebaseBranch *codebaseApi.CodebaseBranch
		want           map[string]string
	}{
		{
			name:           "nil codebase branch uses defaults for both types",
			codebaseBranch: nil,
			want: map[string]string{
				pipelineTypeBuild:  defaultSA,
				pipelineTypeReview: defaultSA,
			},
		},
		{
			name: "annotation present on both pipelines is used",
			kubeObjects: []ctrlClient.Object{
				newPipeline("build-pipeline", map[string]string{serviceAccountAnnotation: "build-sa"}),
				newPipeline("review-pipeline", map[string]string{serviceAccountAnnotation: "review-sa"}),
			},
			codebaseBranch: &codebaseApi.CodebaseBranch{
				Spec: codebaseApi.CodebaseBranchSpec{
					Pipelines: map[string]string{
						pipelineTypeBuild:  "build-pipeline",
						pipelineTypeReview: "review-pipeline",
					},
				},
			},
			want: map[string]string{
				pipelineTypeBuild:  "build-sa",
				pipelineTypeReview: "review-sa",
			},
		},
		{
			name: "annotation absent falls back to default",
			kubeObjects: []ctrlClient.Object{
				newPipeline("build-pipeline", nil),
				newPipeline("review-pipeline", map[string]string{serviceAccountAnnotation: ""}),
			},
			codebaseBranch: &codebaseApi.CodebaseBranch{
				Spec: codebaseApi.CodebaseBranchSpec{
					Pipelines: map[string]string{
						pipelineTypeBuild:  "build-pipeline",
						pipelineTypeReview: "review-pipeline",
					},
				},
			},
			want: map[string]string{
				pipelineTypeBuild:  defaultSA,
				pipelineTypeReview: defaultSA,
			},
		},
		{
			name: "whitespace-only annotation falls back to default",
			kubeObjects: []ctrlClient.Object{
				newPipeline("build-pipeline", map[string]string{serviceAccountAnnotation: "   "}),
				newPipeline("review-pipeline", map[string]string{serviceAccountAnnotation: "\t\n"}),
			},
			codebaseBranch: &codebaseApi.CodebaseBranch{
				Spec: codebaseApi.CodebaseBranchSpec{
					Pipelines: map[string]string{
						pipelineTypeBuild:  "build-pipeline",
						pipelineTypeReview: "review-pipeline",
					},
				},
			},
			want: map[string]string{
				pipelineTypeBuild:  defaultSA,
				pipelineTypeReview: defaultSA,
			},
		},
		{
			name: "annotation value is trimmed",
			kubeObjects: []ctrlClient.Object{
				newPipeline("review-pipeline", map[string]string{serviceAccountAnnotation: " review-sa "}),
			},
			codebaseBranch: &codebaseApi.CodebaseBranch{
				Spec: codebaseApi.CodebaseBranchSpec{
					Pipelines: map[string]string{
						pipelineTypeReview: "review-pipeline",
					},
				},
			},
			want: map[string]string{
				pipelineTypeBuild:  defaultSA,
				pipelineTypeReview: "review-sa",
			},
		},
		{
			name: "types resolve independently: only build annotated",
			kubeObjects: []ctrlClient.Object{
				newPipeline("build-pipeline", map[string]string{serviceAccountAnnotation: "build-sa"}),
				newPipeline("review-pipeline", nil),
			},
			codebaseBranch: &codebaseApi.CodebaseBranch{
				Spec: codebaseApi.CodebaseBranchSpec{
					Pipelines: map[string]string{
						pipelineTypeBuild:  "build-pipeline",
						pipelineTypeReview: "review-pipeline",
					},
				},
			},
			want: map[string]string{
				pipelineTypeBuild:  "build-sa",
				pipelineTypeReview: defaultSA,
			},
		},
		{
			name: "partial pipelines map: review only, review pipeline missing",
			codebaseBranch: &codebaseApi.CodebaseBranch{
				Spec: codebaseApi.CodebaseBranchSpec{
					Pipelines: map[string]string{
						pipelineTypeReview: "missing-review-pipeline",
					},
				},
			},
			want: map[string]string{
				pipelineTypeBuild:  defaultSA,
				pipelineTypeReview: defaultSA,
			},
		},
		{
			name: "pipeline not found falls back to default",
			codebaseBranch: &codebaseApi.CodebaseBranch{
				Spec: codebaseApi.CodebaseBranchSpec{
					Pipelines: map[string]string{
						pipelineTypeBuild:  "missing-build-pipeline",
						pipelineTypeReview: "missing-review-pipeline",
					},
				},
			},
			want: map[string]string{
				pipelineTypeBuild:  defaultSA,
				pipelineTypeReview: defaultSA,
			},
		},
		{
			name: "get error falls back to default without failing",
			kubeObjects: []ctrlClient.Object{
				newPipeline("build-pipeline", map[string]string{serviceAccountAnnotation: "build-sa"}),
			},
			interceptFuncs: &clientinterceptor.Funcs{
				Get: func(
					context.Context, ctrlClient.WithWatch, ctrlClient.ObjectKey, ctrlClient.Object, ...ctrlClient.GetOption,
				) error {
					return errors.New("api server unavailable")
				},
			},
			codebaseBranch: &codebaseApi.CodebaseBranch{
				Spec: codebaseApi.CodebaseBranchSpec{
					Pipelines: map[string]string{
						pipelineTypeBuild: "build-pipeline",
					},
				},
			},
			want: map[string]string{
				pipelineTypeBuild:  defaultSA,
				pipelineTypeReview: defaultSA,
			},
		},
		{
			name: "empty pipeline name uses default",
			codebaseBranch: &codebaseApi.CodebaseBranch{
				Spec: codebaseApi.CodebaseBranchSpec{
					Pipelines: map[string]string{
						pipelineTypeBuild:  "",
						pipelineTypeReview: "",
					},
				},
			},
			want: map[string]string{
				pipelineTypeBuild:  defaultSA,
				pipelineTypeReview: defaultSA,
			},
		},
		{
			name: "missing pipelines map uses default",
			codebaseBranch: &codebaseApi.CodebaseBranch{
				Spec: codebaseApi.CodebaseBranchSpec{},
			},
			want: map[string]string{
				pipelineTypeBuild:  defaultSA,
				pipelineTypeReview: defaultSA,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.kubeObjects...)
			if tt.interceptFuncs != nil {
				builder = builder.WithInterceptorFuncs(*tt.interceptFuncs)
			}

			i := &EDPInterceptor{
				client:    builder.Build(),
				defaultSA: newDefaultServiceAccount(),
			}

			got := i.resolveServiceAccounts(context.Background(), zap.NewNop().Sugar(), "default", tt.codebaseBranch)
			assert.Equal(t, tt.want, got)
			assert.Contains(t, got, pipelineTypeBuild, "serviceAccounts must always carry a build key")
			assert.Contains(t, got, pipelineTypeReview, "serviceAccounts must always carry a review key")
			assert.NotEmpty(t, got[pipelineTypeBuild])
			assert.NotEmpty(t, got[pipelineTypeReview])
		})
	}
}
