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
		{
			name: "Server error response invalid 'status' data type",
			input: input{
				res: &http.Response{
					StatusCode: http.StatusBadRequest,
				},
				body: []byte(`{"status":"string while it should be a number"}`),
			},
			expectedErr: interpreter.ErrUnmarshal,
		},
	}

	connector := Connector{}

	for _, tt := range tests {
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := connector.interpretJSONError(tt.input.res, tt.input.body)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expectedErr, err)
			}
		})
	}

	t.Run("Correct interpretation of error payload", func(t *testing.T) {
		t.Parallel()

		err := connector.interpretJSONError(&http.Response{
			StatusCode: http.StatusBadRequest,
		}, []byte(`{  
				"status": 123,
				"error": "error message from server"}`))
		if !strings.Contains(err.Error(), "error message from server") {
			t.Fatalf("expected errot type mismatched for: (%v)", err)
		}
	})
}
