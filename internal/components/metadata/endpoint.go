package metadata

import (
	"context"
	"io"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

type RequestBuilder func(ctx context.Context, object string) (*http.Request, error)
type ResponseParser func(ctx context.Context, response *http.Response) (*common.ObjectMetadata, error)

// EndpointStrategy is a ObjectMetadataStrategy that uses an endpoint to fetch metadata for
// an object. It uses an HTTP client to send requests and parse responses. The endpoint could be an introspection
// endpoint, or a list endpoint from where sample data can be read for metadata extraction.
type EndpointStrategy struct {
	Client    common.AuthenticatedHTTPClient
	requester RequestBuilder
	parser    ResponseParser
}

// NewEndpointStrategy creates a new EndpointStrategy.
func NewEndpointStrategy(client common.AuthenticatedHTTPClient, requester RequestBuilder, parser ResponseParser) *EndpointStrategy {
	return &EndpointStrategy{
		Client:    client,
		requester: requester,
		parser:    parser,
	}
}

func (i *EndpointStrategy) Run(ctx context.Context, object string) (*common.ObjectMetadata, error) {
	req, err := i.requester(ctx, object)
	if err != nil {
		return nil, err
	}

	resp, err := i.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	return i.parser(ctx, resp)
}
