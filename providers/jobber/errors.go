package jobber

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// Implement error abstraction layers to streamline provider error handling.
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

// ResponseError represents an error response from the Jobber API.
type ResponseError struct {
	Errors []ErrorDetails `json:"errors"`
}

type ErrorDetails struct {
	Message    string `json:"message,omitempty"`
	Locations  any    `json:"locations,omitempty"`
	Path       any    `json:"path,omitempty"`
	Extensions any    `json:"extensions,omitempty"`
}

func (r ResponseError) CombineErr(base error) error {
	if len(r.Errors) == 0 {
		return base
	}

	messages := make([]string, len(r.Errors))
	for i, obj := range r.Errors {
		messages[i] = obj.Message
	}

	return fmt.Errorf("%w: %v", base, strings.Join(messages, ", "))
}

// graphqlErrorInterceptor decorates the authenticated HTTP client so every
// operation (read, write, delete, metadata) passes through responseHandler.
// The components framework calls the client's Do directly, so a
// common.HTTPClient.ResponseHandler would not be applied there.
type graphqlErrorInterceptor struct {
	client common.AuthenticatedHTTPClient
}

func (i *graphqlErrorInterceptor) Do(req *http.Request) (*http.Response, error) {
	resp, err := i.client.Do(req)
	if err != nil {
		return nil, err
	}

	return responseHandler(resp)
}

func (i *graphqlErrorInterceptor) CloseIdleConnections() {
	i.client.CloseIdleConnections()
}

// graphqlResponse captures just enough of a GraphQL payload to decide
// whether a 200 response is actually a failure.
type graphqlResponse struct {
	Errors []struct {
		Extensions struct {
			Code string `json:"code"`
		} `json:"extensions"`
	} `json:"errors"`
	Data json.RawMessage `json:"data"`
}

// responseHandler rewrites the status code of erroneous 200 responses so the
// standard error interpreter takes over. GraphQL reports failures such as
// throttling or validation errors with a 200 status, an "errors" array and a
// missing or null "data" object. Partial successes (data alongside errors,
// e.g. mutations reporting userErrors) pass through unchanged and are
// handled by the parse functions.
func responseHandler(resp *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Reset body so downstream consumers can read it.
	resp.Body = io.NopCloser(bytes.NewReader(body))

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return resp, nil
	}

	var payload graphqlResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		// Not a JSON payload; leave it for downstream handling.
		return resp, nil // nolint:nilerr
	}

	hasData := len(payload.Data) > 0 && !bytes.Equal(payload.Data, []byte("null"))
	if len(payload.Errors) > 0 && !hasData {
		resp.StatusCode = graphqlStatusCode(payload.Errors[0].Extensions.Code)
	}

	return resp, nil
}

// graphqlStatusCode maps Jobber GraphQL error codes (extensions.code) to
// HTTP status codes understood by the error interpreter.
// https://developer.getjobber.com/docs/using_jobbers_api/api_rate_limits
func graphqlStatusCode(code string) int {
	switch code {
	case "UNAUTHENTICATED":
		return http.StatusUnauthorized
	case "THROTTLED":
		return http.StatusTooManyRequests
	default:
		return http.StatusBadRequest
	}
}

// This function uses to check whether the response(200 statuscode) contain error or not.
func checkErrorInResponse(errorArr []*ajson.Node) error {
	if len(errorArr) == 0 {
		return nil
	}

	var errorMsg strings.Builder

	for _, value := range errorArr {
		errMsg, err := jsonquery.New(value).StrWithDefault("message", "")
		if err != nil {
			return err
		}

		if errMsg != "" {
			errorMsg.WriteString(errMsg + "; ")
		}
	}

	return errors.New(strings.TrimSuffix(errorMsg.String(), "; ")) //nolint:err113
}
