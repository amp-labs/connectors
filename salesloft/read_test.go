package salesloft

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/go-test/deep"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseEmptyRead := testutils.DataFromFile(t, "read-empty.json")
	responseListPeople := testutils.DataFromFile(t, "read-list-people.json")
	responseListUsers := testutils.DataFromFile(t, "read-list-users.json")
	responseListAccounts := testutils.DataFromFile(t, "read-list-accounts.json")
	responseListAccountsSince := testutils.DataFromFile(t, "read-list-accounts-since.json")
	accountsSince, err := time.Parse(time.RFC3339Nano, "2024-06-07T10:51:20.851224-04:00")
	mockutils.NoErrors(t, err)

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
			name:         "Mime response header expected",
			server:       mockserver.Dummy(),
			expectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			name: "Correct error message is understood from JSON response",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				mockutils.WriteBody(w, `{
					"error": "Not Found"
				}`)
			})),
			expectedErrs: []error{
				common.ErrBadRequest, errors.New("Not Found"), // nolint:goerr113
			},
		},
		{
			name: "Incorrect key in payload",
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
			name: "Incorrect data type in payload",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `{
					"data": {}
				}`)
			})),
			expectedErrs: []error{jsonquery.ErrNotArray},
		},
		{
			name: "Next page cursor may be missing in payload",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseEmptyRead)
			})),
			expected: &common.ReadResult{
				Data: []common.ReadResultRow{},
				Done: true,
			},
			expectedErrs: nil,
		},
		{
			name: "Next page URL is correctly inferred",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseListPeople)
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				expectedNextPage := strings.ReplaceAll(expected.NextPage.String(), "{{testServerURL}}", baseURL)
				return actual.NextPage.String() == expectedNextPage // nolint:nlreturn
			},
			expected: &common.ReadResult{
				NextPage: "{{testServerURL}}/v2?page=2&per_page=100",
			},
			expectedErrs: nil,
		},
		{
			name: "Successful read with 25 entries, checking one row",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseListPeople)
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return mockutils.ReadResultComparator.SubsetRaw(actual, expected) &&
					actual.Done == expected.Done &&
					actual.Rows == expected.Rows
			},
			expected: &common.ReadResult{
				Rows: 25,
				// We are only interested to validate only first Read Row!
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"first_name":             "Lynnelle",
						"email_address":          "losbourn29@paypal.com",
						"full_email_address":     "\"Lynnelle new\" <losbourn29@paypal.com>",
						"person_company_website": "http://paypal.com",
					},
				}},
				Done: false,
			},
			expectedErrs: nil,
		},
		{
			name: "Successful read with chosen fields",
			input: common.ReadParams{
				Fields: []string{"email_address", "person_company_website"},
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseListPeople)
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
					mockutils.ReadResultComparator.SubsetRaw(actual, expected)
			},
			expected: &common.ReadResult{
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"email_address":          "losbourn29@paypal.com",
						"person_company_website": "http://paypal.com",
					},
					Raw: map[string]any{
						"first_name":             "Lynnelle",
						"email_address":          "losbourn29@paypal.com",
						"full_email_address":     "\"Lynnelle new\" <losbourn29@paypal.com>",
						"person_company_website": "http://paypal.com",
					},
				}},
			},
			expectedErrs: nil,
		},
		{
			name: "Listing Users without pagination payload",
			input: common.ReadParams{
				Fields: []string{"email", "guid"},
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseListUsers)
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
					mockutils.ReadResultComparator.SubsetRaw(actual, expected)
			},
			expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"guid":  "0863ed13-7120-479b-8650-206a3679e2fb",
						"email": "somebody@withampersand.com",
					},
					Raw: map[string]any{
						"name":       "Int User",
						"first_name": "Int",
						"last_name":  "User",
					},
				}},
				NextPage: "",
				Done:     false,
			},
			expectedErrs: nil,
		},
		{
			name:  "Successful read accounts without since query",
			input: common.ReadParams{ObjectName: "accounts"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMissingQueryParameters(w, r, []string{"updated_at[gte]"}, func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseListAccounts)
				})
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return actual.Rows == expected.Rows
			},
			expected: &common.ReadResult{
				Rows: 4,
			},
			expectedErrs: nil,
		},
		{
			name: "Successful read accounts since point in time",
			input: common.ReadParams{
				ObjectName: "accounts",
				Since:      accountsSince,
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToQueryParameters(w, r, url.Values{
					"updated_at[gte]": []string{"2024-06-07T10:51:20.851224-04:00"},
				}, func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseListAccountsSince)
				})
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return actual.Rows == expected.Rows
			},
			expected: &common.ReadResult{
				Rows: 2,
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
