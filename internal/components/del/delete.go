package del

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

type (
	RequestBuilder func(context.Context, common.DeleteParams) (*http.Request, error)
	ResponseParser func(context.Context, common.DeleteParams, *http.Response) (*common.DeleteResult, error)
)

var ErrFailedToDeleteObject = errors.New("failed to delete object")

type SimpleDeleteStrategy struct {
	Client         common.AuthenticatedHTTPClient
	requestBuilder RequestBuilder
}

func NewSimpleDeleteStrategy(client common.AuthenticatedHTTPClient, reqBuilder RequestBuilder) *SimpleDeleteStrategy {
	return &SimpleDeleteStrategy{
		Client:         client,
		requestBuilder: reqBuilder,
	}
}

func (s *SimpleDeleteStrategy) DeleteObject(
	ctx context.Context,
	params common.DeleteParams,
) (*common.DeleteResult, error) {
	req, err := s.requestBuilder(ctx, params)
	if err != nil {
		return nil, err
	}

	rsp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode == http.StatusNoContent || rsp.StatusCode == http.StatusOK {
		return &common.DeleteResult{
			Success: true,
		}, nil
	}

	// Read the body for more information, if any
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	return nil, fmt.Errorf("%w: %v: %s", ErrFailedToDeleteObject, rsp.StatusCode, string(body))
}
