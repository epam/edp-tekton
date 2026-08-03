package event_processor

import (
	"testing"

	"github.com/epam/edp-codebase-operator/v2/pkg/util/gitpathlabel"
	"github.com/stretchr/testify/assert"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"
)

// Golden-vector parity check: the interceptor must compute the identical
// wire-format value the operator stamps into the label.
func TestGitPathLabelWireFormatParity(t *testing.T) {
	assert.Equal(t, "app.edp.epam.com/gitUrlPathHash", codebaseApi.GitUrlPathHashLabel)
	assert.Equal(t, "89f3172369ae09a6b91fa78494b89639", gitpathlabel.Hash("/MyOrg/My-App"))
}
