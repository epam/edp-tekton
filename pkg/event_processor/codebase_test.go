package event_processor

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"
	"github.com/epam/edp-codebase-operator/v2/pkg/util/gitpathlabel"
)

func TestConvertRepositoryPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "With leading slash and mixed case",
			input: "/MyRepo",
			want:  "/myrepo",
		},
		{
			name:  "Without leading slash",
			input: "MyRepo",
			want:  "/myrepo",
		},
		{
			name:  "Already lowercase with leading slash",
			input: "/myrepo",
			want:  "/myrepo",
		},
		{
			name:  "Mixed case with subpath",
			input: "MyRepo/SubPath",
			want:  "/myrepo/subpath",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "/",
		},
		{
			name:  "Only slash",
			input: "/",
			want:  "/",
		},
		{
			name:  "Multiple slashes",
			input: "org/repo/path",
			want:  "/org/repo/path",
		},
		{
			name:  "Uppercase",
			input: "ORG/REPO",
			want:  "/org/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertRepositoryPath(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetCodebaseByRepoPath(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))

	tests := []struct {
		name        string
		namespace   string
		repoPath    string
		kubeObjects []client.Object
		want        *codebaseApi.Codebase
		wantErr     require.ErrorAssertionFunc
	}{
		{
			name:      "Success - find codebase with exact match",
			namespace: "default",
			repoPath:  "/org/repo",
			kubeObjects: []client.Object{
				createTestCodebase("codebase1", "default", "github", "/org/repo"),
			},
			want:    createTestCodebase("codebase1", "default", "github", "/org/repo"),
			wantErr: require.NoError,
		},
		{
			name:      "Success - case insensitive match",
			namespace: "default",
			repoPath:  "/org/repo",
			kubeObjects: []client.Object{
				createTestCodebase("codebase1", "default", "github", "/ORG/REPO"),
			},
			want:    createTestCodebase("codebase1", "default", "github", "/ORG/REPO"),
			wantErr: require.NoError,
		},
		{
			name:      "Success - mixed case match",
			namespace: "default",
			repoPath:  "/Org/Repo",
			kubeObjects: []client.Object{
				createTestCodebase("codebase1", "default", "github", "/org/repo"),
			},
			want:    createTestCodebase("codebase1", "default", "github", "/org/repo"),
			wantErr: require.NoError,
		},
		{
			name:      "Success - multiple codebases, find correct one",
			namespace: "default",
			repoPath:  "/org/repo2",
			kubeObjects: []client.Object{
				createTestCodebase("codebase1", "default", "github", "/org/repo1"),
				createTestCodebase("codebase2", "default", "github", "/org/repo2"),
				createTestCodebase("codebase3", "default", "github", "/org/repo3"),
			},
			want:    createTestCodebase("codebase2", "default", "github", "/org/repo2"),
			wantErr: require.NoError,
		},
		{
			name:        "Error - codebase not found in empty list",
			namespace:   "default",
			repoPath:    "/org/repo",
			kubeObjects: []client.Object{},
			want:        nil,
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "codebase with repository path /org/repo not found")
			},
		},
		{
			name:      "Error - codebase not found",
			namespace: "default",
			repoPath:  "/org/nonexistent",
			kubeObjects: []client.Object{
				createTestCodebase("codebase1", "default", "github", "/org/repo1"),
				createTestCodebase("codebase2", "default", "github", "/org/repo2"),
			},
			want: nil,
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "codebase with repository path /org/nonexistent not found")
			},
		},
		{
			name:      "Success - first match in multiple codebases",
			namespace: "test-ns",
			repoPath:  "/shared/repo",
			kubeObjects: []client.Object{
				createTestCodebase("codebase1", "test-ns", "github", "/shared/repo"),
				createTestCodebase("codebase2", "test-ns", "gitlab", "/shared/repo"),
			},
			want:    createTestCodebase("codebase1", "test-ns", "github", "/shared/repo"),
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.kubeObjects...).
				Build()

			got, err := GetCodebaseByRepoPath(context.Background(), k8sClient, tt.namespace, tt.repoPath, nil)
			tt.wantErr(t, err)

			if tt.want != nil {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.Name, got.Name)
				assert.Equal(t, tt.want.Namespace, got.Namespace)
				assert.Equal(t, tt.want.Spec, got.Spec)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestGetGitServerToken(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	tests := []struct {
		name        string
		codebase    *codebaseApi.Codebase
		kubeObjects []client.Object
		want        string
		wantErr     require.ErrorAssertionFunc
	}{
		{
			name:     "Success - retrieve token from secret",
			codebase: createTestCodebase("codebase1", "default", "github", "/org/repo"),
			kubeObjects: []client.Object{
				createTestGitServer("github", "default", "git-secret"),
				createTestSecret("git-secret", "default", "my-secret-token"),
			},
			want:    "my-secret-token",
			wantErr: require.NoError,
		},
		{
			name:     "Error - GitServer not found",
			codebase: createTestCodebase("codebase1", "default", "nonexistent-server", "/org/repo"),
			kubeObjects: []client.Object{
				createTestGitServer("github", "default", "git-secret"),
				createTestSecret("git-secret", "default", "my-secret-token"),
			},
			want: "",
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get GitServer")
			},
		},
		{
			name:     "Error - Secret not found",
			codebase: createTestCodebase("codebase1", "default", "github", "/org/repo"),
			kubeObjects: []client.Object{
				createTestGitServer("github", "default", "nonexistent-secret"),
			},
			want: "",
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get GitServer secret")
			},
		},
		{
			name:     "Error - empty token in secret",
			codebase: createTestCodebase("codebase1", "default", "github", "/org/repo"),
			kubeObjects: []client.Object{
				createTestGitServer("github", "default", "git-secret"),
				createTestSecret("git-secret", "default", ""),
			},
			want: "",
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "token is empty in GitServer secret")
			},
		},
		{
			name:     "Error - missing token field in secret",
			codebase: createTestCodebase("codebase1", "default", "github", "/org/repo"),
			kubeObjects: []client.Object{
				createTestGitServer("github", "default", "git-secret"),
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "git-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"other-field": []byte("some-value"),
					},
				},
			},
			want: "",
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "token is empty in GitServer secret")
			},
		},
		{
			name:     "Success - token in different namespace",
			codebase: createTestCodebase("codebase1", "test-namespace", "gitlab", "/org/repo"),
			kubeObjects: []client.Object{
				createTestGitServer("gitlab", "test-namespace", "gitlab-secret"),
				createTestSecret("gitlab-secret", "test-namespace", "gitlab-token-123"),
			},
			want:    "gitlab-token-123",
			wantErr: require.NoError,
		},
		{
			name:     "Success - token with special characters",
			codebase: createTestCodebase("codebase1", "default", "github", "/org/repo"),
			kubeObjects: []client.Object{
				createTestGitServer("github", "default", "git-secret"),
				createTestSecret("git-secret", "default", "ghp_1234567890ABCDEF!@#$%^&*()"),
			},
			want:    "ghp_1234567890ABCDEF!@#$%^&*()",
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.kubeObjects...).
				Build()

			got, err := GetGitServerToken(context.Background(), k8sClient, tt.codebase)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// createTestCodebase creates a standard codebase object for testing.
func createTestCodebase(name, namespace, gitServer, gitUrlPath string) *codebaseApi.Codebase {
	return &codebaseApi.Codebase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: codebaseApi.CodebaseSpec{
			GitServer:  gitServer,
			GitUrlPath: gitUrlPath,
		},
	}
}

// createTestGitServer creates a GitServer object for testing.
func createTestGitServer(name, namespace, secretName string) *codebaseApi.GitServer {
	return &codebaseApi.GitServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: codebaseApi.GitServerSpec{
			NameSshKeySecret: secretName,
		},
	}
}

// createTestSecret creates a Secret with token data.
func createTestSecret(name, namespace, token string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			GitServerTokenField: []byte(token),
		},
	}
}

func TestGetCodebaseByRepoPath_NilLoggerIsSafe(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(
			createTestCodebase("codebase-a", "default", "github", "/shared/repo"),
			createTestCodebase("codebase-b", "default", "github", "/Shared/Repo"),
		).
		Build()

	got, err := GetCodebaseByRepoPath(context.Background(), k8sClient, "default", "/shared/repo", nil)
	require.NoError(t, err)
	require.Equal(t, "codebase-a", got.Name)
}

// createLabeledTestCodebase mimics a Codebase reconciled by an operator
// that stamps the gitUrlPathHash label.
func createLabeledTestCodebase(name, gitUrlPath string) *codebaseApi.Codebase {
	cb := createTestCodebase(name, "default", "github", gitUrlPath)
	cb.Labels = map[string]string{
		codebaseApi.GitUrlPathHashLabel: gitpathlabel.Hash(gitUrlPath),
	}

	return cb
}

func TestGetCodebaseByRepoPath_LabelSelectedLookup(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, codebaseApi.AddToScheme(scheme))

	const ns = "default"

	tests := []struct {
		name         string
		kubeObjects  []client.Object
		repoPath     string
		wantName     string
		wantErr      bool
		wantFallback bool
		wantResolved bool
		wantWarnSubs []string
	}{
		{
			name: "labeled codebase resolved without fallback",
			kubeObjects: []client.Object{
				createLabeledTestCodebase("codebase-a", "/Org/Repo"),
				createLabeledTestCodebase("codebase-other", "/org/other"),
			},
			repoPath: "/org/repo",
			wantName: "codebase-a",
		},
		{
			name: "unlabeled codebase resolved via fallback (pre-upgrade operator)",
			kubeObjects: []client.Object{
				createTestCodebase("codebase-a", ns, "github", "/org/repo"),
			},
			repoPath:     "/org/repo",
			wantName:     "codebase-a",
			wantFallback: true,
			wantResolved: true,
		},
		{
			name: "impostor label rejected by spec verification, real codebase found via fallback",
			kubeObjects: []client.Object{
				// Correct hash label for /org/repo but the spec points elsewhere.
				func() *codebaseApi.Codebase {
					cb := createTestCodebase("impostor", ns, "github", "/evil/other")
					cb.Labels = map[string]string{
						codebaseApi.GitUrlPathHashLabel: gitpathlabel.Hash("/org/repo"),
					}

					return cb
				}(),
				createTestCodebase("codebase-real", ns, "github", "/org/repo"),
			},
			repoPath:     "/org/repo",
			wantName:     "codebase-real",
			wantFallback: true,
			wantResolved: true,
		},
		{
			name: "case-insensitive duplicates warn on the label path without fallback",
			kubeObjects: []client.Object{
				createLabeledTestCodebase("codebase-a", "/Shared/Repo"),
				createLabeledTestCodebase("codebase-b", "/shared/repo"),
			},
			repoPath: "/shared/repo",
			wantName: "codebase-a",
			wantWarnSubs: []string{
				"codebase-a", "codebase-b", "/shared/repo",
			},
		},
		{
			name: "not found goes through both paths",
			kubeObjects: []client.Object{
				createLabeledTestCodebase("codebase-a", "/org/repo"),
			},
			repoPath:     "/org/nonexistent",
			wantErr:      true,
			wantFallback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zapcore.InfoLevel)
			logger := zap.New(core).Sugar()

			k8sClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.kubeObjects...).
				Build()

			got, err := GetCodebaseByRepoPath(context.Background(), k8sClient, ns, tt.repoPath, logger)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "not found")
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantName, got.Name)
			}

			var gotFallback, gotResolved bool

			var dupWarns []string

			for _, entry := range logs.All() {
				if entry.Level == zapcore.InfoLevel && strings.Contains(entry.Message, "falling back to full scan") {
					gotFallback = true
				}

				if entry.Level == zapcore.WarnLevel && strings.Contains(entry.Message, "resolved via full scan") {
					gotResolved = true
				}

				if entry.Level == zapcore.WarnLevel && strings.Contains(entry.Message, "Multiple Codebases") {
					dupWarns = append(dupWarns, entry.Message)
				}
			}

			assert.Equal(t, tt.wantFallback, gotFallback, "fallback log mismatch: %v", logs.All())
			assert.Equal(t, tt.wantResolved, gotResolved, "fallback-resolved warning mismatch: %v", logs.All())

			if len(tt.wantWarnSubs) == 0 {
				assert.Empty(t, dupWarns, "unexpected duplicate warning")

				return
			}

			require.Len(t, dupWarns, 1, "expected exactly one duplicate warning")

			for _, sub := range tt.wantWarnSubs {
				assert.Contains(t, dupWarns[0], sub)
			}
		})
	}
}
