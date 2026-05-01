package gotoconn

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
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

	catalogVars := map[string]string{
		"accountKey": accountKey,
	}

	return &common.PostAuthInfo{
		CatalogVars: &catalogVars,
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

	data, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return "", common.ErrFailedToUnmarshalBody
	}

	if data == nil {
		return "", common.ErrMissingExpectedValues
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
	return urlbuilder.New(c.ProviderInfo().BaseURL, "/admin/rest/v1/me")
}
