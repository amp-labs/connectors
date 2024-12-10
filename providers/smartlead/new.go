// nolint
package smartlead

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/connector"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/metadata"
	"github.com/amp-labs/connectors/providers"
)

// TODO: Could this be a part of the providerInfo object?
const apiVersion = "v1"

type Connector struct {
	components.ConnectorComponent
	components.MetadataStrategy
}

func NewConnector(params connector.Parameters) (conn *Connector, outErr error) {
	return components.Initialize(providers.Smartlead, params, setup)
}

func setup(connectorComponent *components.ConnectorComponent) (*Connector, error) {
	conn := &Connector{
		ConnectorComponent: *connectorComponent,
	}

	// Could be an OpenAPI strategy, or a custom one, if nothing fits.
	conn.MetadataStrategy = metadata.NewEndpointStrategy(
		conn.JSON.HTTPClient.Client,
		metadataRequestBuilder(conn),
		metadataResponseParser(conn),
	)

	// Behavior overrides can go in here.
	return conn, nil
}

// metadataRequestBuilder makes a request to sample an object.
func metadataRequestBuilder(conn *Connector) metadata.RequestBuilder {
	return func(ctx context.Context, object string) (*http.Request, error) {
		url, err := urlbuilder.New(conn.ProviderInfo().BaseURL, apiVersion, object)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
		if err != nil {
			return nil, err
		}

		return req, nil
	}
}

// metadataResponseParser parses the response to get metadata.
func metadataResponseParser(conn *Connector) metadata.ResponseParser {
	return func(ctx context.Context, response *http.Response) (*common.ObjectMetadata, error) {
		// Parse the response and return the metadata.
		return nil, nil
	}
}
