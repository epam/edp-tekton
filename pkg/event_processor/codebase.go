package event_processor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"
)

const (
	CodebaseListLimit = 1000
)

// GetCodebaseByRepoPath returns codebase by repository path.
func GetCodebaseByRepoPath(ctx context.Context, client ctrlClient.Reader, ns, repoPath string) (*codebaseApi.Codebase, error) {
	codebaseList := &codebaseApi.CodebaseList{}
	if err := client.List(ctx, codebaseList, ctrlClient.InNamespace(ns), ctrlClient.Limit(CodebaseListLimit)); err != nil {
		return nil, fmt.Errorf("unable to get codebase list: %w", err)
	}

	for n := range codebaseList.Items {
		if strings.EqualFold(codebaseList.Items[n].Spec.GitUrlPath, repoPath) {
			return &codebaseList.Items[n], nil
		}
	}

	return nil, fmt.Errorf("codebase with repository path %s not found", repoPath)
}

// ConvertRepositoryPath converts repository path to the format which is used in codebase.
func ConvertRepositoryPath(repo string) string {
	if !strings.HasPrefix(repo, "/") {
		repo = "/" + repo
	}

	return strings.ToLower(repo)
}

const GitServerTokenField = "token"

func GetGitServerToken(ctx context.Context, client ctrlClient.Reader, codebase *codebaseApi.Codebase) (string, error) {
	gitServer := &codebaseApi.GitServer{}
	if err := client.Get(ctx, types.NamespacedName{Namespace: codebase.Namespace, Name: codebase.Spec.GitServer}, gitServer); err != nil {
		return "", fmt.Errorf("failed to get GitServer: %w", err)
	}

	gitServerSecret := &corev1.Secret{}
	if err := client.Get(ctx, types.NamespacedName{Namespace: codebase.Namespace, Name: gitServer.Spec.NameSshKeySecret}, gitServerSecret); err != nil {
		return "", fmt.Errorf("failed to get GitServer secret: %w", err)
	}

	token := string(gitServerSecret.Data[GitServerTokenField])

	if token == "" {
		return "", errors.New("token is empty in GitServer secret")
	}

	return token, nil
}
