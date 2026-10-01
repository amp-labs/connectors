package confluence

import (
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
)

type Adapter struct {
	components.BaseConnector
	components.SchemaProvider
}

func NewAdapter(base *components.BaseConnector) (*Adapter, error) {
	adapter := &Adapter{
		BaseConnector: *base,
	}

	adapter.SchemaProvider = schema.NewOpenAPISchemaProvider(adapter.ProviderContext.Module(), Schemas)

	return adapter, nil
}
