package backend

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	logicaltest "github.com/hashicorp/vault/helper/testhelpers/logical"
	"github.com/hashicorp/vault/sdk/logical"
	"net/http"
	"strings"
	"testing"
)

func TestHandleConfigUpdateAndRead(t *testing.T) {
	org, token, endpoint, timeout := "test-org-config", "token-123", localSentry.url, 10

	logicaltest.Test(t, logicaltest.TestCase{
		LogicalBackend: testGetBackend(t),
		Steps: []logicaltest.TestStep{
			testReadConfigErr("plugin is not configured"),
			testWriteConfig(org, token, endpoint, timeout),
			testReadConfig(org, "display-name-"+org, endpoint, timeout),
		},
	})
}

func testReadConfigErr(msg string) logicaltest.TestStep {
	return logicaltest.TestStep{
		Operation: logical.ReadOperation,
		Path:      "config",
		ErrorOk:   true,
		Check: func(resp *logical.Response) error {
			if !resp.IsError() {
				return fmt.Errorf("expected error, got valid response")
			}

			if !strings.Contains(resp.Error().Error(), msg) {
				return fmt.Errorf("response error %q does not contain expected %q", resp.Error(), msg)
			}

			return nil
		},
	}
}

func testReadConfig(org, displayName, endpoint string, timeout int) logicaltest.TestStep {
	return logicaltest.TestStep{
		Operation: logical.ReadOperation,
		Path:      "config",
		ErrorOk:   false,
		Check: func(resp *logical.Response) error {
			expect := map[string]interface{}{
				"name":         org,
				"display_name": displayName,
				"endpoint":     endpoint,
				"timeout":      timeout,
			}

			if !cmp.Equal(expect, resp.Data) {
				return fmt.Errorf("unexpected data in response. %s", cmp.Diff(expect, resp.Data))
			}

			return nil
		},
	}

}

func testWriteConfig(org, token, endpoint string, timeout int) logicaltest.TestStep {
	localSentry.handleStatic("/organizations/"+org+"/", http.StatusOK, fmt.Sprintf(getOrgResponseBody, "display-name-"+org, org))

	return logicaltest.TestStep{
		Operation: logical.UpdateOperation,
		Path:      "config",
		ErrorOk:   false,
		Data: map[string]interface{}{
			"org":      org,
			"token":    token,
			"endpoint": endpoint,
			"timeout":  timeout,
		},
		Check: func(resp *logical.Response) error {
			expect := map[string]interface{}{
				"name":         org,
				"display_name": "display-name-" + org,
				"endpoint":     endpoint,
				"timeout":      timeout,
			}

			if !cmp.Equal(expect, resp.Data) {
				return fmt.Errorf("unexpected data in response. %s", cmp.Diff(expect, resp.Data))
			}

			return nil
		},
	}
}

const getOrgResponseBody = `
{
  "id": "2", 
  "name": "%s", 
  "slug": "%s", 
  "dateCreated": "2018-11-06T21:19:55.101Z", 
  "features": ["new-teams", "shared-issues", "new-issue-ui", "repos"], 
  "isEarlyAdopter": false, 
  "pendingAccessRequests": 0, 
  "quota": {
    "accountLimit": 0, 
    "maxRate": 0, 
    "maxRateInterval": 60, 
    "projectLimit": 100
  }, 
  "teams": [
    {
      "avatar": {
        "avatarType": "letter_avatar", 
        "avatarUuid": null
      }, 
      "dateCreated": "2018-11-06T21:20:08.115Z", 
      "hasAccess": true, 
      "id": "3", 
      "isMember": true, 
      "isPending": false, 
      "memberCount": 1, 
      "name": "Ancient Gabelers", 
      "slug": "ancient-gabelers"
    }
  ]
}
`
