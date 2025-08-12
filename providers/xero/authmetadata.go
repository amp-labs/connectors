package xero

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	tenantId, err := c.retrieveTenantId(ctx)
	if err != nil {
		return nil, err
	}

	c.tenantId = tenantId

	catalogVars := map[string]string{
		"tenantId": tenantId,
	}

	return &common.PostAuthInfo{
		CatalogVars: &catalogVars,
	}, nil
}

func (c *Connector) retrieveTenantId(ctx context.Context) (string, error) {
	url, err := c.getTenantURL()
	if err != nil {
		return "", err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return "", err
	}

	data, err := common.UnmarshalJSON[[]any](resp)
	if err != nil {
		return "", common.ErrFailedToUnmarshalBody
	}

	if len(*data) == 0 {
		return "", common.ErrMissingExpectedValues
	}

	firstTenant := (*data)[0]

	tenantId, ok := firstTenant.(map[string]any)["tenantId"].(string)
	if !ok {
		return "", common.ErrMissingExpectedValues
	}

	return tenantId, nil
}

func (c *Connector) getTenantURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, "/connections")
}
