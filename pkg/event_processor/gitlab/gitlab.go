package gitlab

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
	case event_processor.GitLabEventTypeCommentAdded:
		return p.processCommentEvent(ctx, body, ns)
	default:
		return p.processMergeEvent(ctx, body, ns)
	}
}

func (p *EventProcessor) processMergeEvent(ctx context.Context, body []byte, ns string) (*event_processor.EventInfo, error) {
	gitLabEvent := &event_processor.GitLabMergeRequestsEvent{}
	if err := json.Unmarshal(body, gitLabEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GitLab event: %w", err)
	}

	if gitLabEvent.Project.PathWithNamespace == "" {
		return nil, errors.New("gitlab repository path empty")
	}

	if gitLabEvent.ObjectAttributes.TargetBranch == "" {
		return nil, errors.New("gitlab target branch empty")
	}

	repoPath := event_processor.ConvertRepositoryPath(gitLabEvent.Project.PathWithNamespace)

	codebase, err := event_processor.GetCodebaseByRepoPath(ctx, p.ksClient, ns, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase by repo path: %w", err)
	}

	return &event_processor.EventInfo{
		GitProvider:  event_processor.GitProviderGitLab,
		RepoPath:     repoPath,
		TargetBranch: gitLabEvent.ObjectAttributes.TargetBranch,
		Type:         event_processor.EventTypeMerge,
		Codebase:     codebase,
		PullRequest: &event_processor.PullRequest{
			HeadSha:           gitLabEvent.ObjectAttributes.LastCommit.ID,
			Title:             gitLabEvent.ObjectAttributes.Title,
			HeadRef:           gitLabEvent.ObjectAttributes.SourceBranch,
			ChangeNumber:      gitLabEvent.ObjectAttributes.ChangeNumber,
			LastCommitMessage: gitLabEvent.ObjectAttributes.LastCommit.Message,
			Author:            gitLabEvent.User.Username,
			AuthorAvatarUrl:   gitLabEvent.User.AvatarUrl,
		},
	}, nil
}

func (p *EventProcessor) processCommentEvent(ctx context.Context, body []byte, ns string) (*event_processor.EventInfo, error) {
	gitLabEvent := &event_processor.GitLabCommentEvent{}
	if err := json.Unmarshal(body, gitLabEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GitLab event: %w", err)
	}

	if gitLabEvent.Project.PathWithNamespace == "" {
		return nil, errors.New("gitlab repository path empty")
	}

	repoPath := event_processor.ConvertRepositoryPath(gitLabEvent.Project.PathWithNamespace)

	codebase, err := event_processor.GetCodebaseByRepoPath(ctx, p.ksClient, ns, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase by repo path: %w", err)
	}

	// The comment was added not to the merge request if TargetBranch is empty.
	// Skip processing such events.
	if gitLabEvent.MergeRequest.TargetBranch == "" {
		return &event_processor.EventInfo{
			GitProvider:        event_processor.GitProviderGitLab,
			Type:               event_processor.EventTypeReviewComment,
			Codebase:           codebase,
			HasPipelineRecheck: false,
		}, nil
	}

	return &event_processor.EventInfo{
		GitProvider:        event_processor.GitProviderGitLab,
		RepoPath:           repoPath,
		TargetBranch:       gitLabEvent.MergeRequest.TargetBranch,
		Type:               event_processor.EventTypeReviewComment,
		Codebase:           codebase,
		HasPipelineRecheck: event_processor.ContainsPipelineRecheckPrefix(gitLabEvent.ObjectAttributes.Note),
		PullRequest: &event_processor.PullRequest{
			HeadSha:           gitLabEvent.MergeRequest.LastCommit.ID,
			Title:             gitLabEvent.MergeRequest.Title,
			HeadRef:           gitLabEvent.MergeRequest.SourceBranch,
			ChangeNumber:      gitLabEvent.MergeRequest.ChangeNumber,
			LastCommitMessage: gitLabEvent.MergeRequest.LastCommit.Message,
			Author:            gitLabEvent.User.Username,
			AuthorAvatarUrl:   gitLabEvent.User.AvatarUrl,
		},
	}, nil
}
