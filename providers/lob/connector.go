package lob

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers/lob/internal/core"
	"github.com/amp-labs/connectors/providers/lob/internal/metadata"
)

type Connector struct {
	*core.Base

	// Supported operations
	components.SchemaProvider
	components.Reader
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	base, err := core.NewBase(params)
	if err != nil {
		return nil, err
	}

	connector := &Connector{
		Base:           base,
		SchemaProvider: schema.NewOpenAPISchemaProvider(base.ProviderContext.Module(), metadata.Schemas),
	}

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  base.GetErrorHandler(),
		},
	)

	return connector, nil
}
