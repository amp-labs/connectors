package write

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

type (
	RequestBuilder func(context.Context, common.WriteParams) (*http.Request, error)
	ResponseParser func(context.Context, common.WriteParams, *http.Response) (*common.WriteResult, error)
)

type SimpleWriteStrategy struct {
	Client               common.AuthenticatedHTTPClient
	createRequestBuilder RequestBuilder
	updateRequestBuilder RequestBuilder
	responseParser       ResponseParser
}

func NewSimpleWriteStrategy(
	client common.AuthenticatedHTTPClient,
	createRequestBuilder RequestBuilder,
	updateRequestBuilder RequestBuilder,
	responseParser ResponseParser,
) *SimpleWriteStrategy {
	return &SimpleWriteStrategy{
		Client:               client,
		createRequestBuilder: createRequestBuilder,
		updateRequestBuilder: updateRequestBuilder,
		responseParser:       responseParser,
	}
}

func (s *SimpleWriteStrategy) WriteObject(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	var requestBuilder RequestBuilder
	if len(config.RecordId) == 0 {
		requestBuilder = s.createRequestBuilder
	} else {
		requestBuilder = s.updateRequestBuilder
	}

	req, err := requestBuilder(ctx, config)
	if err != nil {
		return nil, err
	}

	rsp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	return s.responseParser(ctx, config, rsp)
}
