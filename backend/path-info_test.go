package backend

import (
	"fmt"
	"github.com/Shuttl-Tech/vault-plugin-sentry/version"
	"github.com/google/go-cmp/cmp"
	logicaltest "github.com/hashicorp/vault/helper/testhelpers/logical"
	"github.com/hashicorp/vault/sdk/logical"
	"testing"
)

func TestHandleInfo(t *testing.T) {
	logicaltest.Test(t, logicaltest.TestCase{
		LogicalBackend: testGetBackend(t),
		Steps: []logicaltest.TestStep{
			testReadInfo(),
		},
	})
}

func testReadInfo() logicaltest.TestStep {
	version.GitCommit = "test-commit"
	version.Name = "v0.0.0-test"

	return logicaltest.TestStep{
		Operation: logical.ReadOperation,
		Path:      "info",
		Check: func(resp *logical.Response) error {
			expect := map[string]interface{}{
				"description": "Manage Sentry projects and their DSN",
				"commit_sha":  version.GitCommit,
				"version":     " v0.1.0 ()",
			}

			if !cmp.Equal(expect, resp.Data) {
				return fmt.Errorf("unexpected data in response. %s", cmp.Diff(expect, resp.Data))
			}

			return nil
		},
	}
}
