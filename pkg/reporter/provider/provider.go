// Package provider publishes pipeline report comments to git providers.
package provider

import (
	"fmt"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"

	"github.com/epam/edp-tekton/pkg/reporter/provider/bitbucket"
	"github.com/epam/edp-tekton/pkg/reporter/provider/github"
	"github.com/epam/edp-tekton/pkg/reporter/provider/gitlab"
	"github.com/epam/edp-tekton/pkg/reporter/provider/types"
)

// New returns the types.Provider implementation for the given git provider
// type (codebaseApi.GitProvider* constants). Implementations key on
// Comment.Marker: the update strategy finds a prior comment carrying the same
// marker and edits it instead of creating a duplicate, the recreate strategy
// creates a new comment and deletes the marked stale ones. Gerrit is not
// supported yet.
func New(gitProvider, host, token string) (types.Provider, error) {
	switch gitProvider {
	case codebaseApi.GitProviderGithub:
		return github.New(host, token)
	case codebaseApi.GitProviderGitlab:
		return gitlab.New(host, token), nil
	case codebaseApi.GitProviderBitbucket:
		return bitbucket.New(token), nil
	default:
		return nil, fmt.Errorf("unsupported git provider %q", gitProvider)
	}
}

// NewCommitStatusSetter returns the types.CommitStatusSetter implementation
// for the given git provider type. Gerrit is not supported: its review
// reporting is vote-based.
func NewCommitStatusSetter(gitProvider, host, token string) (types.CommitStatusSetter, error) {
	switch gitProvider {
	case codebaseApi.GitProviderGithub:
		return github.New(host, token)
	case codebaseApi.GitProviderGitlab:
		return gitlab.New(host, token), nil
	case codebaseApi.GitProviderBitbucket:
		return bitbucket.New(token), nil
	default:
		return nil, fmt.Errorf("unsupported git provider %q", gitProvider)
	}
}
