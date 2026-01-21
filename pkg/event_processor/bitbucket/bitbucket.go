package bitbucket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/event_processor"
)

type EventProcessor struct {
	ksClient    ctrlClient.Reader
	logger      *zap.SugaredLogger
	restyClient *resty.Client
}

type EventProcessorOptions struct {
	Logger      *zap.SugaredLogger
	RestyClient *resty.Client
}

func NewEventProcessor(ksClient ctrlClient.Reader, options *EventProcessorOptions) *EventProcessor {
	if options == nil {
		options = &EventProcessorOptions{}
	}

	if options.Logger == nil {
		options.Logger = zap.NewNop().Sugar()
	}

	if options.RestyClient == nil {
		options.RestyClient = resty.New().SetBaseURL("https://api.bitbucket.org/2.0")
	}

	return &EventProcessor{
		ksClient:    ksClient,
		logger:      options.Logger,
		restyClient: options.RestyClient,
	}
}

func (p *EventProcessor) Process(
	ctx context.Context,
	body []byte,
	ns, eventType string,
) (*event_processor.EventInfo, error) {
	switch eventType {
	case event_processor.BitbucketEventTypeCommentAdded:
		return p.processCommentEvent(ctx, body, ns)
	default:
		return p.processMergeEvent(ctx, body, ns)
	}
}

func (p *EventProcessor) processMergeEvent(
	ctx context.Context,
	body []byte,
	ns string,
) (*event_processor.EventInfo, error) {
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

	commitMessage, err := p.getPRLatestCommitMessage(
		ctx,
		codebase,
		bitbucketEvent.Repository.FullName,
		bitbucketEvent.PullRequest.ID,
	)
	if err != nil {
		return nil, err
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
			LastCommitMessage: commitMessage,
			Author:            bitbucketEvent.PullRequest.Author.DisplayName,
			AuthorAvatarUrl:   bitbucketEvent.PullRequest.Author.Links.Avatar.Href,
			Url:               bitbucketEvent.PullRequest.Links.Html.Href,
		},
	}, nil
}

func (p *EventProcessor) processCommentEvent(
	ctx context.Context,
	body []byte,
	ns string,
) (*event_processor.EventInfo, error) {
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

	commitMessage, err := p.getPRLatestCommitMessage(
		ctx,
		codebase,
		bitbucketEvent.Repository.FullName,
		bitbucketEvent.PullRequest.ID,
	)
	if err != nil {
		return nil, err
	}

	return &event_processor.EventInfo{
		GitProvider:        event_processor.GitProviderBitbucket,
		RepoPath:           repoPath,
		TargetBranch:       bitbucketEvent.PullRequest.Destination.Branch.Name,
		Type:               event_processor.EventTypeReviewComment,
		Codebase:           codebase,
		HasPipelineRecheck: event_processor.ContainsPipelineRecheckPrefix(bitbucketEvent.Comment.Content.Raw),
		PullRequest: &event_processor.PullRequest{
			HeadSha:           bitbucketEvent.PullRequest.Source.Commit.Hash,
			Title:             bitbucketEvent.PullRequest.Title,
			HeadRef:           bitbucketEvent.PullRequest.Source.Branch.Name,
			ChangeNumber:      bitbucketEvent.PullRequest.ID,
			LastCommitMessage: commitMessage,
			Author:            bitbucketEvent.PullRequest.Author.DisplayName,
			AuthorAvatarUrl:   bitbucketEvent.PullRequest.Author.Links.Avatar.Href,
			Url:               bitbucketEvent.PullRequest.Links.Html.Href,
		},
	}, nil
}

type getPRCommitsResp struct {
	Values []struct {
		Message string `json:"message"`
	} `json:"values"`
}

func (p *EventProcessor) getPRLatestCommitMessage(
	ctx context.Context,
	codebase *codebaseApi.Codebase,
	repoFullName string,
	prID int,
) (string, error) {
	gitServerToken, err := event_processor.GetGitServerToken(ctx, p.ksClient, codebase)
	if err != nil {
		return "", fmt.Errorf("failed to get git server token for Bitbucket: %w", err)
	}

	commits := getPRCommitsResp{}

	r, err := p.restyClient.R().
		SetContext(ctx).
		ForceContentType("application/json").
		SetHeader("Authorization", fmt.Sprintf("Basic %s", gitServerToken)).
		SetResult(&commits).
		Get(fmt.Sprintf("/repositories/%s/pullrequests/%d/commits?fields=values.message&pagelen=1", repoFullName, prID))
	if err != nil {
		return "", fmt.Errorf("failed to get PR latest commit message: %w", err)
	}

	if r.IsError() {
		return "", fmt.Errorf("failed to get PR latest commit message: %s", r.String())
	}

	if len(commits.Values) == 0 {
		return "", errors.New("pull request doesn't have commits")
	}

	return commits.Values[0].Message, nil
}
