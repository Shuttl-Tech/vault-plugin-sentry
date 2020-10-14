package backend

import (
	"context"
	"github.com/atlassian/go-sentry-api"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const KeyDsnPrefix = "dsn/"

type SentryDsn struct {
	Name string `json:"name"`
	DSN  string `json:"dsn"`
}

func (d *SentryDsn) Data() map[string]interface{} {
	return map[string]interface{}{
		"name": d.Name,
		"dsn":  d.DSN,
	}
}

func loadDsn(ctx context.Context, storage logical.Storage, project, label string) (*SentryDsn, error) {
	entry, err := storage.Get(ctx, KeyDsnPrefix+project+"/"+label)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	item := new(SentryDsn)
	err = entry.DecodeJSON(item)
	if err != nil {
		return nil, err
	}

	return item, err
}

func handleDsnRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	vaultProjectName := data.Get("project").(string)
	dsnName := data.Get("name").(string)

	vaultProject, err := loadProject(ctx, req.Storage, vaultProjectName)
	if err != nil {
		return nil, err
	}

	if vaultProject == nil {
		return logical.ErrorResponse("project %s is not configured", vaultProjectName), nil
	}

	if dsnName == "" {
		dsnName = vaultProject.DefaultDsnLabel
	}

	if dsnName == "" {
		return logical.ErrorResponse("default DSN label is not set for project %s", vaultProjectName), nil
	}

	dsn, err := loadDsn(ctx, req.Storage, vaultProjectName, dsnName)
	if err != nil {
		return nil, err
	}

	if dsn != nil {
		return &logical.Response{
			Data: dsn.Data(),
		}, nil
	}

	config, err := loadConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return logical.ErrorResponse("plugin is not configured"), nil
	}

	client, err := config.Client()
	if err != nil {
		return nil, err
	}

	key, err := fetchKeyOrMakeNew(
		client,
		sentry.Organization{Slug: &config.Name},
		sentry.Project{Slug: &vaultProject.DisplayName},
		dsnName,
	)

	if err != nil {
		return logical.ErrorResponse("failed to retrieve client keys from sentry. %s", err), nil
	}

	item := &SentryDsn{
		Name: key.Label,
		DSN:  key.DSN.Public,
	}

	entry, err := logical.StorageEntryJSON(KeyDsnPrefix+vaultProjectName+"/"+dsnName, item)
	if err != nil {
		return nil, err
	}

	err = req.Storage.Put(ctx, entry)
	if err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: item.Data(),
	}, nil
}

func fetchKeyOrMakeNew(client *sentry.Client, org sentry.Organization, project sentry.Project, label string) (*sentry.Key, error) {
	keys, err := client.GetClientKeys(org, project)
	if err != nil {
		return nil, err
	}

	for _, k := range keys {
		if k.Label == label {
			return &k, nil
		}
	}

	key, err := client.CreateClientKey(org, project, label)
	if err != nil {
		return nil, err
	}

	return &key, nil
}
