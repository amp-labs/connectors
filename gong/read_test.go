package gong

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"

	"github.com/amp-labs/connectors/utils"
	"github.com/go-test/deep"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []struct {
		name             string
		input            common.ReadParams
		server           *httptest.Server
		connector        Connector
		expected         *common.ReadResult
		expectedErrs     []error
		expectedErrTypes []error
	}{
		{
			name: "Mime response header expected",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			name: "Correct error message is understood from JSON response",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				writeBody(w, `{
					"error": {
						"code": "0x80060888",
						"message":"Resource not found"
					}
				}`)
			})),
			expectedErrs: []error{
				errors.New("unsupported operation"), // nolint:goerr113
			},
		},

		{
			name: "Incorrect data type in payload",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				writeBody(w, `{
					"values": {}
				}`)
			})),
			expectedErrs: []error{common.ErrNotArray},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.server.Close()

			ctx := context.Background()

			connector, err := NewConnector(
				WithAuthenticatedClient(http.DefaultClient),
				WithWorkspace(utils.GongWorkspace),
				WithModule(DefaultModule),
			)
			if err != nil {
				t.Fatalf("%s: error in test while constructin connector %v", tt.name, err)
			}

			// for testing we want to redirect calls to our mock server
			connector.setBaseURL(tt.server.URL)

			// start of tests
			output, err := connector.Read(ctx, tt.input)
			if err != nil {
				if len(tt.expectedErrs)+len(tt.expectedErrTypes) == 0 {
					t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
				}
			} else {
				// check that missing error is what is expected
				if len(tt.expectedErrs) != 0 {
					t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
				}

				if len(tt.expectedErrTypes) != 0 {
					t.Fatalf("%s: expected error types (%v), but got nothing", tt.name, tt.expectedErrTypes)
				}
			}

			// check every error
			for _, expectedErr := range tt.expectedErrTypes {
				if reflect.TypeOf(err) != reflect.TypeOf(expectedErr) {
					t.Fatalf("%s: expected Error type: (%T), got: (%T)", tt.name, expectedErr, err)
				}
			}

			for _, expectedErr := range tt.expectedErrs {
				if !errors.Is(err, expectedErr) && !strings.Contains(err.Error(), expectedErr.Error()) {
					t.Fatalf("%s: expected Error: (%v), got: (%v)", tt.name, expectedErr, err)
				}
			}

			// compare desired output
			if !reflect.DeepEqual(output, tt.expected) {
				diff := deep.Equal(output, tt.expected)
				t.Fatalf("%s:, \nexpected: (%v), \ngot: (%v), \ndiff: (%v)", tt.name, tt.expected, output, diff)
			}
		})
	}
}

func writeBody(w http.ResponseWriter, body string) {
	_, _ = w.Write([]byte(body))
}
