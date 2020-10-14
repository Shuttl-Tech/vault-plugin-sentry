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

func TestHandleDsnRead(t *testing.T) {
	org, token, endpoint, timeout := "dsn-org", "abcd-token", localSentry.url, 10
	project, team, dsnname := "test-app", "testers-team", "testing"

	logicaltest.Test(t, logicaltest.TestCase{
		LogicalBackend: testGetBackend(t),
		Steps: []logicaltest.TestStep{
			testReadDsnErr(project, dsnname, ""),
			testWriteConfig(org, token, endpoint, timeout),
			testReadDsnErr(project, dsnname, ""),
			testWriteProjectExisting(org, project, team, ""),
			testReadDsn(org, project, dsnname),

			testWriteProjectExisting(org, "default-dsn-"+project, team, "default-dsn-name"),
			testReadDefaultDsn(org, "default-dsn-"+project, "default-dsn-name"),
		},
	})
}

func testReadDsnErr(project, dsnname, msg string) logicaltest.TestStep {
	return logicaltest.TestStep{
		Operation: logical.ReadOperation,
		Path:      "dsn/" + project + "/" + dsnname,
		ErrorOk:   true,
		Check: func(resp *logical.Response) error {
			if !resp.IsError() {
				return fmt.Errorf("expected error in response, got none")
			}

			if !strings.Contains(resp.Error().Error(), msg) {
				return fmt.Errorf("unexpected error %q does not match %q", resp.Error(), msg)
			}

			return nil
		},
	}
}

func testReadDefaultDsn(org, project, dsnname string) logicaltest.TestStep {
	localSentry.handleStatic(fmt.Sprintf("/projects/%s/display-name-%s/keys/", org, project), http.StatusOK, fmt.Sprintf(getClientKeyResponseBody, dsnname))

	return logicaltest.TestStep{
		Operation: logical.ReadOperation,
		Path:      "dsn/" + project,
		ErrorOk:   false,
		Check: func(resp *logical.Response) error {
			expect := map[string]interface{}{
				"name": dsnname,
				"dsn":  "https://test@sentry.io/2",
			}

			if !cmp.Equal(expect, resp.Data) {
				return fmt.Errorf("unexpected response. %s", cmp.Diff(expect, resp.Data))
			}

			return nil
		},
	}
}

func testReadDsn(org, project, dsnname string) logicaltest.TestStep {
	localSentry.handleStatic(fmt.Sprintf("/projects/%s/display-name-%s/keys/", org, project), http.StatusOK, fmt.Sprintf(getClientKeyResponseBody, dsnname))

	return logicaltest.TestStep{
		Operation: logical.ReadOperation,
		Path:      "dsn/" + project + "/" + dsnname,
		ErrorOk:   false,
		Check: func(resp *logical.Response) error {
			expect := map[string]interface{}{
				"name": dsnname,
				"dsn":  "https://test@sentry.io/2",
			}

			if !cmp.Equal(expect, resp.Data) {
				return fmt.Errorf("unexpected response. %s", cmp.Diff(expect, resp.Data))
			}

			return nil
		},
	}
}

const getClientKeyResponseBody = `
[{
    "dateCreated": "2018-11-06T21:20:07.941Z", 
    "dsn": {
      "cdn": "https://sentry.io/js-sdk-loader/cec9dfceb0b74c1c9a5e3c135585f364.min.js", 
      "csp": "https://sentry.io/api/2/csp-report/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364", 
      "minidump": "https://sentry.io/api/2/minidump/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364", 
      "public": "https://test@sentry.io/2", 
      "secret": "https://test-deprecated-dsn@sentry.io/2", 
      "security": "https://sentry.io/api/2/security/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364"
    }, 
    "id": "cec9dfceb0b74c1c9a5e3c135585f364", 
    "isActive": true, 
    "label": "%s", 
    "name": "Fabulous Key", 
    "projectId": 2, 
    "public": "cec9dfceb0b74c1c9a5e3c135585f364", 
    "rateLimit": null, 
    "secret": "4f6a592349e249c5906918393766718d"
}]
`
