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

// U+2713/U+2717 have no emoji presentation form, so renderers draw them
// monochrome in the text font. Each mark is paired with its label to keep the
// outcome readable without relying on the glyph.
const (
	markPassed  = "✓"
	markFailed  = "✗"
	labelPassed = "Passed"
	labelFailed = "Failed"
)

// retriggerNotice tells the reviewer how to re-run the pipeline. It sits
// directly under the task table rather than at the end of the comment so
// Truncate, which cuts oversized bodies from the tail, cannot drop it.
const retriggerNotice = "\n> Pushing new commits re-runs this pipeline automatically.\n" +
	"> To re-run it without new commits, comment `/recheck`.\n"

// LinkBuilder renders links to pipeline details. Implementations returning an
// empty string produce plain text instead of a link.
type LinkBuilder interface {
	PipelineRunURL(namespace, name string) string
	// TaskURL returns a link to a single task of the PipelineRun. Reserved for
	// per-task deep links; return "" to render the task name as plain text.
	TaskURL(namespace, pipelineRunName, taskName string) string
}

type PortalLinkBuilder struct {
	// BaseURL is the portal pipelineruns base, e.g.
	// https://portal.example.com/c/cluster/cicd/pipelineruns. Empty disables links.
	BaseURL string
}

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

type Formatter struct {
	links LinkBuilder
}

func New(links LinkBuilder) *Formatter {
	return &Formatter{links: links}
}

// Options control how Format renders the failed-step log sections.
type Options struct {
	// TailLines is the log-line count echoed in each failed-step heading.
	TailLines int64
	// CollapsibleSections wraps failed-step logs in <details> blocks; leave
	// false for renderers (e.g. Bitbucket Cloud) that escape embedded HTML
	// instead of executing it.
	CollapsibleSections bool
}

// Format renders the report as a markdown comment starting with the given
// hidden marker. Every task is listed in a status table; failed steps get a
// log section below the table, shaped per opts.
func (f *Formatter) Format(report *collector.Report, marker string, opts Options) string {
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
		name := task.Name
		if url := f.links.TaskURL(report.PipelineRunNamespace, report.PipelineRunName, task.Name); url != "" {
			name = fmt.Sprintf("[%s](%s)", task.Name, url)
		}

		status := fmt.Sprintf("%s %s", markPassed, labelPassed)
		if !task.Succeeded {
			// Failed rows are bold so the one row that matters stands out in a
			// long, mostly-green table; renderers do not color the glyphs.
			status = fmt.Sprintf("**%s %s**", markFailed, labelFailed)
			name = fmt.Sprintf("**%s**", name)
		}

		fmt.Fprintf(&b, "| %s | %s | %s |\n", status, name, formatDuration(task.Duration))
	}

	b.WriteString(retriggerNotice)

	for _, task := range tasks {
		if task.Succeeded {
			continue
		}

		for _, step := range task.Steps {
			if step.Succeeded || step.LogTail == "" {
				continue
			}

			tmpl := "\n%s **%s / %s** — exit code %d, last %d log lines\n\n```\n%s\n```\n"
			if opts.CollapsibleSections {
				tmpl = "\n<details><summary>%s <b>%s / %s</b> — exit code %d, last %d log lines</summary>" +
					"\n\n```\n%s\n```\n</details>\n"
			}

			fmt.Fprintf(&b, tmpl,
				markFailed, task.Name, step.Name, step.ExitCode, opts.TailLines, sanitizeCodeFence(step.LogTail),
			)
		}
	}

	return b.String()
}

func (f *Formatter) header(report *collector.Report) string {
	status := fmt.Sprintf("%s %s", markPassed, labelPassed)
	if !report.Succeeded {
		status = fmt.Sprintf("%s %s", markFailed, labelFailed)
	}

	name := fmt.Sprintf("`%s`", report.PipelineRunName)
	if url := f.links.PipelineRunURL(report.PipelineRunNamespace, report.PipelineRunName); url != "" {
		name = fmt.Sprintf("[`%s`](%s)", report.PipelineRunName, url)
	}

	return fmt.Sprintf("## Pipeline %s — %s", name, status)
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
