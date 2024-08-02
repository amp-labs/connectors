package salesloft

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

func TestInterpretJSONError(t *testing.T) { //nolint:funlen
	t.Parallel()

	type input struct {
		res  *http.Response
		body []byte
	}

	tests := []struct {
		name        string
		input       input
		comparator  func(actual, expected error) bool
		expectedErr error
	}{
		{
			name: "Missing response body cannot be unmarshalled",
			input: input{
				res:  &http.Response{},
				body: nil,
			},
			expectedErr: interpreter.ErrEmptyResponse,
		},
		{
			name: "Empty response body cannot be unmarshalled",
			input: input{
				res:  &http.Response{},
				body: []byte(``),
			},
			expectedErr: interpreter.ErrEmptyResponse,
		},
		{
			name: "Unknown response status produces caller error",
			input: input{
				res: &http.Response{
					StatusCode: http.StatusTeapot,
				},
				body: []byte(`{}`),
			},
			expectedErr: common.ErrCaller,
		},
		{
			name: "Correct status of TooManyRequests",
			input: input{
				res: &http.Response{
					StatusCode: http.StatusTooManyRequests,
				},
				body: []byte(`{"data":{"base":"BTC","currency":"USD","amount":4225.87}}`),
			},
			expectedErr: common.ErrLimitExceeded,
		},
		{
			name: "Server unknown error response, because mismatching 'status' data type",
			input: input{
				res: &http.Response{
					StatusCode: http.StatusBadRequest,
				},
				body: []byte(`{"status":"string while it should be a number"}`),
			},
			expectedErr: interpreter.ErrUnknownResponseFormat,
		},
		{
			name: "Correct interpretation of singular error payload",
			input: input{
				res: &http.Response{
					StatusCode: http.StatusBadRequest,
				},
				body: []byte(`{"status": 123,"error": "error message from server"}`),
			},
			comparator:  softStringErrComparison,
			expectedErr: errors.New("error message from server"), // nolint:goerr113
		},
		{
			name: "List schema is selected for error response",
			input: input{
				res: &http.Response{
					StatusCode: http.StatusBadRequest,
				},
				body: []byte(`{"errors": {"details":"some helpful message"}}`),
			},
			comparator: softStringErrComparison,
			// this error message is coming from response payload
			expectedErr: errors.New("some helpful message"), // nolint:goerr113
		},
	}

	connector := Connector{}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := connector.interpretJSONError(tt.input.res, tt.input.body)

			var ok bool
			if tt.comparator == nil {
				ok = errors.Is(err, tt.expectedErr)
			} else {
				ok = tt.comparator(err, tt.expectedErr)
			}

			if !ok {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expectedErr, err)
			}
		})
	}
}

func softStringErrComparison(actual, expected error) bool {
	return strings.Contains(actual.Error(), expected.Error())
}
