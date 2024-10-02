package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-github/v31/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/epam/edp-tekton/pkg/event_processor"
)

type EventProcessor struct {
	ksClient     ctrlClient.Reader
	logger       *zap.SugaredLogger
	gitHubClient func(ctx context.Context, token string) *github.Client
}

type EventProcessorOptions struct {
	Logger       *zap.SugaredLogger
	GitHubClient func(ctx context.Context, token string) *github.Client
}

func NewEventProcessor(
	ksClient ctrlClient.Reader,
	options *EventProcessorOptions,
) *EventProcessor {
	if options == nil {
		options = &EventProcessorOptions{}
	}

	if options.Logger == nil {
		options.Logger = zap.NewNop().Sugar()
	}

	if options.GitHubClient == nil {
		options.GitHubClient = func(ctx context.Context, token string) *github.Client {
			return github.NewClient(
				oauth2.NewClient(
					ctx,
					oauth2.StaticTokenSource(
						&oauth2.Token{AccessToken: token},
					),
				),
			)
		}
	}

	return &EventProcessor{
		ksClient:     ksClient,
		logger:       options.Logger,
		gitHubClient: options.GitHubClient,
	}
}

func (p *EventProcessor) Process(ctx context.Context, body []byte, ns, eventType string) (*event_processor.EventInfo, error) {
	switch eventType {
	case event_processor.GitHubEventTypeCommentAdded:
		return p.processCommentEvent(ctx, body, ns)
	default:
		return p.processMergeEvent(ctx, body, ns)
	}
}

// processCommentEvent processes GitHub comment event.
// nolint:cyclop // function is not complex
func (p *EventProcessor) processMergeEvent(ctx context.Context, body []byte, ns string) (*event_processor.EventInfo, error) {
	gitHubEvent := &github.PullRequestEvent{}
	if err := json.Unmarshal(body, gitHubEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GitHub event: %w", err)
	}

	if gitHubEvent.Repo == nil || gitHubEvent.Repo.FullName == nil || *gitHubEvent.Repo.FullName == "" {
		return nil, errors.New("github repository path empty")
	}

	if gitHubEvent.PullRequest == nil ||
		gitHubEvent.PullRequest.Base == nil ||
		gitHubEvent.PullRequest.Base.Ref == nil ||
		*gitHubEvent.PullRequest.Base.Ref == "" {
		return nil, errors.New("github target branch empty")
	}

	if gitHubEvent.GetPullRequest().GetHead() == nil {
		return nil, errors.New("github head branch empty")
	}

	repoPath := event_processor.ConvertRepositoryPath(*gitHubEvent.Repo.FullName)

	codebase, err := event_processor.GetCodebaseByRepoPath(ctx, p.ksClient, ns, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase %s for GitHub IssueCommentEvent: %w", repoPath, err)
	}

	gitServerToken, err := event_processor.GetGitServerToken(ctx, p.ksClient, codebase)
	if err != nil {
		return nil, fmt.Errorf("failed to get git server token for GitHub PullRequestEvent: %w", err)
	}

	client := p.gitHubClient(ctx, gitServerToken)

	commitMessage, err := p.getCommitMessage(
		ctx,
		client,
		gitHubEvent.GetRepo().GetOwner().GetLogin(),
		gitHubEvent.GetRepo().GetName(),
		gitHubEvent.GetNumber(),
	)
	if err != nil {
		return nil, err
	}

	return &event_processor.EventInfo{
		GitProvider:  event_processor.GitProviderGitHub,
		RepoPath:     repoPath,
		TargetBranch: *gitHubEvent.PullRequest.Base.Ref,
		Codebase:     codebase,
		Type:         event_processor.EventTypeMerge,
		PullRequest: &event_processor.PullRequest{
			HeadRef:           gitHubEvent.GetPullRequest().GetHead().GetRef(),
			HeadSha:           gitHubEvent.GetPullRequest().GetHead().GetSHA(),
			Title:             gitHubEvent.GetPullRequest().GetTitle(),
			ChangeNumber:      gitHubEvent.GetNumber(),
			LastCommitMessage: commitMessage,
		},
	}, nil
}

// processCommentEvent processes GitHub comment event.
// nolint:funlen // function is not so complex
func (p *EventProcessor) processCommentEvent(ctx context.Context, body []byte, ns string) (*event_processor.EventInfo, error) {
	event := &github.IssueCommentEvent{}
	if err := json.Unmarshal(body, event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GitHub IssueCommentEvent: %w", err)
	}

	if event.GetAction() != "created" {
		return createEventInfoWithoutRecheck(), nil
	}

	repoPath := event_processor.ConvertRepositoryPath(event.GetRepo().GetFullName())

	codebase, err := event_processor.GetCodebaseByRepoPath(ctx, p.ksClient, ns, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase %s for GitHub IssueCommentEvent: %w", repoPath, err)
	}

	gitServerToken, err := event_processor.GetGitServerToken(ctx, p.ksClient, codebase)
	if err != nil {
		return nil, fmt.Errorf("failed to get git server token for GitHub IssueCommentEvent: %w", err)
	}

	client := p.gitHubClient(ctx, gitServerToken)

	pullReq, _, err := client.PullRequests.Get(
		ctx,
		event.GetRepo().GetOwner().GetLogin(),
		event.GetRepo().GetName(),
		event.GetIssue().GetNumber(),
	)
	if err != nil {
		if isPRNotFoundErr(err) {
			p.logger.Info("GitHub pull request not found")

			return createEventInfoWithoutRecheck(), nil
		}

		return nil, fmt.Errorf("failed to get GitHub pull request: %w", err)
	}

	if pullReq.GetBase() == nil {
		return nil, errors.New("github target branch empty")
	}

	if pullReq.GetHead() == nil {
		return nil, errors.New("github head branch empty")
	}

	commitMessage, err := p.getCommitMessage(
		ctx,
		client,
		event.GetRepo().GetOwner().GetLogin(),
		event.GetRepo().GetName(),
		event.GetIssue().GetNumber(),
	)
	if err != nil {
		return nil, err
	}

	return &event_processor.EventInfo{
		GitProvider:        event_processor.GitProviderGitHub,
		RepoPath:           repoPath,
		TargetBranch:       pullReq.GetBase().GetRef(),
		Type:               event_processor.EventTypeReviewComment,
		HasPipelineRecheck: event_processor.ContainsPipelineRecheck(event.GetComment().GetBody()),
		Codebase:           codebase,
		PullRequest: &event_processor.PullRequest{
			HeadRef:           pullReq.GetHead().GetRef(),
			HeadSha:           pullReq.GetHead().GetSHA(),
			Title:             pullReq.GetTitle(),
			ChangeNumber:      pullReq.GetNumber(),
			LastCommitMessage: commitMessage,
		},
	}, nil
}

func (*EventProcessor) getCommitMessage(
	ctx context.Context,
	client *github.Client,
	owner string,
	repo string,
	number int,
) (string, error) {
	commits, _, err := client.PullRequests.ListCommits(ctx, owner, repo, number, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub pull request commits: %w", err)
	}

	if len(commits) == 0 {
		return "", errors.New("github pull request commits empty")
	}

	m := commits[len(commits)-1].Commit.Message

	if m == nil {
		return "", nil
	}

	return *m, nil
}

func createEventInfoWithoutRecheck() *event_processor.EventInfo {
	return &event_processor.EventInfo{
		GitProvider:        event_processor.GitProviderGitHub,
		Type:               event_processor.EventTypeReviewComment,
		HasPipelineRecheck: false,
	}
}

func isPRNotFoundErr(err error) bool {
	var responseErr *github.ErrorResponse
	if errors.As(err, &responseErr) {
		if responseErr.Response.StatusCode == http.StatusNotFound {
			return true
		}
	}

	return false
}
