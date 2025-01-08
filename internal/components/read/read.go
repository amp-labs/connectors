package read

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

type (
	RequestBuilder func(context.Context, common.ReadParams) (*http.Request, error)
	ResponseParser func(context.Context, common.ReadParams, *http.Response) (*common.ReadResult, error)
)

type SimpleReadStrategy struct {
	Client         common.AuthenticatedHTTPClient
	requestBuilder RequestBuilder
	responseParser ResponseParser
}

func NewSimpleReadStrategy(c common.AuthenticatedHTTPClient, rb RequestBuilder, rp ResponseParser) *SimpleReadStrategy {
	return &SimpleReadStrategy{
		Client:         c,
		requestBuilder: rb,
		responseParser: rp,
	}
}

func (s *SimpleReadStrategy) ReadObject(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	req, err := s.requestBuilder(ctx, config)
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
