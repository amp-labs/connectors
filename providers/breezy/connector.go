package breezy

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/breezy/metadata"
)

type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient

	components.SchemaProvider

	// CompanyID scopes company-level API paths (e.g. positions, webhook endpoints).
	CompanyID string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Breezy, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	if params.Metadata != nil {
		connector.CompanyID = params.Metadata["company_id"]
	}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	return connector, nil
}
