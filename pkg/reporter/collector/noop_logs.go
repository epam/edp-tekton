package collector

import "context"

// NoopLogFetcher never contacts the pods/log API. It is wired in when log
// snippet reporting is disabled, so the reporter needs no pods/log RBAC and
// no code path can fall through to an unbounded log read.
type NoopLogFetcher struct{}

// GetLogs implements LogFetcher without reading anything; it always returns an empty log.
func (NoopLogFetcher) GetLogs(_ context.Context, _, _, _ string, _ int64) (string, error) {
	return "", nil
}
