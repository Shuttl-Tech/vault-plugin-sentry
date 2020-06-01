package backend

import (
	"context"
	"github.com/atlassian/go-sentry-api"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"strings"
)

const KeyConfig string = "config"

type SentryOrg struct {
	Name              string `json:"name"`
	DisplayName       string `json:"display_name"`
	ApiToken          string `json:"api_token"`
	Endpoint          string `json:"endpoint"`
	ConnectionTimeout int    `json:"connection_timeout"`
}

func (o *SentryOrg) Data() map[string]interface{} {
	return map[string]interface{}{
		"name":         o.Name,
		"display_name": o.DisplayName,
		"endpoint":     o.Endpoint,
		"timeout":      o.ConnectionTimeout,
	}
}

func (o *SentryOrg) Client() (*sentry.Client, error) {
	client, err := sentry.NewClient(o.ApiToken, &o.Endpoint, &o.ConnectionTimeout)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func handleConfigUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	orgSlug := data.Get("org").(string)
	token := data.Get("token").(string)
	endpoint := data.Get("endpoint").(string)
	timeout := data.Get("timeout").(int)

	endpoint = strings.TrimRight(endpoint, "/") + "/"

	uo := &SentryOrg{
		ApiToken:          token,
		Endpoint:          endpoint,
		ConnectionTimeout: timeout,
	}

	client, err := uo.Client()
	if err != nil {
		return logical.ErrorResponse("failed to initialize sentry client with given configuration. %s", err), nil
	}

	org, err := client.GetOrganization(orgSlug)
	if err != nil {
		return logical.ErrorResponse("failed to retrieve organization details from sentry. %s", err), nil
	}

	item := new(SentryOrg)
	if org.Slug != nil {
		item.Name = *org.Slug
	}

	item.DisplayName = org.Name
	item.ApiToken = token
	item.Endpoint = endpoint
	item.ConnectionTimeout = timeout

	entry, err := logical.StorageEntryJSON(KeyConfig, item)
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

func loadConfig(ctx context.Context, storage logical.Storage) (*SentryOrg, error) {
	entry, err := storage.Get(ctx, KeyConfig)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	item := new(SentryOrg)
	err = entry.DecodeJSON(item)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func handleConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	item, err := loadConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return logical.ErrorResponse("plugin is not configured"), nil
	}

	return &logical.Response{
		Data: item.Data(),
	}, nil
}
