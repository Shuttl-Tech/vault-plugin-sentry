package backend

import (
	"context"
	"github.com/Shuttl-Tech/vault-plugin-sentry/version"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func handleInfoRead(context.Context, *logical.Request, *framework.FieldData) (*logical.Response, error) {
	return &logical.Response{
		Data: map[string]interface{}{
			"description": "Manage Sentry projects and their DSN",
			"commit_sha":  version.GitCommit,
			"version":     version.HumanVersion,
		},
	}, nil
}
