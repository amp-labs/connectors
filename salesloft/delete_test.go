package salesloft

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
	"github.com/go-test/deep"
)

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	listSchema := mockutils.DataFromFile(t, "write-signals-error.json")

	tests := []struct {
		name         string
		input        common.DeleteParams
		server       *httptest.Server
		connector    Connector
		expected     *common.DeleteResult
		expectedErrs []error
	}{
		{
			name: "Write object must be included",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{common.ErrMissingObjects},
		},
		{
			name:  "Write object and its ID must be included",
			input: common.DeleteParams{ObjectName: "signals"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			name:  "Mime response header expected",
			input: common.DeleteParams{ObjectName: "signals", RecordId: "22165"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			name:  "Correct error message is understood from JSON response",
			input: common.DeleteParams{ObjectName: "signals", RecordId: "22165"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write(listSchema)
			})),
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New("no Signal Registration found for integration id 5167 and given type"), // nolint:goerr113
			},
		},
		{
			name:  "Successful delete",
			input: common.DeleteParams{ObjectName: "signals", RecordId: "22165"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondNoContentForMethod(w, r, "DELETE")
			})),
			expected:     &common.DeleteResult{Success: true},
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
			)
			if err != nil {
				t.Fatalf("%s: error in test while constructing connector %v", tt.name, err)
			}

			// for testing we want to redirect calls to our server
			connector.setBaseURL(tt.server.URL)

			// start of tests
			output, err := connector.Delete(ctx, tt.input)
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
