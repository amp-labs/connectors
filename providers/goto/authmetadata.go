package gotoconn

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	accountKey, err := c.retrieveAccountKey(ctx)
	if err != nil {
		return nil, err
	}

	if accountKey == "" {
		return nil, common.ErrMissingExpectedValues
	}

	c.accountKey = accountKey

	return &common.PostAuthInfo{
		CatalogVars: AuthMetadataVars{
			AccountKey: accountKey,
		}.AsMap(),
		RawResponse: nil,
	}, nil
}

func (c *Connector) retrieveAccountKey(ctx context.Context) (string, error) {
	url, err := c.getMeURL()
	if err != nil {
		return "", err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return "", err
	}

	if _, ok := resp.Body(); !ok {
		return "", common.ErrEmptyJSONHTTPResponse
	}

	data, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return "", common.ErrFailedToUnmarshalBody
	}

	rawAccountKey, ok := (*data)["accountKey"]
	if !ok {
		return "", common.ErrMissingExpectedValues
	}

	switch v := rawAccountKey.(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatInt(int64(v), 10), nil
	default:
		return "", common.ErrMissingExpectedValues
	}
}

func (c *Connector) getMeURL() (*urlbuilder.URL, error) {
	// /me lives on the goTo (api.getgo.com) module, so we resolve that
	// module's BaseURL explicitly — even when the connector was created
	// with the goToConnect module selected.
	baseURL := c.ProviderInfo().ReadModuleInfo(providers.ModuleGoTo).BaseURL

	return urlbuilder.New(baseURL, "/admin/rest/v1/me")
}
