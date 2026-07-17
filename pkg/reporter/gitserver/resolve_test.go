package gitserver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"
)

func newReader(t *testing.T, objects ...ctrlClient.Object) ctrlClient.Reader {
	t.Helper()

	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, codebaseApi.AddToScheme(scheme))

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

func testObjects() []ctrlClient.Object {
	return []ctrlClient.Object{
		&codebaseApi.Codebase{
			ObjectMeta: metav1.ObjectMeta{Name: "my-app", Namespace: "krci"},
			Spec:       codebaseApi.CodebaseSpec{GitServer: "github"},
		},
		&codebaseApi.GitServer{
			ObjectMeta: metav1.ObjectMeta{Name: "github", Namespace: "krci"},
			Spec: codebaseApi.GitServerSpec{
				GitHost:          "github.com",
				GitProvider:      codebaseApi.GitProviderGithub,
				NameSshKeySecret: "ci-github",
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "ci-github", Namespace: "krci"},
			Data:       map[string][]byte{"token": []byte("gh-token")},
		},
	}
}

func TestResolve(t *testing.T) {
	t.Parallel()

	info, err := Resolve(context.Background(), newReader(t, testObjects()...), "krci", "my-app")
	require.NoError(t, err)

	assert.Equal(t, &Info{Provider: codebaseApi.GitProviderGithub, Host: "github.com", Token: "gh-token"}, info)
}

func TestResolveGerritWithoutTokenSucceeds(t *testing.T) {
	t.Parallel()

	// Gerrit's GitServer secret is an SSH key with no "token" field; Resolve
	// must pass it through so provider.New can reject it with a permanent
	// error instead of the run requeuing here forever.
	objects := []ctrlClient.Object{
		&codebaseApi.Codebase{
			ObjectMeta: metav1.ObjectMeta{Name: "gr-app", Namespace: "krci"},
			Spec:       codebaseApi.CodebaseSpec{GitServer: "gerrit"},
		},
		&codebaseApi.GitServer{
			ObjectMeta: metav1.ObjectMeta{Name: "gerrit", Namespace: "krci"},
			Spec: codebaseApi.GitServerSpec{
				GitHost:          "gerrit.example.com",
				GitProvider:      codebaseApi.GitProviderGerrit,
				NameSshKeySecret: "ci-gerrit",
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "ci-gerrit", Namespace: "krci"},
			Data:       map[string][]byte{"id_rsa": []byte("key")},
		},
	}

	info, err := Resolve(context.Background(), newReader(t, objects...), "krci", "gr-app")
	require.NoError(t, err)

	assert.Equal(t, &Info{Provider: codebaseApi.GitProviderGerrit, Host: "gerrit.example.com", Token: ""}, info)
}

func TestResolveErrors(t *testing.T) {
	t.Parallel()

	t.Run("codebase not found", func(t *testing.T) {
		t.Parallel()

		_, err := Resolve(context.Background(), newReader(t), "krci", "missing")
		assert.ErrorContains(t, err, "failed to get Codebase")
	})

	t.Run("gitserver not found", func(t *testing.T) {
		t.Parallel()

		objects := testObjects()[:1]

		_, err := Resolve(context.Background(), newReader(t, objects...), "krci", "my-app")
		assert.ErrorContains(t, err, "failed to get GitServer")
	})

	t.Run("secret not found", func(t *testing.T) {
		t.Parallel()

		objects := testObjects()[:2]

		_, err := Resolve(context.Background(), newReader(t, objects...), "krci", "my-app")
		assert.ErrorContains(t, err, "failed to get GitServer secret")
	})

	t.Run("empty token", func(t *testing.T) {
		t.Parallel()

		objects := testObjects()
		objects[2] = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "ci-github", Namespace: "krci"},
			Data:       map[string][]byte{"id_rsa": []byte("key")},
		}

		_, err := Resolve(context.Background(), newReader(t, objects...), "krci", "my-app")
		assert.ErrorContains(t, err, "token is empty")
	})
}
