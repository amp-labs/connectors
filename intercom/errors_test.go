package intercom

import (
	"errors"
	"net/http"
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
				res:  nil,
				body: nil,
			},
			expectedErr: interpreter.ErrUnmarshal,
		},
		{
			name: "Empty response body cannot be unmarshalled",
			input: input{
				res:  nil,
				body: []byte(``),
			},
			expectedErr: interpreter.ErrUnmarshal,
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
