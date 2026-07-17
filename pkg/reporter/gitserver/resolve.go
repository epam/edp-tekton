// Package gitserver resolves the git provider connection details for a
// PipelineRun: Codebase -> GitServer -> access token.
package gitserver

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"
)

const tokenSecretField = "token"

// tokenRequiredProviders are the git providers whose reporter integration
// authenticates with an API token, so their GitServer secret must carry a
// non-empty one. Providers outside this set (currently Gerrit, which the
// reporter reaches over SSH and does not support as a comment target) resolve
// without a token; report() rejects them later via provider.New, whose error
// is classified as permanent so the PipelineRun is marked skipped once
// instead of endlessly requeuing on a token that will never appear.
var tokenRequiredProviders = map[string]bool{
	codebaseApi.GitProviderGithub:    true,
	codebaseApi.GitProviderGitlab:    true,
	codebaseApi.GitProviderBitbucket: true,
}

// Info is the resolved git provider connection info for a Codebase.
type Info struct {
	// Provider is one of the codebaseApi.GitProvider* constants (github, gitlab, bitbucket, gerrit).
	Provider string
	// Host is the git server host, e.g. github.com.
	Host string
	// Token is the provider API access token.
	Token string
}

// Resolve looks up the Codebase by name, follows its GitServer reference and
// returns the provider type, host and API token.
func Resolve(ctx context.Context, reader ctrlClient.Reader, namespace, codebaseName string) (*Info, error) {
	codebase := &codebaseApi.Codebase{}
	if err := reader.Get(ctx, types.NamespacedName{Namespace: namespace, Name: codebaseName}, codebase); err != nil {
		return nil, fmt.Errorf("failed to get Codebase %s: %w", codebaseName, err)
	}

	return ResolveGitServer(ctx, reader, namespace, codebase.Spec.GitServer)
}

// ResolveGitServer returns the provider connection info for the named GitServer.
// Callers that already hold the Codebase can use it directly to skip the extra
// Codebase lookup Resolve performs.
func ResolveGitServer(ctx context.Context, reader ctrlClient.Reader, namespace, gitServerName string) (*Info, error) {
	gitServer := &codebaseApi.GitServer{}
	if err := reader.Get(
		ctx,
		types.NamespacedName{Namespace: namespace, Name: gitServerName},
		gitServer,
	); err != nil {
		return nil, fmt.Errorf("failed to get GitServer %s: %w", gitServerName, err)
	}

	secret := &corev1.Secret{}
	if err := reader.Get(
		ctx,
		types.NamespacedName{Namespace: namespace, Name: gitServer.Spec.NameSshKeySecret},
		secret,
	); err != nil {
		return nil, fmt.Errorf("failed to get GitServer secret %s: %w", gitServer.Spec.NameSshKeySecret, err)
	}

	token := string(secret.Data[tokenSecretField])
	if token == "" && tokenRequiredProviders[gitServer.Spec.GitProvider] {
		return nil, errors.New("token is empty in GitServer secret")
	}

	return &Info{
		Provider: gitServer.Spec.GitProvider,
		Host:     gitServer.Spec.GitHost,
		Token:    token,
	}, nil
}
