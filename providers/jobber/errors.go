package jobber

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

// graphqlErrorResponder formats GraphQL error payloads into typed errors,
// reusing the same error schema as the standard (non-2xx) error handler.
//
//nolint:gochecknoglobals
var graphqlErrorResponder = interpreter.NewFaultyResponder(errorFormats, nil)

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

// interpretGraphQLError returns a typed error when a 2xx GraphQL response is
// actually a failure. Jobber reports errors such as throttling or invalid
// queries with a 200 status, an "errors" array and a missing or null "data"
// object, so the operation's status-based error handler never fires. The read
// path calls this explicitly. Partial successes (data alongside errors, e.g.
// mutations reporting userErrors) return nil and are handled by the parse
// functions themselves.
//
// The error code is mapped to an HTTP status so the shared FaultyResponder can
// produce the same typed errors (ErrLimitExceeded, ErrAccessToken, ...) it
// would for a genuine non-2xx response.
func interpretGraphQLError(resp *common.JSONHTTPResponse) error {
	node, ok := resp.Body()
	if !ok {
		return nil
	}

	body := node.Source()

	var payload graphqlResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil //nolint:nilerr
	}

	hasData := len(payload.Data) > 0 && !bytes.Equal(payload.Data, []byte("null"))
	if len(payload.Errors) == 0 || hasData {
		return nil
	}

	synthetic := &http.Response{
		StatusCode: graphqlStatusCode(payload.Errors[0].Extensions.Code),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}

	return graphqlErrorResponder.HandleErrorResponse(synthetic, body)
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
