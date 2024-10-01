package bitbucket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/epam/edp-tekton/pkg/event_processor"
)

type EventProcessor struct {
	ksClient ctrlClient.Reader
	logger   *zap.SugaredLogger
}

func NewEventProcessor(ksClient ctrlClient.Reader, logger *zap.SugaredLogger) *EventProcessor {
	return &EventProcessor{
		ksClient: ksClient,
		logger:   logger,
	}
}

func (p *EventProcessor) Process(ctx context.Context, body []byte, ns, eventType string) (*event_processor.EventInfo, error) {
	switch eventType {
	case event_processor.BitbucketEventTypeCommentAdded:
		return p.processCommentEvent(ctx, body, ns)
	default:
		return p.processMergeEvent(ctx, body, ns)
	}
}

func (p *EventProcessor) processMergeEvent(ctx context.Context, body []byte, ns string) (*event_processor.EventInfo, error) {
	bitbucketEvent := &event_processor.BitbucketEvent{}
	if err := json.Unmarshal(body, bitbucketEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Bitbucket event: %w", err)
	}

	if bitbucketEvent.Repository.FullName == "" {
		return nil, errors.New("bitbucket repository path empty")
	}

	if bitbucketEvent.PullRequest.Destination.Branch.Name == "" {
		return nil, errors.New("bitbucket target branch empty")
	}

	repoPath := event_processor.ConvertRepositoryPath(bitbucketEvent.Repository.FullName)

	codebase, err := event_processor.GetCodebaseByRepoPath(ctx, p.ksClient, ns, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase by repo path: %w", err)
	}

	return &event_processor.EventInfo{
		GitProvider:  event_processor.GitProviderBitbucket,
		RepoPath:     repoPath,
		TargetBranch: bitbucketEvent.PullRequest.Destination.Branch.Name,
		Type:         event_processor.EventTypeMerge,
		Codebase:     codebase,
		PullRequest: &event_processor.PullRequest{
			HeadSha:           bitbucketEvent.PullRequest.Source.Commit.Hash,
			Title:             bitbucketEvent.PullRequest.Title,
			HeadRef:           bitbucketEvent.PullRequest.Source.Branch.Name,
			ChangeNumber:      bitbucketEvent.PullRequest.ID,
			LastCommitMessage: bitbucketEvent.PullRequest.LastCommit.Hash,
		},
	}, nil
}

func (p *EventProcessor) processCommentEvent(ctx context.Context, body []byte, ns string) (*event_processor.EventInfo, error) {
	bitbucketEvent := &event_processor.BitbucketCommentEvent{}
	if err := json.Unmarshal(body, bitbucketEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Bitbucket event: %w", err)
	}

	if bitbucketEvent.Repository.FullName == "" {
		return nil, errors.New("bitbucket repository path empty")
	}

	repoPath := event_processor.ConvertRepositoryPath(bitbucketEvent.Repository.FullName)

	codebase, err := event_processor.GetCodebaseByRepoPath(ctx, p.ksClient, ns, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase by repo path: %w", err)
	}

	return &event_processor.EventInfo{
		GitProvider:        event_processor.GitProviderBitbucket,
		RepoPath:           repoPath,
		TargetBranch:       bitbucketEvent.PullRequest.Destination.Branch.Name,
		Type:               event_processor.EventTypeReviewComment,
		Codebase:           codebase,
		HasPipelineRecheck: event_processor.ContainsPipelineRecheck(bitbucketEvent.Comment.Content.Raw),
		PullRequest: &event_processor.PullRequest{
			HeadSha:           bitbucketEvent.PullRequest.Source.Commit.Hash,
			Title:             bitbucketEvent.PullRequest.Title,
			HeadRef:           bitbucketEvent.PullRequest.Source.Branch.Name,
			ChangeNumber:      bitbucketEvent.PullRequest.ID,
			LastCommitMessage: bitbucketEvent.PullRequest.LastCommit.Hash,
		},
	}, nil
}
