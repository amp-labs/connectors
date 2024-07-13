package zendesksupport

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
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/go-test/deep"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	const workspace = "testWorkspace"

	responseErrorFormat := mockutils.DataFromFile(t, "resource-not-found.json")
	responseForbiddenError := mockutils.DataFromFile(t, "forbidden.json")
	responseUsersFirstPage := mockutils.DataFromFile(t, "read-users-1-first-page.json")
	responseUsersEmptyPage := mockutils.DataFromFile(t, "read-users-2-empty-page.json")
	responseReadTickets := mockutils.DataFromFile(t, "read-list-tickets.json")

	tests := []struct {
		name         string
		input        common.ReadParams
		server       *httptest.Server
		connector    Connector
		comparator   func(serverURL string, actual, expected *common.ReadResult) bool // custom comparison
		expected     *common.ReadResult
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
			name:  "Mime response header expected",
			input: common.ReadParams{ObjectName: "triggers"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			name:  "Correct error message is understood from JSON response",
			input: common.ReadParams{ObjectName: "triggers"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(responseErrorFormat)
			})),
			expectedErrs: []error{
				common.ErrBadRequest, errors.New("[InvalidEndpoint]Not found"), // nolint:goerr113
			},
		},
		{
			name:  "Forbidden error code and response",
			input: common.ReadParams{ObjectName: "triggers"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write(responseForbiddenError)
			})),
			expectedErrs: []error{
				common.ErrForbidden, errors.New("[Forbidden]You do not have access to this page"), // nolint:goerr113
			},
		},
		{
			name:  "Incorrect key in payload",
			input: common.ReadParams{ObjectName: "triggers"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `{
					"garbage": {}
				}`)
			})),
			expectedErrs: []error{jsonquery.ErrKeyNotFound},
		},
		{
			name:  "Incorrect data type in payload",
			input: common.ReadParams{ObjectName: "triggers"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `{
					"triggers": {}
				}`)
			})),
			expectedErrs: []error{jsonquery.ErrNotArray},
		},
		{
			name:  "Next page cursor may be missing in payload",
			input: common.ReadParams{ObjectName: "users"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseUsersEmptyPage)
			})),
			expected: &common.ReadResult{
				Data: []common.ReadResultRow{},
				Done: true,
			},
			expectedErrs: nil,
		},
		{
			name:  "Next page URL is resolved, when provided with a string",
			input: common.ReadParams{ObjectName: "users"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseUsersFirstPage)
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return actual.NextPage.String() == expected.NextPage.String()
			},
			expected: &common.ReadResult{
				NextPage: "https://d3v-ampersand.zendesk.com/api/v2/users" +
					"?page%5Bafter%5D=eyJvIjoiaWQiLCJ2IjoiYVJOc1cwVDZGd0FBIn0%3D&page%5Bsize%5D=3",
			},
			expectedErrs: nil,
		},
		{
			name: "Successful read with chosen fields",
			input: common.ReadParams{
				ObjectName: "tickets",
				Fields:     []string{"type", "subject", "status"},
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseReadTickets)
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				// custom comparison focuses on subset of fields to keep the test short
				return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
					mockutils.ReadResultComparator.SubsetRaw(actual, expected) &&
					actual.NextPage.String() == expected.NextPage.String() &&
					actual.Done == expected.Done
			},
			expected: &common.ReadResult{
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"type":    "incident",
						"subject": "Updated to: Hello World",
						"status":  "open",
					},
					Raw: map[string]any{
						"url":           "https://d3v-ampersand.zendesk.com/api/v2/tickets/1.json",
						"problem_id":    nil,
						"has_incidents": false,
						"brand_id":      float64(26363596759827),
					},
				}},
				NextPage: "https://d3v-ampersand.zendesk.com/api/v2/tickets.json" +
					"?page%5Bafter%5D=eyJvIjoibmljZV9pZCIsInYiOiJhUUVBQUFBQUFBQUEifQ%3D%3D&page%5Bsize%5D=1",
				Done: false,
			},
			expectedErrs: nil,
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
				WithWorkspace(workspace),
			)
			if err != nil {
				t.Fatalf("%s: error in test while constructing connector %v", tt.name, err)
			}

			// for testing we want to redirect calls to our mock server
			connector.setBaseURL(tt.server.URL)

			// start of tests
			output, err := connector.Read(ctx, tt.input)
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
			for _, expectedErr := range tt.expectedErrs {
				if !errors.Is(err, expectedErr) && !strings.Contains(err.Error(), expectedErr.Error()) {
					t.Fatalf("%s: expected Error: (%v), got: (%v)", tt.name, expectedErr, err)
				}
			}

			// compare desired output
			var ok bool
			if tt.comparator == nil {
				// default comparison is concerned about all fields
				ok = reflect.DeepEqual(output, tt.expected)
			} else {
				ok = tt.comparator(tt.server.URL, output, tt.expected)
			}

			if !ok {
				diff := deep.Equal(output, tt.expected)
				t.Fatalf("%s:, \nexpected: (%v), \ngot: (%v), \ndiff: (%v)", tt.name, tt.expected, output, diff)
			}
		})
	}
}
