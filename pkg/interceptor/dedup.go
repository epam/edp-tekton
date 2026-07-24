package interceptor

import (
	"fmt"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"

	"github.com/epam/edp-tekton/pkg/event_processor"
)

const (
	reviewedHeadsTTL      = 2 * time.Hour
	reviewedHeadsCapacity = 1024
)

// reviewDeduper skips Bitbucket pullrequest:updated events that carry no new commit.
// Bitbucket fires the same event for code pushes and metadata-only edits, so a head SHA
// already recorded for the same review means the update carries nothing new. Best-effort
// and fail-open: a miss (restart, eviction, >1 replica) triggers a harmless extra review.
type reviewDeduper struct {
	heads *expirable.LRU[string, struct{}]
}

func newReviewDeduper() *reviewDeduper {
	return &reviewDeduper{
		heads: expirable.NewLRU[string, struct{}](reviewedHeadsCapacity, nil, reviewedHeadsTTL),
	}
}

func (d *reviewDeduper) alreadyReviewed(ns string, e *event_processor.EventInfo) bool {
	return e.IsPullRequestUpdateEvent() && applies(e) && d.heads.Contains(reviewHeadKey(ns, e))
}

// record excludes comment events (/recheck, /ok-to-test): an explicit command must always
// trigger, so its head is never recorded and can never be suppressed by a later update.
func (d *reviewDeduper) record(ns string, e *event_processor.EventInfo) {
	if applies(e) && !e.IsReviewCommentEvent() {
		d.heads.Add(reviewHeadKey(ns, e), struct{}{})
	}
}

func applies(e *event_processor.EventInfo) bool {
	return e.GitProvider == event_processor.GitProviderBitbucket &&
		e.PullRequest != nil && e.PullRequest.HeadSha != "" && e.PullRequest.ChangeNumber > 0
}

// reviewHeadKey scopes a reviewed head by namespace, codebase, target branch and change
// number, so an identical commit in another tenant, repo, target branch or PR is never
// mistaken for an already-reviewed one, and a PR retarget re-triggers even with the same head.
func reviewHeadKey(ns string, e *event_processor.EventInfo) string {
	pr := e.PullRequest

	return fmt.Sprintf("%s/%s/%s/%d/%s", ns, e.Codebase.Name, e.TargetBranch, pr.ChangeNumber, pr.HeadSha)
}
