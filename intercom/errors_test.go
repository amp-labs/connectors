package intercom

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestInterpretJSONError(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseNotAcceptable := testutils.DataFromFile(t, "media-not-acceptable.json")

	type input struct {
		res  *http.Response
		body []byte
	}

	tests := []struct {
		name         string
		input        input
		comparator   func(actual error, expectedErrs []error) bool
		expectedErrs []error
	}{
		{
			name: "Missing response body is reported as empty",
			input: input{
				res:  &http.Response{},
				body: nil,
			},
			expectedErrs: []error{interpreter.ErrEmptyResponse},
		},
		{
			name: "Empty response body is reported as empty",
			input: input{
				res:  &http.Response{},
				body: []byte(``),
			},
			expectedErrs: []error{interpreter.ErrEmptyResponse},
		},
		{
			name: "Unknown response status produces caller error",
			input: input{
				res: &http.Response{
					StatusCode: http.StatusTeapot,
				},
				body: []byte(`{}`),
			},
			expectedErrs: []error{common.ErrCaller},
		},
		{
			name: "Correct status of TooManyRequests",
			input: input{
				res: &http.Response{
					StatusCode: http.StatusTooManyRequests,
				},
				body: []byte(`{"data":{"base":"BTC","currency":"USD","amount":4225.87}}`),
			},
			expectedErrs: []error{common.ErrLimitExceeded},
		},
		{
			name: "Conflict response with missing error codes",
			input: input{
				res: &http.Response{
					StatusCode: http.StatusConflict,
				},
				// response body should have `code`. In addition, 409 is mapped correctly
				body: []byte(`{"errors": [{"details":"conflicting values"}]}`),
			},
			expectedErrs: []error{common.ErrBadRequest, ErrUnknownErrorResponseFormat}, // nolint:goerr113
		},
		{
			name: "Correct status and message handling",
			input: input{
				res: &http.Response{
					StatusCode: http.StatusNotAcceptable,
				},
				body: responseNotAcceptable,
			},
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New( // nolint:goerr113
					"media_type_not_acceptable[The Accept header should send a media type of application/json]",
				),
			},
		},
	}

	connector := Connector{}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := connector.interpretJSONError(tt.input.res, tt.input.body)
			if err != nil {
				if len(tt.expectedErrs) == 0 {
					t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
				}
			} else {
				// check that missing error is what is expected
				if len(tt.expectedErrs) != 0 {
					t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
				}
			}

			// check every error
			if tt.comparator == nil {
				for _, expectedErr := range tt.expectedErrs {
					if !errors.Is(err, expectedErr) && !strings.Contains(err.Error(), expectedErr.Error()) {
						t.Fatalf("%s: expected Error: (%v), got: (%v)", tt.name, expectedErr, err)
					}
				}
			} else { // nolint:gocritic
				if !tt.comparator(err, tt.expectedErrs) {
					t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expectedErrs, err)
				}
			}
		})
	}
}
