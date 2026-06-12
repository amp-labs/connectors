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
	conn, err := components.Initialize(providers.Breezy, params, constructor)
	if err != nil {
		return nil, err
	}

	if params.Metadata != nil {
		conn.CompanyID = params.Metadata["company_id"]
	}

	return conn, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	return connector, nil
}
