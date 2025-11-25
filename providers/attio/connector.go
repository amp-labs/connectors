package attio

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const (
	apiVersion = "v2"
)

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
}

// GetRecordsByIds implements connectors.SubscribeConnector.
func (c *Connector) GetRecordsByIds(ctx context.Context, objectName string, recordIds []string, fields []string, associations []string) ([]common.ReadResultRow, error) {
	panic("unimplemented")
}

// UpdateSubscription implements connectors.SubscribeConnector.
func (c *Connector) UpdateSubscription(ctx context.Context, params common.SubscribeParams, previousResult *common.SubscriptionResult) (*common.SubscriptionResult, error) {
	panic("unimplemented")
}

// VerifyWebhookMessage implements connectors.SubscribeConnector.
func (c *Connector) VerifyWebhookMessage(ctx context.Context, request *common.WebhookRequest, params *common.VerificationParams) (bool, error) {
	panic("unimplemented")
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: params.Client.Caller,
		},
	}

	// Read provider info
	providerInfo, err := providers.ReadInfo(conn.Provider())
	if err != nil {
		return nil, err
	}

	conn.setBaseURL(providerInfo.BaseURL)

	return conn, nil
}

// Provider returns the connector provider.
func (c *Connector) Provider() providers.Provider {
	return providers.Attio
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) getApiURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, apiVersion, arg)
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
