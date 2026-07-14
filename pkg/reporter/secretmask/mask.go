// Package secretmask hides Kubernetes Secret values referenced by a
// PipelineRun's steps from published log snippets.
package secretmask

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"sort"
	"strings"

	tektonpipelineApi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	replacement = "*****"
	// minEncodedLen guards against masking very short encoded forms that could
	// collide with ordinary log text.
	minEncodedLen = 6
)

// Masker replaces known secret values with a placeholder.
//
// Masking is best-effort defence-in-depth, not a security boundary: review
// pipelines run the pull request's own code in steps that may have credentials
// injected, so a malicious author can print an arbitrarily transformed secret
// that literal masking cannot recognise. Publishing failed-step logs to a pull
// request should therefore only be enabled for pipelines whose secret-bearing
// steps do not execute untrusted code. We mask the raw value plus its common
// reversible encodings (base64, hex) to catch accidental leaks and the most
// common exfiltration shapes.
type Masker struct {
	values []string
}

// NewMasker builds a Masker from an explicit list of secret values, expanding
// each into its common reversible encodings. Values are matched longest-first
// so a short secret sharing a prefix with a longer one cannot partially unmask
// the longer one.
func NewMasker(values []string) *Masker {
	seen := make(map[string]struct{})
	expanded := make([]string, 0, len(values)*maskFormsPerValue)

	for _, v := range values {
		for _, form := range encodedForms(v) {
			if len(form) < minEncodedLen {
				continue
			}

			if _, ok := seen[form]; ok {
				continue
			}

			seen[form] = struct{}{}
			expanded = append(expanded, form)
		}
	}

	sort.Slice(expanded, func(i, j int) bool {
		return len(expanded[i]) > len(expanded[j])
	})

	return &Masker{values: expanded}
}

// maskFormsPerValue is the number of representations encodedForms produces, used
// only to size the initial slice.
const maskFormsPerValue = 6

// encodedForms returns the raw secret plus the reversible encodings an attacker
// or a build tool is most likely to emit instead of the literal value.
func encodedForms(v string) []string {
	if v == "" {
		return nil
	}

	raw := []byte(v)

	return []string{
		v,
		base64.StdEncoding.EncodeToString(raw),
		base64.RawStdEncoding.EncodeToString(raw),
		base64.URLEncoding.EncodeToString(raw),
		hex.EncodeToString(raw),
		strings.ToUpper(hex.EncodeToString(raw)),
	}
}

// Mask replaces every occurrence of a known secret value with a placeholder.
func (m *Masker) Mask(text string) string {
	for _, v := range m.values {
		text = strings.ReplaceAll(text, v, replacement)
	}

	return text
}

// CollectSecretValues gathers the values of all Secrets referenced via
// SecretKeyRef env vars by the given TaskRuns' steps, plus the extra secret
// values passed in (e.g. the git provider token used by the reporter itself).
func CollectSecretValues(
	ctx context.Context,
	reader ctrlClient.Reader,
	taskRuns []*tektonpipelineApi.TaskRun,
	extra ...string,
) []string {
	secretNames := map[types.NamespacedName]struct{}{}

	for _, taskRun := range taskRuns {
		if taskRun.Status.TaskSpec == nil {
			continue
		}

		for _, step := range taskRun.Status.TaskSpec.Steps {
			for _, env := range step.Env {
				if env.ValueFrom == nil || env.ValueFrom.SecretKeyRef == nil {
					continue
				}

				secretNames[types.NamespacedName{
					Namespace: taskRun.Namespace,
					Name:      env.ValueFrom.SecretKeyRef.Name,
				}] = struct{}{}
			}

			for _, envFrom := range step.EnvFrom {
				if envFrom.SecretRef == nil {
					continue
				}

				secretNames[types.NamespacedName{
					Namespace: taskRun.Namespace,
					Name:      envFrom.SecretRef.Name,
				}] = struct{}{}
			}
		}
	}

	values := make([]string, 0, len(secretNames)+len(extra))
	values = append(values, extra...)

	for name := range secretNames {
		secret := &corev1.Secret{}
		// A secret that cannot be read cannot leak into the logs either; skip it.
		if err := reader.Get(ctx, name, secret); err != nil {
			continue
		}

		for _, data := range secret.Data {
			values = append(values, string(data))
		}
	}

	return values
}
