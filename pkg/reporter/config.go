package reporter

import (
	"fmt"
	"os"
	"strconv"
)

const (
	// CommentStrategyUpdate finds the previous report comment by its hidden
	// marker and edits it in place.
	CommentStrategyUpdate = "update"

	// CommentStrategyNew always creates a new comment.
	CommentStrategyNew = "new"

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
	// CommentStrategy is either CommentStrategyUpdate or CommentStrategyNew.
	CommentStrategy string
	// PortalBaseURL is the base URL of the KubeRocketCI portal used to render
	// links to PipelineRun details. Links are omitted when empty.
	PortalBaseURL string
}

// LoadConfig reads the reporter configuration from environment variables.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Namespace:       os.Getenv("SYSTEM_NAMESPACE"),
		TailLines:       defaultTailLines,
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

	if v, ok := os.LookupEnv("REPORTER_COMMENT_STRATEGY"); ok {
		if v != CommentStrategyUpdate && v != CommentStrategyNew {
			return nil, fmt.Errorf("REPORTER_COMMENT_STRATEGY must be %q or %q, got %q",
				CommentStrategyUpdate, CommentStrategyNew, v)
		}

		cfg.CommentStrategy = v
	}

	return cfg, nil
}
