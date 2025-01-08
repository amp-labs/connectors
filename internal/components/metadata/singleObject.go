package metadata

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

type (
	RequestBuilder func(ctx context.Context, object string) (*http.Request, error)
	ResponseParser func(ctx context.Context, response *http.Response) (*common.ObjectMetadata, error)
)

type SingleObjectEndpointStrategy struct {
	Client         common.AuthenticatedHTTPClient
	requestBuilder RequestBuilder
	responseParser ResponseParser
}

// NewSingleObjectEndpointStrategy creates a new SingleObjectEndpointStrategy.
func NewSingleObjectEndpointStrategy(
	client common.AuthenticatedHTTPClient,
	requestBuilder RequestBuilder,
	responseParser ResponseParser,
) *SingleObjectEndpointStrategy {
	return &SingleObjectEndpointStrategy{
		Client:         client,
		requestBuilder: requestBuilder,
		responseParser: responseParser,
	}
}

func (i *SingleObjectEndpointStrategy) String() string {
	return "metadata.SingleObjectEndpointStrategy"
}

func (i *SingleObjectEndpointStrategy) GetObjectMetadata(
	ctx context.Context,
	objects ...string,
) (*common.ListObjectMetadataResult, error) {
	result := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, object := range objects {
		req, err := i.requestBuilder(ctx, object)
		if err != nil {
			return nil, err
		}

		resp, err := i.Client.Do(req)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		metadata, err := i.responseParser(ctx, resp)
		if err != nil {
			return nil, err
		}

		result.Result[object] = *metadata
	}

	return &result, nil
}
