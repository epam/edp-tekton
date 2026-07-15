// Package retry holds the shared transient-failure retry policy for git
// provider API calls. Only failures that can heal on their own are retried
// (transport errors, 429, 5xx); permanent API errors such as GitLab commit
// status transition conflicts must be handled by the caller, never retried.
package retry

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	// attempts is the total number of tries for one API call. Kept small on
	// purpose: the interceptor posts statuses on the webhook hot path and the
	// request context bounds the whole call chain.
	attempts = 3

	baseDelay = 200 * time.Millisecond
	maxDelay  = time.Second
)

// Transient reports whether an API call failure is worth retrying.
// statusCode must be 0 when no HTTP response was received (transport error).
func Transient(statusCode int, err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	if statusCode == 0 {
		return err != nil
	}

	return statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError
}

// ConfigureResty enables the retry policy on a resty client. The request
// context still bounds the total time spent, including backoff waits.
func ConfigureResty(client *resty.Client) *resty.Client {
	return client.
		SetRetryCount(attempts - 1).
		SetRetryWaitTime(baseDelay).
		SetRetryMaxWaitTime(maxDelay).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			// resty can invoke conditions with a nil response when the request
			// fails before any HTTP attempt is made.
			if r == nil {
				return Transient(0, err)
			}

			return Transient(r.StatusCode(), err)
		})
}

// Do runs call up to the policy's attempt count, backing off exponentially
// between transient failures; transient classifies the returned error. The
// last error is returned unchanged so callers keep their own wrapping. It is
// meant for SDK clients (e.g. go-github) that cannot use ConfigureResty.
func Do(ctx context.Context, call func() error, transient func(error) bool) error {
	delay := baseDelay

	for attempt := 1; ; attempt++ {
		err := call()
		if err == nil || attempt == attempts || !transient(err) {
			return err
		}

		select {
		case <-ctx.Done():
			return err
		case <-time.After(delay):
		}

		delay = min(delay*2, maxDelay)
	}
}
