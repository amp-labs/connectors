package bill

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	sessionId, err := c.retrieveSessionId(ctx)
	if err != nil {
		return nil, err
	}

	c.sessionId = sessionId

	catalogVars := map[string]string{
		"sessionId": sessionId,
	}

	return &common.PostAuthInfo{
		CatalogVars: &catalogVars,
	}, nil
}

func (c *Connector) retrieveSessionId(ctx context.Context) (string, error) {
	url, err := c.getSessionURL()
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

	sessionId, ok := firstTenant.(map[string]any)["sessionId"].(string)
	if !ok {
		return "", common.ErrMissingExpectedValues
	}

	return sessionId, nil
}

func (c *Connector) getSessionURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, "/connect/v3/login")
}
