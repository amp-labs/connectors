package core

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v1"

type Base struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient
}

func NewBase(params common.ConnectorParams) (*Base, error) {
	return components.Init(providers.Stripe, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Base, error) {
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}.Handle
	base.SetErrorHandler(errorHandler)

	return &Base{
		Connector: base,
	}, nil
}

func (b *Base) GetURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, objectName)
}
