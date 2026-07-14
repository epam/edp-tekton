// Package formatter renders a PipelineRun report as a markdown pull request comment.
package formatter

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/epam/edp-tekton/pkg/reporter/collector"
)

const (
	statusPassed = "✅"
	statusFailed = "❌"
)

// LinkBuilder renders links to pipeline details. Implementations returning an
// empty string produce plain text instead of a link.
type LinkBuilder interface {
	// PipelineRunURL returns a link to the PipelineRun details page.
	PipelineRunURL(namespace, name string) string
	// TaskURL returns a link to a single task of the PipelineRun. Reserved for
	// per-task deep links; return "" to render the task name as plain text.
	TaskURL(namespace, pipelineRunName, taskName string) string
}

// PortalLinkBuilder builds links to the KubeRocketCI portal.
type PortalLinkBuilder struct {
	// BaseURL is the portal pipelineruns base, e.g.
	// https://portal.example.com/c/cluster/cicd/pipelineruns. Empty disables links.
	BaseURL string
}

// PipelineRunURL returns the portal PipelineRun details URL.
func (b PortalLinkBuilder) PipelineRunURL(namespace, name string) string {
	if b.BaseURL == "" {
		return ""
	}

	return fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(b.BaseURL, "/"), namespace, name)
}

// TaskURL returns "" until the portal exposes per-task deep links.
func (b PortalLinkBuilder) TaskURL(_, _, _ string) string {
	return ""
}

// Formatter renders reports as markdown comments.
type Formatter struct {
	links LinkBuilder
}

// New creates a Formatter with the given link builder.
func New(links LinkBuilder) *Formatter {
	return &Formatter{links: links}
}

// Format renders the report as a markdown comment starting with the given
// hidden marker. Every task is listed in a status table; failed steps get a
// collapsible log section below the table.
func (f *Formatter) Format(report *collector.Report, marker string, tailLines int64) string {
	var b strings.Builder

	b.WriteString(marker)
	b.WriteString("\n")
	b.WriteString(f.header(report))
	b.WriteString("\n\n| Status | Task | Duration |\n|---|---|---|\n")

	tasks := make([]collector.TaskResult, len(report.Tasks))
	copy(tasks, report.Tasks)
	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].StartTime.Before(tasks[j].StartTime)
	})

	for _, task := range tasks {
		status := statusPassed
		if !task.Succeeded {
			status = statusFailed
		}

		name := task.Name
		if url := f.links.TaskURL(report.PipelineRunNamespace, report.PipelineRunName, task.Name); url != "" {
			name = fmt.Sprintf("[%s](%s)", task.Name, url)
		}

		fmt.Fprintf(&b, "| %s | %s | %s |\n", status, name, formatDuration(task.Duration))
	}

	for _, task := range tasks {
		if task.Succeeded {
			continue
		}

		for _, step := range task.Steps {
			if step.Succeeded || step.LogTail == "" {
				continue
			}

			fmt.Fprintf(&b,
				"\n<details><summary>%s <b>%s</b> / %s (exit code %d, last %d log lines)</summary>\n\n```\n%s\n```\n</details>\n",
				statusFailed, task.Name, step.Name, step.ExitCode, tailLines, sanitizeCodeFence(step.LogTail),
			)
		}
	}

	return b.String()
}

func (f *Formatter) header(report *collector.Report) string {
	status := fmt.Sprintf("%s Passed", statusPassed)
	if !report.Succeeded {
		status = fmt.Sprintf("%s Failed", statusFailed)
	}

	name := fmt.Sprintf("`%s`", report.PipelineRunName)
	if url := f.links.PipelineRunURL(report.PipelineRunNamespace, report.PipelineRunName); url != "" {
		name = fmt.Sprintf("[`%s`](%s)", report.PipelineRunName, url)
	}

	return fmt.Sprintf("## Pipeline %s %s", name, status)
}

// codeFenceRun matches any run of three or more backticks; a plain
// strings.ReplaceAll of "```" can re-create a fence from longer runs.
var codeFenceRun = regexp.MustCompile("`{3,}")

// sanitizeCodeFence prevents log content from breaking out of the enclosing
// markdown code fence by interleaving zero-width spaces into backtick runs.
func sanitizeCodeFence(log string) string {
	return codeFenceRun.ReplaceAllStringFunc(log, func(run string) string {
		return strings.Join(strings.Split(run, ""), "\u200b")
	})
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "-"
	}

	return d.Round(time.Second).String()
}
