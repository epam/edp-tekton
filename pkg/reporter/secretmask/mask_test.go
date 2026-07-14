package secretmask

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	tektonpipelineApi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestMaskerLongestFirst(t *testing.T) {
	t.Parallel()

	// The short secret is a prefix of the long one: replacing the short one
	// first would leave the tail of the long secret exposed.
	masker := NewMasker([]string{"secret", "secret-with-long-tail", ""})

	got := masker.Mask("token=secret-with-long-tail other=secret done")

	assert.Equal(t, "token=***** other=***** done", got)
}

func TestMaskerNoValues(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "as is", NewMasker(nil).Mask("as is"))
}

func TestMaskerEncodedForms(t *testing.T) {
	t.Parallel()

	secret := "s3cr3t-token-value"
	masker := NewMasker([]string{secret})

	tests := []struct {
		name    string
		encoded string
	}{
		{name: "raw", encoded: secret},
		{name: "base64", encoded: base64.StdEncoding.EncodeToString([]byte(secret))},
		{name: "base64u", encoded: base64.URLEncoding.EncodeToString([]byte(secret))},
		{name: "hex", encoded: hex.EncodeToString([]byte(secret))},
		{name: "HEX", encoded: strings.ToUpper(hex.EncodeToString([]byte(secret)))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			masked := masker.Mask("log line: " + tt.encoded + " end")
			assert.NotContains(t, masked, tt.encoded, "encoded secret form must be masked")
			assert.Contains(t, masked, replacement)
		})
	}
}

func TestCollectSecretValues(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	assert.NoError(t, corev1.AddToScheme(scheme))

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "sonar-token", Namespace: "krci"},
		Data:       map[string][]byte{"token": []byte("sonar-secret-value")},
	}

	envFromSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "registry-creds", Namespace: "krci"},
		Data: map[string][]byte{
			"CI_USERNAME": []byte("registry-user"),
			"CI_PASSWORD": []byte("registry-password"),
		},
	}

	taskRun := &tektonpipelineApi.TaskRun{
		ObjectMeta: metav1.ObjectMeta{Name: "tr", Namespace: "krci"},
		Status: tektonpipelineApi.TaskRunStatus{
			TaskRunStatusFields: tektonpipelineApi.TaskRunStatusFields{
				TaskSpec: &tektonpipelineApi.TaskSpec{
					Steps: []tektonpipelineApi.Step{
						{
							Name: "scan",
							Env: []corev1.EnvVar{
								{
									Name: "SONAR_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: "sonar-token"},
											Key:                  "token",
										},
									},
								},
								{Name: "PLAIN", Value: "not-a-secret"},
							},
						},
						{
							Name: "push",
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{Name: "registry-creds"},
									},
								},
								{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{Name: "not-a-secret-cm"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	missingRefTaskRun := &tektonpipelineApi.TaskRun{
		ObjectMeta: metav1.ObjectMeta{Name: "tr2", Namespace: "krci"},
		Status: tektonpipelineApi.TaskRunStatus{
			TaskRunStatusFields: tektonpipelineApi.TaskRunStatusFields{
				TaskSpec: &tektonpipelineApi.TaskSpec{
					Steps: []tektonpipelineApi.Step{
						{
							Name: "s",
							Env: []corev1.EnvVar{
								{
									Name: "GONE",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: "deleted-secret"},
											Key:                  "token",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	reader := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret, envFromSecret).Build()

	values := CollectSecretValues(
		context.Background(),
		reader,
		[]*tektonpipelineApi.TaskRun{taskRun, missingRefTaskRun},
		"git-token",
	)

	assert.ElementsMatch(t,
		[]string{"git-token", "sonar-secret-value", "registry-user", "registry-password"},
		values,
		"envFrom secretRef values must be collected alongside env secretKeyRef values")
}
