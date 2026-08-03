// Package controller reconciles finished review PipelineRuns and publishes
// their results as pull request comments.
package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	tektonpipelineApi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/reporter"
	"github.com/epam/edp-tekton/pkg/reporter/collector"
	"github.com/epam/edp-tekton/pkg/reporter/formatter"
	"github.com/epam/edp-tekton/pkg/reporter/gitserver"
	"github.com/epam/edp-tekton/pkg/reporter/provider/types"
	"github.com/epam/edp-tekton/pkg/reporter/secretmask"
)

// ProviderFactory builds a git provider client; injectable for tests.
type ProviderFactory func(gitProvider, host, token string) (types.Provider, error)

// PipelineRunReconciler publishes finished review PipelineRun results to pull requests.
type PipelineRunReconciler struct {
	// client is the cached client, backed by the label-filtered PipelineRun watch.
	client ctrlClient.Client
	// reader performs direct (uncached) reads of TaskRuns, Secrets, Codebases
	// and GitServers, so the reporter never opens watches on those types.
	reader      ctrlClient.Reader
	collector   *collector.Collector
	formatter   *formatter.Formatter
	newProvider ProviderFactory
	config      *reporter.Config
}

// NewPipelineRunReconciler creates the reconciler.
func NewPipelineRunReconciler(
	client ctrlClient.Client,
	reader ctrlClient.Reader,
	logCollector *collector.Collector,
	commentFormatter *formatter.Formatter,
	newProvider ProviderFactory,
	config *reporter.Config,
) *PipelineRunReconciler {
	return &PipelineRunReconciler{
		client:      client,
		reader:      reader,
		collector:   logCollector,
		formatter:   commentFormatter,
		newProvider: newProvider,
		config:      config,
	}
}

// permanentError marks failures that will not resolve on retry, so the
// reconciler logs them without requeueing.
type permanentError struct {
	err error
}

func (e *permanentError) Error() string { return e.err.Error() }

func (e *permanentError) Unwrap() error { return e.err }

func permanent(err error) error { return &permanentError{err: err} }

// SetupWithManager registers the reconciler for finished review PipelineRuns.
func (r *PipelineRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.NewPredicateFuncs(func(obj ctrlClient.Object) bool {
		pipelineRun, ok := obj.(*tektonpipelineApi.PipelineRun)
		if !ok {
			return false
		}

		return isReportable(pipelineRun)
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&tektonpipelineApi.PipelineRun{}, builder.WithPredicates(pred)).
		Complete(r)
}

// isReportable tells whether the PipelineRun is a finished review run that has
// not been reported yet.
func isReportable(pipelineRun *tektonpipelineApi.PipelineRun) bool {
	if pipelineRun.Labels[reporter.PipelineTypeLabel] != reporter.PipelineTypeReview {
		return false
	}

	if !pipelineRun.IsDone() {
		return false
	}

	_, reported := pipelineRun.Annotations[reporter.ReportedAnnotation]

	return !reported
}

// Reconcile publishes the PipelineRun result to its pull request exactly once.
func (r *PipelineRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	pipelineRun := &tektonpipelineApi.PipelineRun{}
	if err := r.client.Get(ctx, req.NamespacedName, pipelineRun); err != nil {
		if k8sErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get PipelineRun: %w", err)
	}

	if !isReportable(pipelineRun) {
		return ctrl.Result{}, nil
	}

	if err := r.report(ctx, pipelineRun); err != nil {
		permErr := &permanentError{}
		if errors.As(err, &permErr) {
			logger.Info("Skipping PipelineRun report: not supported", "reason", err.Error())

			// Mark it handled so an unsupported/misconfigured run (e.g. a Gerrit
			// review, which has no provider yet) is not re-reconciled and
			// re-logged on every informer resync.
			return ctrl.Result{}, r.markHandled(ctx, pipelineRun, reportSkipped)
		}

		return ctrl.Result{}, err
	}

	if err := r.markHandled(ctx, pipelineRun, reportPublished); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Published PipelineRun report to the pull request")

	return ctrl.Result{}, nil
}

func (r *PipelineRunReconciler) report(ctx context.Context, pipelineRun *tektonpipelineApi.PipelineRun) error {
	codebaseName := pipelineRun.Labels[codebaseApi.CodebaseLabel]
	if codebaseName == "" {
		return permanent(fmt.Errorf("PipelineRun has no %s label", codebaseApi.CodebaseLabel))
	}

	gitInfo, err := gitserver.Resolve(ctx, r.reader, pipelineRun.Namespace, codebaseName)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return permanent(err)
		}

		return err
	}

	pullRequest, err := pullRequestRef(pipelineRun)
	if err != nil {
		return permanent(err)
	}

	gitProvider, err := r.newProvider(gitInfo.Provider, gitInfo.Host, gitInfo.Token)
	if err != nil {
		return permanent(err)
	}

	fetchLogs := resolveLogsReporting(ctx, r.config, gitInfo, pipelineRun)
	collectOpts := collector.Options{FetchLogs: fetchLogs, TailLines: r.config.TailLines}

	report, err := r.collector.Collect(ctx, pipelineRun, collectOpts)
	if err != nil {
		return err
	}

	masker := secretmask.NewMasker([]string{gitInfo.Token})

	for taskIdx := range report.Tasks {
		for stepIdx := range report.Tasks[taskIdx].Steps {
			step := &report.Tasks[taskIdx].Steps[stepIdx]
			step.LogTail = masker.Mask(step.LogTail)
		}
	}

	marker := fmt.Sprintf("<!-- krci-pipeline-report codebase=%s -->", codebaseName)

	collapsible := false
	if c, ok := gitProvider.(types.CollapsibleSectionsSupport); ok {
		collapsible = c.SupportsCollapsibleSections()
	}

	body := formatter.Truncate(
		r.formatter.Format(report, marker, formatter.Options{
			TailLines:           r.config.TailLines,
			CollapsibleSections: collapsible,
		}),
		reporter.MaxCommentBytes,
	)

	if err := gitProvider.UpsertComment(ctx, pullRequest, types.Comment{
		Marker: marker,
		Body:   body,
		Update: r.config.CommentStrategy == reporter.CommentStrategyUpdate,
	}); err != nil {
		return fmt.Errorf("failed to publish report comment: %w", err)
	}

	return nil
}

// Report outcomes recorded in the reported annotation. isReportable only checks
// for the annotation's presence, so the value is informational.
const (
	reportPublished = "true"
	reportSkipped   = "skipped"
)

func (r *PipelineRunReconciler) markHandled(
	ctx context.Context,
	pipelineRun *tektonpipelineApi.PipelineRun,
	outcome string,
) error {
	patch := ctrlClient.MergeFrom(pipelineRun.DeepCopy())

	if pipelineRun.Annotations == nil {
		pipelineRun.Annotations = map[string]string{}
	}

	pipelineRun.Annotations[reporter.ReportedAnnotation] = outcome

	if err := r.client.Patch(ctx, pipelineRun, patch); err != nil {
		return fmt.Errorf("failed to mark PipelineRun as reported: %w", err)
	}

	return nil
}

// resolveLogsReporting decides whether failed-step log tails are fetched for
// this PipelineRun. The global config flag is a hard ceiling: when it is off,
// the reporter has no pods/log RBAC, so annotations are never consulted. When
// it is on, a PipelineRun override wins over a GitServer override, which wins
// over the global default (enabled).
func resolveLogsReporting(
	ctx context.Context,
	cfg *reporter.Config,
	gitInfo *gitserver.Info,
	pipelineRun *tektonpipelineApi.PipelineRun,
) bool {
	if !cfg.LogsEnabled {
		return false
	}

	if v := logsOverride(ctx, pipelineRun.Annotations, "PipelineRun "+pipelineRun.Name); v != nil {
		return *v
	}

	if v := logsOverride(ctx, gitInfo.Annotations, "GitServer"); v != nil {
		return *v
	}

	return true
}

// logsOverride reads the logs-reporting override annotation from an object's
// annotations. A malformed value is logged and treated as absent, so a typo
// falls through the cascade instead of failing the report.
func logsOverride(ctx context.Context, annotations map[string]string, subject string) *bool {
	raw, ok := annotations[reporter.LogsReportingAnnotation]
	if !ok {
		return nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		log.FromContext(ctx).Info(
			"Malformed logs-reporting annotation, falling through",
			"annotation", reporter.LogsReportingAnnotation,
			"value", raw,
			"subject", subject,
		)

		return nil
	}

	return &value
}

// pullRequestRef extracts the repository full name and pull request number
// from the PipelineRun metadata rendered by the review TriggerTemplates.
func pullRequestRef(pipelineRun *tektonpipelineApi.PipelineRun) (types.PullRequestRef, error) {
	resultAnnotations := map[string]string{}

	if raw := pipelineRun.Annotations[reporter.ResultAnnotationsKey]; raw != "" {
		if err := json.Unmarshal([]byte(raw), &resultAnnotations); err != nil {
			return types.PullRequestRef{}, fmt.Errorf("failed to parse %s annotation: %w", reporter.ResultAnnotationsKey, err)
		}
	}

	repo := resultAnnotations[reporter.GitRepositoryAnnotation]
	if repo == "" {
		return types.PullRequestRef{},
			fmt.Errorf("PipelineRun metadata has no repository name (%s)", reporter.GitRepositoryAnnotation)
	}

	changeNumber := resultAnnotations[reporter.GitChangeNumberAnnotation]
	if changeNumber == "" {
		changeNumber = pipelineRun.Labels[reporter.ChangeNumberLabel]
	}

	number, err := strconv.Atoi(changeNumber)
	if err != nil || number <= 0 {
		return types.PullRequestRef{}, fmt.Errorf("invalid pull request number %q", changeNumber)
	}

	return types.PullRequestRef{RepoFullName: repo, Number: number}, nil
}
