package gerrit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/epam/edp-tekton/pkg/event_processor"
)

// EventProcessor is an implementation of Processor for Gerrit.
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

func (p *EventProcessor) Process(ctx context.Context, body []byte, ns, _ string) (*event_processor.EventInfo, error) {
	gerritEvent := &event_processor.GerritEvent{}
	if err := json.Unmarshal(body, gerritEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gerrit event: %w", err)
	}

	if gerritEvent.Project.Name == "" {
		return nil, errors.New("gerrit repository path empty")
	}

	if gerritEvent.Change.Branch == "" {
		return nil, errors.New("gerrit target branch empty")
	}

	repoPath := event_processor.ConvertRepositoryPath(gerritEvent.Project.Name)

	codebase, err := event_processor.GetCodebaseByRepoPath(ctx, p.ksClient, ns, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase: %w", err)
	}

	event := &event_processor.EventInfo{
		GitProvider:  event_processor.GitProviderGerrit,
		RepoPath:     strings.ToLower(gerritEvent.Project.Name),
		TargetBranch: gerritEvent.Change.Branch,
		Type:         event_processor.EventTypeMerge,
		Codebase:     codebase,
	}

	if gerritEvent.Type == event_processor.GerritEventTypeCommentAdded {
		event.Type = event_processor.EventTypeReviewComment
		event.HasPipelineRecheck = event_processor.ContainsPipelineRecheck(gerritEvent.Comment)
	}

	return event, nil
}
