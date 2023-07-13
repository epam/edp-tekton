package event_processor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-github/v31/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/pkg/apis/edp/v1"
)

const gitServerTokenField = "token"

type GitHubEventProcessor struct {
	ksClient     ctrlClient.Reader
	logger       *zap.SugaredLogger
	gitHubClient func(ctx context.Context, token string) *github.Client
}

type GitHubEventProcessorOptions struct {
	Logger       *zap.SugaredLogger
	GitHubClient func(ctx context.Context, token string) *github.Client
}

func NewGitHubEventProcessor(
	ksClient ctrlClient.Reader,
	options *GitHubEventProcessorOptions,
) *GitHubEventProcessor {
	if options == nil {
		options = &GitHubEventProcessorOptions{}
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

	return &GitHubEventProcessor{
		ksClient:     ksClient,
		logger:       options.Logger,
		gitHubClient: options.GitHubClient,
	}
}

func (p *GitHubEventProcessor) Process(ctx context.Context, body []byte, ns, eventType string) (*EventInfo, error) {
	switch eventType {
	case GitHubEventTypeCommentAdded:
		return p.processCommentEvent(ctx, body, ns)
	default:
		return p.processMergeEvent(ctx, body, ns)
	}
}

// processCommentEvent processes GitHub comment event.
// nolint:cyclop // function is not complex
func (p *GitHubEventProcessor) processMergeEvent(ctx context.Context, body []byte, ns string) (*EventInfo, error) {
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

	repoPath := convertRepositoryPath(*gitHubEvent.Repo.FullName)

	codebase, err := getCodebaseByRepoPath(ctx, p.ksClient, ns, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase %s for GitHub IssueCommentEvent: %w", repoPath, err)
	}

	return &EventInfo{
		GitProvider:  GitProviderGitHub,
		RepoPath:     repoPath,
		TargetBranch: *gitHubEvent.PullRequest.Base.Ref,
		Codebase:     codebase,
		Type:         EventTypeMerge,
		PullRequest: &PullRequest{
			HeadRef: gitHubEvent.GetPullRequest().GetHead().GetRef(),
			HeadSha: gitHubEvent.GetPullRequest().GetHead().GetSHA(),
			Title:   gitHubEvent.GetPullRequest().GetTitle(),
		},
	}, nil
}

// getCodebaseByRepoPath returns codebase by repository path.
// nolint:funlen // function is not so complex
func (p *GitHubEventProcessor) processCommentEvent(ctx context.Context, body []byte, ns string) (*EventInfo, error) {
	event := new(github.IssueCommentEvent)
	if err := json.Unmarshal(body, event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GitHub IssueCommentEvent: %w", err)
	}

	if event.GetAction() != "created" {
		return &EventInfo{
			GitProvider:        GitProviderGitHub,
			Type:               EventTypeReviewComment,
			HasPipelineRecheck: false,
		}, nil
	}

	repoPath := convertRepositoryPath(event.GetRepo().GetFullName())

	codebase, err := getCodebaseByRepoPath(ctx, p.ksClient, ns, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase %s for GitHub IssueCommentEvent: %w", repoPath, err)
	}

	gitServerToken, err := p.getGitServerToken(ctx, codebase)
	if err != nil {
		return nil, fmt.Errorf("failed to get git server token for GitHub IssueCommentEvent: %w", err)
	}

	client := p.gitHubClient(ctx, gitServerToken)

	pullReq, _, err := client.PullRequests.Get(
		context.Background(),
		event.GetRepo().GetOwner().GetLogin(),
		event.GetRepo().GetName(),
		event.GetIssue().GetNumber(),
	)
	if err != nil {
		var responseErr *github.ErrorResponse
		if errors.As(err, &responseErr) {
			p.logger.Info("GitHub pull request not found")

			if responseErr.Response.StatusCode == http.StatusNotFound {
				return &EventInfo{
					GitProvider:        GitProviderGitHub,
					Type:               EventTypeReviewComment,
					HasPipelineRecheck: false,
				}, nil
			}
		}

		return nil, fmt.Errorf("failed to get GitHub pull request: %w", err)
	}

	if pullReq.GetBase() == nil {
		return nil, errors.New("github target branch empty")
	}

	if pullReq.GetHead() == nil {
		return nil, errors.New("github head branch empty")
	}

	return &EventInfo{
		GitProvider:        GitProviderGitHub,
		RepoPath:           repoPath,
		TargetBranch:       pullReq.GetBase().GetRef(),
		Type:               EventTypeReviewComment,
		HasPipelineRecheck: containsPipelineRecheck(event.GetComment().GetBody()),
		Codebase:           codebase,
		PullRequest: &PullRequest{
			HeadRef: pullReq.GetHead().GetRef(),
			HeadSha: pullReq.GetHead().GetSHA(),
			Title:   pullReq.GetTitle(),
		},
	}, nil
}

func (p *GitHubEventProcessor) getGitServerToken(ctx context.Context, codebase *codebaseApi.Codebase) (string, error) {
	gitServer := &codebaseApi.GitServer{}
	if err := p.ksClient.Get(ctx, types.NamespacedName{Namespace: codebase.Namespace, Name: codebase.Spec.GitServer}, gitServer); err != nil {
		return "", fmt.Errorf("failed to get GitServer: %w", err)
	}

	gitServerSecret := &corev1.Secret{}
	if err := p.ksClient.Get(ctx, types.NamespacedName{Namespace: codebase.Namespace, Name: gitServer.Spec.NameSshKeySecret}, gitServerSecret); err != nil {
		return "", fmt.Errorf("failed to get GitServer secret: %w", err)
	}

	token := string(gitServerSecret.Data[gitServerTokenField])

	if token == "" {
		return "", errors.New("token is empty in GitServer secret")
	}

	return token, nil
}
