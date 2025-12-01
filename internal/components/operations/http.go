package operations

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrNoResponse     = errors.New("no response")
)

// HTTPOperation provides a generic implementation for HTTP-based operations like read, write, delete, etc.
type HTTPOperation[RequestType any, ResponseType any] struct {
	client   common.AuthenticatedHTTPClient
	handlers HTTPHandlers[RequestType, ResponseType]
}

// HTTPHandlers contains operation-specific HTTP handlers for building and parsing HTTP requests and responses.
type HTTPHandlers[RequestType any, ResponseType any] struct {
	BuildRequest  func(context.Context, RequestType) (*http.Request, error)
	ParseResponse func(context.Context, RequestType, *http.Request, *common.JSONHTTPResponse) (ResponseType, error)
	ErrorHandler  func(*http.Response, []byte) error
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

// nolint:ireturn,cyclop
func (op *HTTPOperation[RequestType, ResponseType]) ExecuteRequest(
	ctx context.Context,
	params RequestType,
) (ResponseType, error) {
	var response ResponseType

	req, err := op.handlers.BuildRequest(ctx, params)
	if err != nil {
		return response, err
	}

	if req == nil {
		return response, ErrInvalidRequest
	}

	req = common.AddJSONContentTypeIfNotPresent(req)

	resp, err := op.client.Do(req)
	if err != nil {
		return response, err
	}

	if resp == nil {
		return response, ErrNoResponse
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}
	// Check the response status code
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		if op.handlers.ErrorHandler != nil {
			err = op.handlers.ErrorHandler(resp, body)
			if err != nil {
				return response, err
			}
		}

		return response, common.InterpretError(resp, body)
	}

	jsonResp, err := common.ParseJSONResponse(resp, body)
	if err != nil {
		return response, err
	}

	return op.handlers.ParseResponse(ctx, params, req, jsonResp)
}
