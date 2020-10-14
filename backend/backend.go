package backend

import (
	"context"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const SecretTypeCreds = "creds"

type backend struct {
	*framework.Backend
}

func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
	b := New(c)
	err := b.Setup(ctx, c)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func New(c *logical.BackendConfig) *backend {
	b := new(backend)

	b.Backend = &framework.Backend{
		BackendType: logical.TypeLogical,
		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{"info"},
		},
		Paths: []*framework.Path{
			{
				Pattern: "info",
				Operations: map[logical.Operation]framework.OperationHandler{
					logical.ReadOperation: &framework.PathOperation{
						Callback: handleInfoRead,
					},
				},
			},
			{
				Pattern: "config/?$",
				Fields: map[string]*framework.FieldSchema{
					"org": {
						Type:        framework.TypeString,
						Required:    true,
						Description: "Slug of the sentry organization",
					},
					"token": {
						Type:        framework.TypeString,
						Required:    true,
						Description: "Sentry API token",
					},
					"endpoint": {
						Type:        framework.TypeString,
						Default:     "https://sentry.io/api/0/",
						Description: "Sentry endpoint to connect with",
					},
					"timeout": {
						Type:        framework.TypeInt,
						Default:     10,
						Description: "Connection timeout for API requests",
					},
				},
				Operations: map[logical.Operation]framework.OperationHandler{
					logical.ReadOperation: &framework.PathOperation{
						Callback: handleConfigRead,
					},
					logical.UpdateOperation: &framework.PathOperation{
						Callback: handleConfigUpdate,
					},
				},
			},
			{
				Pattern: "projects/?",
				Operations: map[logical.Operation]framework.OperationHandler{
					logical.ListOperation: &framework.PathOperation{
						Callback: handleProjectsList,
					},
				},
			},
			{
				Pattern: "project/" + framework.GenericNameRegex("project") + "/?$",
				Fields: map[string]*framework.FieldSchema{
					"project": {
						Type:        framework.TypeString,
						Required:    true,
						Description: "Name of the project in Vault",
					},
					"team": {
						Type:        framework.TypeString,
						Required:    true,
						Description: "Name of the team that owns the project",
					},
					"default_dsn_label": {
						Type:        framework.TypeString,
						Required:    false,
						Description: "DSN label to use by default when not specified",
					},
					"sentry_project": {
						Type:        framework.TypeString,
						Required:    false,
						Description: "Name of the project in sentry",
					},
				},
				Operations: map[logical.Operation]framework.OperationHandler{
					logical.ReadOperation: &framework.PathOperation{
						Callback: handleProjectRead,
					},
					logical.UpdateOperation: &framework.PathOperation{
						Callback: handleProjectUpdate,
					},
					logical.DeleteOperation: &framework.PathOperation{
						Callback: handleProjectDelete,
					},
				},
			},
			{
				Pattern: "dsn/" + framework.GenericNameRegex("project") + framework.OptionalParamRegex("name"),
				Fields: map[string]*framework.FieldSchema{
					"project": {
						Type:        framework.TypeString,
						Required:    true,
						Description: "Name of the project in sentry",
					},
					"name": {
						Type:        framework.TypeString,
						Required:    false,
						Description: "Name of the DSN",
					},
				},
				Operations: map[logical.Operation]framework.OperationHandler{
					logical.ReadOperation: &framework.PathOperation{
						Callback: handleDsnRead,
					},
				},
			},
		},
	}

	return b
}
