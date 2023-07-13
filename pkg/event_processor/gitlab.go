package event_processor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type GitLabEventProcessor struct {
	ksClient ctrlClient.Reader
	logger   *zap.SugaredLogger
}

func NewGitLabEventProcessor(ksClient ctrlClient.Reader, logger *zap.SugaredLogger) *GitLabEventProcessor {
	return &GitLabEventProcessor{
		ksClient: ksClient,
		logger:   logger,
	}
}

func (p *GitLabEventProcessor) Process(ctx context.Context, body []byte, ns, eventType string) (*EventInfo, error) {
	switch eventType {
	case GitLabEventTypeCommentAdded:
		return p.processCommentEvent(ctx, body, ns)
	default:
		return p.processMergeEvent(ctx, body, ns)
	}
}

func (p *GitLabEventProcessor) processMergeEvent(ctx context.Context, body []byte, ns string) (*EventInfo, error) {
	gitLabEvent := &GitLabMergeRequestsEvent{}
	if err := json.Unmarshal(body, gitLabEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GitLab event: %w", err)
	}

	if gitLabEvent.Project.PathWithNamespace == "" {
		return nil, errors.New("gitlab repository path empty")
	}

	if gitLabEvent.ObjectAttributes.TargetBranch == "" {
		return nil, errors.New("gitlab target branch empty")
	}

	repoPath := convertRepositoryPath(gitLabEvent.Project.PathWithNamespace)

	codebase, err := getCodebaseByRepoPath(ctx, p.ksClient, ns, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase by repo path: %w", err)
	}

	return &EventInfo{
		GitProvider:  GitProviderGitLab,
		RepoPath:     repoPath,
		TargetBranch: gitLabEvent.ObjectAttributes.TargetBranch,
		Type:         EventTypeMerge,
		Codebase:     codebase,
		PullRequest: &PullRequest{
			HeadSha: gitLabEvent.ObjectAttributes.LastCommit.ID,
			Title:   gitLabEvent.ObjectAttributes.Title,
		},
	}, nil
}

func (p *GitLabEventProcessor) processCommentEvent(ctx context.Context, body []byte, ns string) (*EventInfo, error) {
	gitLabEvent := &GitLabCommentEvent{}
	if err := json.Unmarshal(body, gitLabEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GitLab event: %w", err)
	}

	if gitLabEvent.Project.PathWithNamespace == "" {
		return nil, errors.New("gitlab repository path empty")
	}

	repoPath := convertRepositoryPath(gitLabEvent.Project.PathWithNamespace)

	codebase, err := getCodebaseByRepoPath(ctx, p.ksClient, ns, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase by repo path: %w", err)
	}

	// The comment was not added to the merge request if the target branch was empty.
	if gitLabEvent.MergeRequest.TargetBranch == "" {
		return &EventInfo{
			GitProvider:        GitProviderGitLab,
			Type:               EventTypeReviewComment,
			Codebase:           codebase,
			HasPipelineRecheck: false,
		}, nil
	}

	return &EventInfo{
		GitProvider:        GitProviderGitLab,
		RepoPath:           repoPath,
		TargetBranch:       gitLabEvent.MergeRequest.TargetBranch,
		Type:               EventTypeReviewComment,
		Codebase:           codebase,
		HasPipelineRecheck: containsPipelineRecheck(gitLabEvent.ObjectAttributes.Note),
		PullRequest: &PullRequest{
			HeadSha: gitLabEvent.MergeRequest.LastCommit.ID,
			Title:   gitLabEvent.MergeRequest.Title,
		},
	}, nil
}
