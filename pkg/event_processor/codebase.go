package event_processor

import (
	"context"
	"fmt"
	"strings"

	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/pkg/apis/edp/v1"
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
		if codebaseList.Items[n].Spec.GitUrlPath != nil && strings.EqualFold(*codebaseList.Items[n].Spec.GitUrlPath, repoPath) {
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
