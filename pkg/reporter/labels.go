// Package reporter contains the pipeline reporter component that publishes
// review PipelineRun results as pull request comments to git providers.
package reporter

const (
	// PipelineTypeLabel marks a PipelineRun with its pipeline type (review, build).
	PipelineTypeLabel = "app.edp.epam.com/pipelinetype"

	// PipelineTypeReview is the PipelineTypeLabel value for review pipelines.
	PipelineTypeReview = "review"

	// ChangeNumberLabel holds the pull/merge request number of a review PipelineRun.
	ChangeNumberLabel = "app.edp.epam.com/git-change-number"

	// ReportedAnnotation marks a PipelineRun whose result has been published to
	// the pull request. It is a plain annotation (never a finalizer), so it puts
	// no constraints on PipelineRun deletion or pruning.
	ReportedAnnotation = "app.edp.epam.com/pipeline-reported"

	// ResultAnnotationsKey is the annotation holding a JSON object with git
	// metadata (repository full name, change number, change URL) rendered by
	// the review TriggerTemplates.
	ResultAnnotationsKey = "results.tekton.dev/resultAnnotations"

	// GitRepositoryAnnotation is the key inside ResultAnnotationsKey JSON with
	// the full repository name (org/repo).
	GitRepositoryAnnotation = "app.edp.epam.com/git-repository"

	// GitChangeNumberAnnotation is the key inside ResultAnnotationsKey JSON with
	// the pull/merge request number.
	GitChangeNumberAnnotation = "app.edp.epam.com/git-change-number"
)
