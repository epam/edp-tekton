package event_processor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// GerritEventProcessor is an implementation of Processor for Gerrit.
type GerritEventProcessor struct {
	ksClient ctrlClient.Reader
	logger   *zap.SugaredLogger
}

func NewGerritEventProcessor(ksClient ctrlClient.Reader, logger *zap.SugaredLogger) *GerritEventProcessor {
	return &GerritEventProcessor{
		ksClient: ksClient,
		logger:   logger,
	}
}

func (p *GerritEventProcessor) Process(ctx context.Context, body []byte, ns, _ string) (*EventInfo, error) {
	gerritEvent := &GerritEvent{}
	if err := json.Unmarshal(body, gerritEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gerrit event: %w", err)
	}

	if gerritEvent.Project.Name == "" {
		return nil, errors.New("gerrit repository path empty")
	}

	if gerritEvent.Change.Branch == "" {
		return nil, errors.New("gerrit target branch empty")
	}

	repoPath := convertRepositoryPath(gerritEvent.Project.Name)

	codebase, err := getCodebaseByRepoPath(ctx, p.ksClient, ns, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase: %w", err)
	}

	event := &EventInfo{
		GitProvider:  GitProviderGerrit,
		RepoPath:     strings.ToLower(gerritEvent.Project.Name),
		TargetBranch: gerritEvent.Change.Branch,
		Type:         EventTypeMerge,
		Codebase:     codebase,
	}

	if gerritEvent.Type == GerritEventTypeCommentAdded {
		event.Type = EventTypeReviewComment
		event.HasPipelineRecheck = containsPipelineRecheck(gerritEvent.Comment)
	}

	return event, nil
}
