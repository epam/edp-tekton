package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	tektonpipelineApi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/epam/edp-tekton/pkg/reporter"
	"github.com/epam/edp-tekton/pkg/reporter/gitserver"
)

func TestResolveLogsReporting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		logsEnabled   bool
		runAnnotation string
		gitAnnotation string
		want          bool
	}{
		{
			name:          "global off overrides run annotation and gitserver opinion",
			logsEnabled:   false,
			runAnnotation: "true",
			gitAnnotation: "true",
			want:          false,
		},
		{
			name:        "global on, no opinions anywhere defaults to true",
			logsEnabled: true,
			want:        true,
		},
		{
			name:          "run annotation false wins",
			logsEnabled:   true,
			runAnnotation: "false",
			gitAnnotation: "true",
			want:          false,
		},
		{
			name:          "run annotation true wins over gitserver false",
			logsEnabled:   true,
			runAnnotation: "true",
			gitAnnotation: "false",
			want:          true,
		},
		{
			name:          "malformed run annotation falls through to gitserver false",
			logsEnabled:   true,
			runAnnotation: "yes",
			gitAnnotation: "false",
			want:          false,
		},
		{
			name:          "malformed run annotation with no gitserver opinion defaults to true",
			logsEnabled:   true,
			runAnnotation: "yes",
			want:          true,
		},
		{
			name:          "malformed gitserver annotation defaults to true",
			logsEnabled:   true,
			gitAnnotation: "yes",
			want:          true,
		},
		{
			name:          "gitserver false with no run annotation",
			logsEnabled:   true,
			gitAnnotation: "false",
			want:          false,
		},
		{
			name:          "gitserver true with no run annotation",
			logsEnabled:   true,
			gitAnnotation: "true",
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &reporter.Config{LogsEnabled: tt.logsEnabled}

			gitInfo := &gitserver.Info{}
			if tt.gitAnnotation != "" {
				gitInfo.Annotations = map[string]string{
					reporter.LogsReportingAnnotation: tt.gitAnnotation,
				}
			}

			pipelineRun := &tektonpipelineApi.PipelineRun{ObjectMeta: metav1.ObjectMeta{Name: "review-run"}}
			if tt.runAnnotation != "" {
				pipelineRun.Annotations = map[string]string{
					reporter.LogsReportingAnnotation: tt.runAnnotation,
				}
			}

			got := resolveLogsReporting(context.Background(), cfg, gitInfo, pipelineRun)
			assert.Equal(t, tt.want, got)
		})
	}
}
