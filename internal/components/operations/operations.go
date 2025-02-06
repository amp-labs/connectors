package operations

import (
	"context"
	"io"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

// HTTPOperation provides a generic implementation for HTTP-based operations like read, write, delete, etc.
type HTTPOperation[RequestType any, ResponseType any] struct {
	client   common.AuthenticatedHTTPClient
	handlers HTTPHandlers[RequestType, ResponseType]
}

// HTTPHandlers contains operation-specific HTTP handlers for building and parsing HTTP requests and responses.
type HTTPHandlers[RequestType any, ResponseType any] struct {
	BuildRequest  func(context.Context, RequestType) (*http.Request, error)
	ParseResponse func(context.Context, RequestType, *common.JSONHTTPResponse) (ResponseType, error)
}

func NewHTTPOperation[RequestType any, ResponseType any](
	client common.AuthenticatedHTTPClient,
	handlers HTTPHandlers[RequestType, ResponseType],
) *HTTPOperation[RequestType, ResponseType] {
	return &HTTPOperation[RequestType, ResponseType]{
		client:   client,
		handlers: handlers,
	}
}

// nolint:ireturn
func (op *HTTPOperation[RequestType, ResponseType]) ExecuteRequest(
	ctx context.Context,
	params RequestType,
) (ResponseType, error) {
	var response ResponseType

	req, err := op.handlers.BuildRequest(ctx, params)
	if err != nil {
		return response, err
	}

	resp, err := op.client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	jsonResp, err := common.ParseJSONResponse(resp, body)
	if err != nil {
		return response, err
	}

	return op.handlers.ParseResponse(ctx, params, jsonResp)
}
