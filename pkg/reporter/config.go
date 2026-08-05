package reporter

import (
	"fmt"
	"os"
	"strconv"

	"github.com/epam/edp-tekton/pkg/reporter/provider/types"
)

const (
	// CommentStrategyUpdate finds the previous report comment by its hidden
	// marker and edits it in place.
	CommentStrategyUpdate = types.CommentStrategyUpdate

	// CommentStrategyNew always creates a new comment.
	CommentStrategyNew = types.CommentStrategyNew

	// CommentStrategyRecreate creates a new comment at the bottom of the
	// thread and deletes the previous report comments.
	CommentStrategyRecreate = types.CommentStrategyRecreate

	// MaxCommentBytes is the comment body size cap. GitHub allows 65536
	// characters per comment and it is the strictest of the supported providers.
	MaxCommentBytes = 65536

	defaultTailLines = 100
)

// Config holds the reporter runtime configuration sourced from environment variables.
type Config struct {
	// Namespace limits the PipelineRun watch to a single namespace. Empty means cluster-wide.
	Namespace string
	// TailLines is the number of trailing log lines fetched for every failed step.
	TailLines int64
	// LogsEnabled controls whether trailing log lines of failed steps are
	// published in the pull request comment. Disabled by default: pipeline
	// output may carry secrets, and not republishing it to the VCS is the
	// point.
	LogsEnabled bool
	// CommentStrategy is one of CommentStrategyUpdate, CommentStrategyNew or
	// CommentStrategyRecreate.
	CommentStrategy types.CommentStrategy
	// PortalBaseURL is the base URL of the KubeRocketCI portal used to render
	// links to PipelineRun details. Links are omitted when empty.
	PortalBaseURL string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Namespace:       os.Getenv("SYSTEM_NAMESPACE"),
		TailLines:       defaultTailLines,
		LogsEnabled:     false,
		CommentStrategy: CommentStrategyUpdate,
		PortalBaseURL:   os.Getenv("PORTAL_BASE_URL"),
	}

	if v, ok := os.LookupEnv("REPORTER_TAIL_LINES"); ok {
		lines, err := strconv.ParseInt(v, 10, 64)
		if err != nil || lines <= 0 {
			return nil, fmt.Errorf("REPORTER_TAIL_LINES must be a positive integer, got %q", v)
		}

		cfg.TailLines = lines
	}

	if v, ok := os.LookupEnv("REPORTER_LOGS_ENABLED"); ok {
		enabled, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("REPORTER_LOGS_ENABLED must be a boolean, got %q", v)
		}

		cfg.LogsEnabled = enabled
	}

	if v, ok := os.LookupEnv("REPORTER_COMMENT_STRATEGY"); ok {
		strategy := types.CommentStrategy(v)
		if strategy != CommentStrategyUpdate && strategy != CommentStrategyNew && strategy != CommentStrategyRecreate {
			return nil, fmt.Errorf("REPORTER_COMMENT_STRATEGY must be %q, %q or %q, got %q",
				CommentStrategyUpdate, CommentStrategyNew, CommentStrategyRecreate, v)
		}

		cfg.CommentStrategy = strategy
	}

	return cfg, nil
}
