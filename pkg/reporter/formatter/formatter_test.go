package formatter

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/epam/edp-tekton/pkg/reporter/collector"
)

func TestFormat(t *testing.T) {
	t.Parallel()

	report := &collector.Report{
		PipelineRunName:      "review-my-app-xyz",
		PipelineRunNamespace: "krci",
		Succeeded:            false,
		Tasks: []collector.TaskResult{
			{
				Name:      "build",
				Succeeded: false,
				Duration:  92 * time.Second,
				StartTime: time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC),
				Steps: []collector.StepResult{
					{Name: "npm-build", Container: "step-npm-build", ExitCode: 1, LogTail: "npm ERR! build failed"},
					{Name: "skipped-after", Container: "step-skipped-after", ExitCode: 1, LogTail: ""},
				},
			},
			{
				Name:      "fetch-repository",
				Succeeded: true,
				Duration:  5 * time.Second,
				StartTime: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Steps:     []collector.StepResult{{Name: "clone", Succeeded: true}},
			},
		},
	}

	body := New(PortalLinkBuilder{BaseURL: "https://portal.example.com/c/main/cicd/pipelineruns"}).
		Format(report, "<!-- marker -->", Options{TailLines: 100, CollapsibleSections: true})

	require.True(t, strings.HasPrefix(body, "<!-- marker -->\n"), "comment must start with the marker")
	assert.Contains(t, body, "❌ Failed")
	assert.Contains(t, body,
		"[`review-my-app-xyz`](https://portal.example.com/c/main/cicd/pipelineruns/krci/review-my-app-xyz)")

	// Tasks are ordered by start time: fetch-repository ran first.
	fetchIdx := strings.Index(body, "| ✅ | fetch-repository | 5s |")
	buildIdx := strings.Index(body, "| ❌ | build | 1m32s |")

	require.GreaterOrEqual(t, fetchIdx, 0)
	require.GreaterOrEqual(t, buildIdx, 0)
	assert.Less(t, fetchIdx, buildIdx)

	assert.Contains(t, body, "<details><summary>❌ <b>build</b> / npm-build (exit code 1, last 100 log lines)</summary>")
	assert.Contains(t, body, "npm ERR! build failed")
	// Steps without a log tail (cascading skips) get no details section.
	assert.NotContains(t, body, "skipped-after")
	// Successful tasks get no details section.
	assert.Equal(t, 1, strings.Count(body, "<details>"))
}

func TestFormatWithoutCollapsibleSections(t *testing.T) {
	t.Parallel()

	report := &collector.Report{
		PipelineRunName: "review-bb-app",
		Succeeded:       false,
		Tasks: []collector.TaskResult{{
			Name:      "test",
			Succeeded: false,
			Steps: []collector.StepResult{
				{Name: "mvn-goals", ExitCode: 1, LogTail: "BUILD FAILURE"},
			},
		}},
	}

	body := New(PortalLinkBuilder{}).Format(report, "<!-- marker -->", Options{TailLines: 100})

	// Renderers that escape embedded HTML (Bitbucket Cloud) must get pure markdown.
	assert.NotContains(t, body, "<details>")
	assert.NotContains(t, body, "<summary>")
	assert.NotContains(t, body, "<b>")
	assert.Contains(t, body, "❌ **test / mvn-goals** (exit code 1, last 100 log lines)")
	assert.Contains(t, body, "```\nBUILD FAILURE\n```")
}

func TestFormatSucceededWithoutLinks(t *testing.T) {
	t.Parallel()

	report := &collector.Report{
		PipelineRunName: "review-ok",
		Succeeded:       true,
		Tasks:           []collector.TaskResult{{Name: "build", Succeeded: true, Duration: time.Second}},
	}

	body := New(PortalLinkBuilder{}).Format(report, "<!-- marker -->", Options{TailLines: 100, CollapsibleSections: true})

	assert.Contains(t, body, "## Pipeline `review-ok` ✅ Passed")
	assert.NotContains(t, body, "<details>")
	assert.NotContains(t, body, "](")
}

func TestFormatEscapesCodeFences(t *testing.T) {
	t.Parallel()

	report := &collector.Report{
		PipelineRunName: "review-x",
		Tasks: []collector.TaskResult{{
			Name: "build",
			Steps: []collector.StepResult{
				{Name: "s", ExitCode: 1, LogTail: "before\n```\ninjected"},
			},
		}},
	}

	body := New(PortalLinkBuilder{}).Format(report, "<!-- m -->", Options{TailLines: 10, CollapsibleSections: true})

	assert.NotContains(t, body, "```\ninjected", "log content must not close the code fence")
}

func TestFormatEscapesLongBacktickRuns(t *testing.T) {
	t.Parallel()

	report := &collector.Report{
		PipelineRunName: "review-x",
		Tasks: []collector.TaskResult{{
			Name: "build",
			Steps: []collector.StepResult{
				{Name: "s", ExitCode: 1, LogTail: "before\n`````\ninjected"},
			},
		}},
	}

	body := New(PortalLinkBuilder{}).Format(report, "<!-- m -->", Options{TailLines: 10, CollapsibleSections: true})

	// Strip the two legitimate fence delimiters, then no run of 3+ backticks may remain.
	inner := strings.SplitN(body, "```\n", 3)
	require.Len(t, inner, 3)
	assert.NotContains(t, inner[2][:strings.Index(inner[2], "</details>")], "```",
		"a 5-backtick run must not survive sanitization as a fence")
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	t.Run("short body is unchanged", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "short", Truncate("short", 100))
	})

	t.Run("long body is cut with notice within budget", func(t *testing.T) {
		t.Parallel()

		body := strings.Repeat("a", 200)
		got := Truncate(body, 100)

		assert.LessOrEqual(t, len(got), 100)
		assert.True(t, strings.HasSuffix(got, "… (truncated)"))
	})

	t.Run("multi-byte runes are not split", func(t *testing.T) {
		t.Parallel()

		body := strings.Repeat("é", 200)

		for maxBytes := 20; maxBytes < 30; maxBytes++ {
			got := Truncate(body, maxBytes)

			assert.LessOrEqual(t, len(got), maxBytes)
			assert.True(t, strings.HasSuffix(got, "… (truncated)"))

			for _, r := range got {
				assert.NotEqual(t, '�', r, "truncation produced an invalid rune at maxBytes=%d", maxBytes)
			}
		}
	})
}
