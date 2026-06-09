package zoominfo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

const (
	// jsonAPIMediaType is the media type ZoomInfo's Data API speaks. Several
	// endpoints (e.g. intent, lookup) reject "application/json" with a 406, so
	// every request must advertise JSON:API for both Accept and Content-Type.
	jsonAPIMediaType = "application/vnd.api+json"

	// metadataPageSize limits sampling requests to a single record — that's all
	// we need to infer an object's fields.
	metadataPageSize = "1"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// Supported operations
	components.SchemaProvider
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.ZoomInfo, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	// ZoomInfo has no OpenAPI spec and no GET list endpoints, so object metadata
	// is derived by sampling a single record per object and inferring field types.
	// FetchModeSerial is deliberate: ZoomInfo rate-limits aggressively (25 rps by
	// default) and firing one request per object in parallel reliably triggers 429s.
	connector.SchemaProvider = schema.NewObjectSchemaProvider(
		connector.HTTPClient().Client,
		schema.FetchModeSerial,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  connector.buildSingleObjectMetadataRequest,
			ParseResponse: connector.parseSingleObjectMetadataResponse,
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
			}.Handle,
		},
	)

	return connector, nil
}
