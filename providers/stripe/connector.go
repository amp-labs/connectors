package stripe

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v1"

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Stripe, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
	}

	base.SetErrorHandler(interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}.Handle)

	return connector, nil
}

func (c *Connector) getURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName)
}

// https://docs.stripe.com/api/connected-accounts
func makeConnectedAccountHeader(accountID string) common.Header {
	return common.Header{
		Key:   "Stripe-Account",
		Value: accountID,
	}
}
