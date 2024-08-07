package dynamicscrm

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
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/go-test/deep"
)

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	tests := []struct {
		name         string
		input        common.WriteParams
		server       *httptest.Server
		connector    Connector
		expected     *common.WriteResult
		expectedErrs []error
	}{
		{
			name:         "Write object must be included",
			server:       mockserver.Dummy(),
			expectedErrs: []error{common.ErrMissingObjects},
		},
		{
			name:         "Mime response header expected",
			input:        common.WriteParams{ObjectName: "fax"},
			server:       mockserver.Dummy(),
			expectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			name:  "Correct error message is understood from JSON response",
			input: common.WriteParams{ObjectName: "fax"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				mockutils.WriteBody(w, `{
					"error": {
						"code": "0x80060888",
						"message":"Resource not found for the segment 'conacs'."
					}
				}`)
			})),
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Resource not found for the segment 'conacs'"), // nolint:goerr113
			},
		},
		{
			name:  "Write must act as a Create",
			input: common.WriteParams{ObjectName: "fax"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondNoContentForMethod(w, r, "POST")
			})),
			expected:     &common.WriteResult{Success: true},
			expectedErrs: nil,
		},
		{
			name:  "Write must act as an Update",
			input: common.WriteParams{ObjectName: "fax", RecordId: "dd2f7870-3fe8-ee11-a204-0022481f9e3c"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondNoContentForMethod(w, r, "PATCH")
			})),
			expected:     &common.WriteResult{Success: true},
			expectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.server.Close()

			ctx := context.Background()

			connector, err := NewConnector(
				WithAuthenticatedClient(http.DefaultClient),
				WithWorkspace("test-workspace"),
			)
			if err != nil {
				t.Fatalf("%s: error in test while constructing connector %v", tt.name, err)
			}

			// for testing we want to redirect calls to our server
			connector.setBaseURL(tt.server.URL)

			// start of tests
			output, err := connector.Write(ctx, tt.input)
			if len(tt.expectedErrs) == 0 && err != nil {
				t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
			}

			if len(tt.expectedErrs) != 0 && err == nil {
				t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
			}

			for _, expectedErr := range tt.expectedErrs {
				if !errors.Is(err, expectedErr) && !strings.Contains(err.Error(), expectedErr.Error()) {
					t.Fatalf("%s: expected Error: (%v), got: (%v)", tt.name, expectedErr, err)
				}
			}

			if !reflect.DeepEqual(output, tt.expected) {
				diff := deep.Equal(output, tt.expected)
				t.Fatalf("%s:, \nexpected: (%v), \ngot: (%v), \ndiff: (%v)", tt.name, tt.expected, output, diff)
			}
		})
	}
}
