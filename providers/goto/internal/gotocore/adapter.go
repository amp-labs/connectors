// Package gotocore handles GoTo's core API functionality.
// This includes endpoints for managing webinars and other products.
// The package name "gotocore" is internal shorthand —
// it does not imply core-only access.
package gotocore

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

type Adapter struct {
	*components.Connector
	components.SchemaProvider

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

func (a *Adapter) SetAccountKey(accountKey string) {
	a.accountKey = accountKey
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

	return adapter, nil
}
