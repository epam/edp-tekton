package retry

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		err        error
		want       bool
	}{
		{"transport error", 0, errors.New("connection refused"), true},
		{"context canceled", 0, context.Canceled, false},
		{"context deadline", 0, context.DeadlineExceeded, false},
		{"success", http.StatusCreated, nil, false},
		{"bad request", http.StatusBadRequest, nil, false},
		{"unauthorized", http.StatusUnauthorized, nil, false},
		{"too many requests", http.StatusTooManyRequests, nil, true},
		{"internal server error", http.StatusInternalServerError, nil, true},
		{"bad gateway", http.StatusBadGateway, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, Transient(tt.statusCode, tt.err))
		})
	}
}

func TestDoRetriesTransientThenSucceeds(t *testing.T) {
	t.Parallel()

	calls := 0

	err := Do(context.Background(), func() error {
		calls++
		if calls == 1 {
			return errors.New("transient")
		}

		return nil
	}, func(error) bool { return true })
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
}

func TestDoStopsOnPermanentError(t *testing.T) {
	t.Parallel()

	calls := 0
	permanent := errors.New("permanent")

	err := Do(context.Background(), func() error {
		calls++

		return permanent
	}, func(error) bool { return false })
	require.ErrorIs(t, err, permanent)
	assert.Equal(t, 1, calls)
}

func TestDoGivesUpAfterAttempts(t *testing.T) {
	t.Parallel()

	calls := 0
	transientErr := errors.New("transient")

	err := Do(context.Background(), func() error {
		calls++

		return transientErr
	}, func(error) bool { return true })
	require.ErrorIs(t, err, transientErr)
	assert.Equal(t, attempts, calls)
}

func TestDoStopsWhenContextIsDone(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	transientErr := errors.New("transient")

	err := Do(ctx, func() error {
		calls++

		return transientErr
	}, func(error) bool { return true })
	require.ErrorIs(t, err, transientErr)
	assert.Equal(t, 1, calls)
}
