package secretmask

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
