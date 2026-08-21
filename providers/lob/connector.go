package lob

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers/lob/internal/core"
	"github.com/amp-labs/connectors/providers/lob/internal/metadata"
)

type Connector struct {
	*core.Base

	// Supported operations
	components.SchemaProvider
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

	return connector, nil
}
