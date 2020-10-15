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

func TestHandleProject(t *testing.T) {
	logicaltest.Test(t, logicaltest.TestCase{
		LogicalBackend: testGetBackend(t),
		Steps: []logicaltest.TestStep{
			testReadProjectErr("unregistered", "project unregistered is not configured in Vault"),
			testWriteProjectErr("test-project", "test-team", "plugin is not configured"),
			testWriteConfig("project-org", "token", localSentry.url, 10),
			testWriteProjectExisting("project-org", "existing-project", "test-team", ""),
			testReadProject("existing-project", "display-name-existing-project", "project-org", "test-team", ""),
			testWriteProjectFresh("project-org", "fresh-project", "test-team", ""),
			testWriteProjectWithSentryName("project-org", "frs", "free-ride-service", "my-team", ""),
			testReadProject("fresh-project", "display-name-fresh-project", "project-org", "test-team", ""),
			testListProjects("existing-project", "fresh-project", "frs"),
			testWriteProjectExisting("project-org", "project-with-default-dsn", "test-team", "default-dsn-for-tests"),
			testReadProject("project-with-default-dsn", "display-name-project-with-default-dsn", "project-org", "test-team", "default-dsn-for-tests"),
			testDeleteProject("project-with-default-dsn"),
		},
	})
}

func testListProjects(names ...string) logicaltest.TestStep {
	return logicaltest.TestStep{
		Operation: logical.ListOperation,
		Path:      "projects",
		Check: func(resp *logical.Response) error {
			if !cmp.Equal(names, resp.Data["keys"]) {
				return fmt.Errorf("unexpected list result. %s", cmp.Diff(names, resp.Data))
			}
			return nil
		},
	}
}

func testWriteProjectExisting(org, name, team, dsnLabel string) logicaltest.TestStep {
	localSentry.handleStatic("/projects/"+org+"/"+name+"/", http.StatusOK, fmt.Sprintf(getProjectResponseBody, "display-name-"+name))

	return logicaltest.TestStep{
		Operation: logical.UpdateOperation,
		Path:      "project/" + name,
		ErrorOk:   false,
		Data: map[string]interface{}{
			"team":              team,
			"default_dsn_label": dsnLabel,
		},
		Check: func(resp *logical.Response) error {
			expect := map[string]interface{}{
				"name":              name,
				"display_name":      "display-name-" + name,
				"team":              team,
				"org":               org,
				"default_dsn_label": dsnLabel,
			}

			if !cmp.Equal(expect, resp.Data) {
				return fmt.Errorf("unexpected data in response. %s", cmp.Diff(expect, resp.Data))
			}

			return nil
		},
	}
}

func testWriteProjectFresh(org, name, team, dsnName string) logicaltest.TestStep {
	localSentry.handleStatic("/projects/"+org+"/"+name+"/", http.StatusOK, fmt.Sprintf(getProjectResponseBody, "display-name-"+name))
	localSentry.handleStatic(fmt.Sprintf("/teams/%s/%s/projects/", org, team), http.StatusOK, fmt.Sprintf(getProjectResponseBody, "display-name-"+name))

	return logicaltest.TestStep{
		Operation: logical.UpdateOperation,
		Path:      "project/" + name,
		ErrorOk:   false,
		Data: map[string]interface{}{
			"team":              team,
			"default_dsn_label": dsnName,
		},
		Check: func(resp *logical.Response) error {
			expect := map[string]interface{}{
				"name":              name,
				"display_name":      "display-name-" + name,
				"team":              team,
				"org":               org,
				"default_dsn_label": dsnName,
			}

			if !cmp.Equal(expect, resp.Data) {
				return fmt.Errorf("unexpected data in response. %s", cmp.Diff(expect, resp.Data))
			}

			return nil
		},
	}
}

func testWriteProjectWithSentryName(org, name, sentryName, team, dsnName string) logicaltest.TestStep {
	localSentry.handleStatic("/projects/"+org+"/"+name+"/", http.StatusOK, fmt.Sprintf(getProjectResponseBody, sentryName))
	localSentry.handleStatic(fmt.Sprintf("/teams/%s/%s/projects/", org, team), http.StatusOK, fmt.Sprintf(getProjectResponseBody, sentryName))

	return logicaltest.TestStep{
		Operation: logical.UpdateOperation,
		Path:      "project/" + name,
		ErrorOk:   false,
		Data: map[string]interface{}{
			"team":              team,
			"default_dsn_label": dsnName,
		},
		Check: func(resp *logical.Response) error {
			expect := map[string]interface{}{
				"name":              name,
				"display_name":      sentryName,
				"team":              team,
				"org":               org,
				"default_dsn_label": dsnName,
			}

			if !cmp.Equal(expect, resp.Data) {
				return fmt.Errorf("unexpected data in response. %s", cmp.Diff(expect, resp.Data))
			}

			return nil
		},
	}
}

func testWriteProjectErr(name string, team string, msg string) logicaltest.TestStep {
	return logicaltest.TestStep{
		Operation: logical.UpdateOperation,
		Path:      "project/" + name,
		ErrorOk:   true,
		Data: map[string]interface{}{
			"team": team,
		},
		Check: func(resp *logical.Response) error {
			if !resp.IsError() {
				return fmt.Errorf("expected error in write response. got none")
			}

			if !strings.Contains(resp.Error().Error(), msg) {
				return fmt.Errorf("unexpected error %q does not match %q", resp.Error(), msg)
			}

			return nil
		},
	}
}

func testReadProject(name, displayName, org, team, dsnLabel string) logicaltest.TestStep {
	return logicaltest.TestStep{
		Operation: logical.ReadOperation,
		Path:      "project/" + name,
		Check: func(resp *logical.Response) error {
			expect := map[string]interface{}{
				"name":              name,
				"display_name":      displayName,
				"team":              team,
				"org":               org,
				"default_dsn_label": dsnLabel,
			}

			if !cmp.Equal(expect, resp.Data) {
				return fmt.Errorf("unexpected data in read response. %s", cmp.Diff(expect, resp.Data))
			}

			return nil
		},
	}
}

func testReadProjectErr(name, msg string) logicaltest.TestStep {
	return logicaltest.TestStep{
		Operation: logical.ReadOperation,
		Path:      "project/" + name,
		ErrorOk:   true,
		Check: func(resp *logical.Response) error {
			if !resp.IsError() {
				return fmt.Errorf("expected error in read response, got none")
			}

			if !strings.Contains(resp.Error().Error(), msg) {
				return fmt.Errorf("unexpected error message %q does not match %q", resp.Error(), msg)
			}

			return nil
		},
	}
}

func testDeleteProject(name string) logicaltest.TestStep {
	return logicaltest.TestStep{
		Operation: logical.DeleteOperation,
		Path:      "project/" + name,
		Check: func(resp *logical.Response) error {
			expect := map[string]interface{}{
				logical.HTTPContentType: "application/json",
				logical.HTTPStatusCode:  http.StatusOK,
			}

			if !cmp.Equal(expect, resp.Data) {
				return fmt.Errorf("unexpected data in read response. %s", cmp.Diff(expect, resp.Data))
			}

			return nil
		},
	}
}

const getProjectResponseBody = `
{
  "name": "%s"
}
`
