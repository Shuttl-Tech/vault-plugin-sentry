package backend

import (
	"context"
	"github.com/atlassian/go-sentry-api"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"net/http"
)

const KeyProjectConfigPrefix = "projects/"

type SentryProject struct {
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	Team            string `json:"team"`
	Org             string `json:"org"`
	DefaultDsnLabel string `json:"default_dsn_label"`
}

func (p *SentryProject) Data() map[string]interface{} {
	return map[string]interface{}{
		"name":              p.Name,
		"display_name":      p.DisplayName,
		"team":              p.Team,
		"org":               p.Org,
		"default_dsn_label": p.DefaultDsnLabel,
	}
}

func loadProject(ctx context.Context, storage logical.Storage, name string) (*SentryProject, error) {
	entry, err := storage.Get(ctx, KeyProjectConfigPrefix+name)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	item := new(SentryProject)
	err = entry.DecodeJSON(item)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func handleProjectRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	projectName := data.Get("project").(string)
	project, err := loadProject(ctx, req.Storage, projectName)
	if err != nil {
		return nil, err
	}

	if project == nil {
		return logical.ErrorResponse("project %s is not configured in Vault", projectName), nil
	}

	return &logical.Response{
		Data: project.Data(),
	}, nil
}

func handleProjectsList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	items, err := req.Storage.List(ctx, KeyProjectConfigPrefix)
	if err != nil {
		return nil, err
	}

	return logical.ListResponse(items), nil
}

func handleProjectUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	projectName := data.Get("project").(string)
	teamName := data.Get("team").(string)
	defaultDsnLabel := data.Get("default_dsn_label").(string)

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

	// Attempt to read project from sentry or create a new one
	// if the project does not exist.
	p, err := client.GetProject(sentry.Organization{
		Slug: &config.Name,
	}, projectName)

	if err != nil {
		apiErr, ok := err.(sentry.APIError)
		if !ok {
			return logical.ErrorResponse("failed to read project information from sentry. %s", err), nil
		}

		if apiErr.StatusCode != http.StatusNotFound {
			return logical.ErrorResponse("failed to read project information from sentry. %s", apiErr), nil
		}

		p, err = client.CreateProject(
			sentry.Organization{Slug: &config.Name},
			sentry.Team{Slug: &teamName},
			projectName,
			nil,
		)

		if err != nil {
			return logical.ErrorResponse("failed to create new project in sentry. %s", err), nil
		}
	}

	item := &SentryProject{
		Name:            projectName,
		DisplayName:     p.Name,
		Org:             config.Name,
		Team:            teamName,
		DefaultDsnLabel: defaultDsnLabel,
	}

	entry, err := logical.StorageEntryJSON(KeyProjectConfigPrefix+projectName, item)
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
