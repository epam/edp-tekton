package event_processor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"
	"github.com/epam/edp-codebase-operator/v2/pkg/util/gitpathlabel"
)

const (
	CodebaseListLimit = 1000
)

// GetCodebaseByRepoPath returns codebase by repository path.
//
// The lookup selects by the gitUrlPathHash label the codebase-operator stamps
// on every reconcile, so the apiserver returns at most a handful of objects
// instead of the whole namespace. The hash is only a filter: candidates are
// verified against spec.gitUrlPath, so a stale or colliding label can never
// yield a wrong codebase — at worst the lookup degrades to the full scan.
func GetCodebaseByRepoPath(
	ctx context.Context,
	client ctrlClient.Reader,
	ns, repoPath string,
	log *zap.SugaredLogger,
) (*codebaseApi.Codebase, error) {
	if log == nil {
		log = zap.NewNop().Sugar()
	}

	labeled := &codebaseApi.CodebaseList{}
	if err := client.List(
		ctx,
		labeled,
		ctrlClient.InNamespace(ns),
		ctrlClient.MatchingLabels{codebaseApi.GitUrlPathHashLabel: gitpathlabel.Hash(repoPath)},
	); err != nil {
		return nil, fmt.Errorf("unable to list codebases by git url path hash: %w", err)
	}

	if codebase := matchCodebaseByRepoPath(labeled.Items, repoPath, log); codebase != nil {
		return codebase, nil
	}

	log.Infof("Codebase for repository path %s not found by label, falling back to full scan", repoPath)

	full := &codebaseApi.CodebaseList{}
	if err := client.List(ctx, full, ctrlClient.InNamespace(ns), ctrlClient.Limit(CodebaseListLimit)); err != nil {
		return nil, fmt.Errorf("unable to get codebase list: %w", err)
	}

	if codebase := matchCodebaseByRepoPath(full.Items, repoPath, log); codebase != nil {
		// The retirement signal for the full scan: it fires only when a real
		// codebase is missing its hash label (operator not upgraded, or label
		// removed), never for webhooks from unknown repositories. Once this
		// warning stays silent for a soak window, the fallback can be deleted.
		log.Warnf("Codebase %s resolved via full scan: its gitUrlPathHash label is missing or stale", codebase.Name)

		return codebase, nil
	}

	return nil, fmt.Errorf("codebase with repository path %s not found", repoPath)
}

func matchCodebaseByRepoPath(
	items []codebaseApi.Codebase,
	repoPath string,
	log *zap.SugaredLogger,
) *codebaseApi.Codebase {
	var matches []*codebaseApi.Codebase

	for n := range items {
		if strings.EqualFold(items[n].Spec.GitUrlPath, repoPath) {
			matches = append(matches, &items[n])
		}
	}

	if len(matches) == 0 {
		return nil
	}

	// Two codebases claiming one repository is a misconfiguration: they compete
	// for webhooks, tags and branch triggers. Surface it instead of silently
	// serving whichever sorts first.
	if len(matches) > 1 {
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name
		}

		log.Warnf("Multiple Codebases %v match repository path %s (case-insensitive), using %s",
			names, repoPath, matches[0].Name)
	}

	return matches[0]
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
	if err := client.Get(
		ctx,
		types.NamespacedName{Namespace: codebase.Namespace, Name: codebase.Spec.GitServer},
		gitServer,
	); err != nil {
		return "", fmt.Errorf("failed to get GitServer: %w", err)
	}

	gitServerSecret := &corev1.Secret{}
	if err := client.Get(
		ctx,
		types.NamespacedName{Namespace: codebase.Namespace, Name: gitServer.Spec.NameSshKeySecret},
		gitServerSecret,
	); err != nil {
		return "", fmt.Errorf("failed to get GitServer secret: %w", err)
	}

	token := string(gitServerSecret.Data[GitServerTokenField])

	if token == "" {
		return "", errors.New("token is empty in GitServer secret")
	}

	return token, nil
}
