// Package gotocore handles GoTo's core API functionality.
// This includes endpoints for managing webinars and other products.
// The package name "gotocore" is internal shorthand —
// it does not imply core-only access.
package gotocore

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
)

type Adapter struct {
	*components.Connector
	components.SchemaProvider
	components.Reader
	components.Writer

	accountKey string
}

func NewAdapter(params common.ConnectorParams, accountKey string) (*Adapter, error) {
	adapter, err := components.Initialize(providers.GoTo, params, constructor)
	if err != nil {
		return nil, err
	}

	adapter.accountKey = accountKey

	return adapter, nil
}

func constructor(base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{Connector: base}

	adapter.SchemaProvider = schema.NewObjectSchemaProvider(
		adapter.HTTPClient().Client,
		schema.FetchModeParallel,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  adapter.buildSingleObjectMetadataRequest,
			ParseResponse: adapter.parseSingleObjectMetadataResponse,
		},
	)

	adapter.Reader = reader.NewHTTPReader(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  adapter.buildReadRequest,
			ParseResponse: adapter.parseReadResponse,
		},
	)

	adapter.Writer = writer.NewHTTPWriter(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  adapter.buildWriteRequest,
			ParseResponse: adapter.parseWriteResponse,
		},
	)

	return adapter, nil
}
