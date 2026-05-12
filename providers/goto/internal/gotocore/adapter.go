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
	"github.com/amp-labs/connectors/providers"
)

type Adapter struct {
	*components.Connector
	components.SchemaProvider
	components.Reader

	accountKey string
}

const (
	queryParamSize     = "size"
	queryParamPageSize = "pageSize"
	sampleSize         = "1"
	readPageSize       = "200"
	// corporatePageSize caps Corporate API page size at 100 (its documented
	// maximum) and lets corporateNextPage detect the last page by counting
	// returned records.
	corporatePageSize = 100
	queryParamPage    = "page"
	queryParamOffset  = "offset"

	// metadataSampleWindowDays is the size in days of the time-range filter
	// applied when sampling records for schema. Wide enough to
	// catch at least one record on endpoints that mandate a
	// time-range filter.
	metadataSampleWindowDays = 120
)

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

	return adapter, nil
}
