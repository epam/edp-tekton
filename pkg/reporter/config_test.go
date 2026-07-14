package reporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigDefaults(t *testing.T) {
	// t.Setenv also clears the var for the duration of the test if we set it empty,
	// but to assert defaults we rely on these vars being unset in the test env.
	t.Setenv("SYSTEM_NAMESPACE", "")
	t.Setenv("PORTAL_BASE_URL", "")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, int64(defaultTailLines), cfg.TailLines)
	assert.Equal(t, CommentStrategyUpdate, cfg.CommentStrategy)
	assert.Equal(t, MaxCommentBytes, 65536)
	assert.Empty(t, cfg.Namespace)
	assert.Empty(t, cfg.PortalBaseURL)
}

func TestLoadConfigOverrides(t *testing.T) {
	t.Setenv("SYSTEM_NAMESPACE", "krci")
	t.Setenv("PORTAL_BASE_URL", "https://portal.example.com/c/cluster/cicd/pipelineruns")
	t.Setenv("REPORTER_TAIL_LINES", "50")
	t.Setenv("REPORTER_COMMENT_STRATEGY", CommentStrategyNew)

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "krci", cfg.Namespace)
	assert.Equal(t, "https://portal.example.com/c/cluster/cicd/pipelineruns", cfg.PortalBaseURL)
	assert.Equal(t, int64(50), cfg.TailLines)
	assert.Equal(t, CommentStrategyNew, cfg.CommentStrategy)
}

func TestLoadConfigInvalidTailLines(t *testing.T) {
	// Cannot use t.Parallel here: LoadConfig reads process env via t.Setenv,
	// which panics if the test is marked parallel.
	tests := []struct {
		name  string
		value string
	}{
		{name: "not a number", value: "abc"},
		{name: "zero", value: "0"},
		{name: "negative", value: "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("REPORTER_TAIL_LINES", tt.value)

			_, err := LoadConfig()
			assert.ErrorContains(t, err, "REPORTER_TAIL_LINES must be a positive integer")
		})
	}
}

func TestLoadConfigInvalidCommentStrategy(t *testing.T) {
	t.Setenv("REPORTER_COMMENT_STRATEGY", "bogus")

	_, err := LoadConfig()
	assert.ErrorContains(t, err, "REPORTER_COMMENT_STRATEGY must be")
}
