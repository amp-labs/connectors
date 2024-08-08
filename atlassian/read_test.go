package atlassian

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

	responseErrorFormat := testutils.DataFromFile(t, "jql-error.json")
	responseIssuesFirstPage := testutils.DataFromFile(t, "read-issues.json")
	responsePathNotFoundError := testutils.DataFromFile(t, "path-not-found.json")

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
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseErrorFormat)
			})),
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Date value '-53s' for field 'updated' is invalid"), // nolint:goerr113
			},
		},
		{
			name: "Invalid path understood as not found error",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(responsePathNotFoundError)
			})),
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Not Found - No message available"), // nolint:goerr113
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
					"issues": {}
				}`)
			})),
			expectedErrs: []error{jsonquery.ErrNotArray},
		},
		{
			name: "Empty array produces no next page",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `
				{
				  "startAt": 6,
				  "issues": []
				}`)
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return nextPageComparator(actual, expected)
			},
			expected: &common.ReadResult{
				Rows:     0,
				NextPage: "",
				Done:     true,
			},
			expectedErrs: nil,
		},
		{
			name: "Issue must have fields property",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `
				{
				  "issues": [{}]
				}`)
			})),
			expectedErrs: []error{jsonquery.ErrKeyNotFound},
		},
		{
			name: "Issue must have id property",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `
				{
				  "issues": [{"fields":{}}]
				}`)
			})),
			expectedErrs: []error{jsonquery.ErrKeyNotFound},
		},
		{
			name: "Missing starting index produces no next page",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `
				{
				  "issues": [
					{"fields":{}, "id": "0"},
					{"fields":{}, "id": "1"}
				]}`)
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return nextPageComparator(actual, expected)
			},
			expected: &common.ReadResult{
				Rows:     2,
				NextPage: "",
				Done:     true,
			},
			expectedErrs: nil,
		},
		{
			name: "Next page is implied from start index and issues size",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `
				{
				  "startAt": 6,
				  "issues": [
					{"fields":{}, "id": "0"},
					{"fields":{}, "id": "1"}
				]}`)
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return nextPageComparator(actual, expected)
			},
			expected: &common.ReadResult{
				Rows:     2,
				NextPage: "8",
				Done:     false,
			},
			expectedErrs: nil,
		},
		{
			name: "Since rounds to minute time frame",
			input: common.ReadParams{
				Since: time.Now().Add(-5 * time.Minute),
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.RespondToQueryParameters(w, r, url.Values{
					// server was asked to get issues that occurred in the last 5 min
					"jql": []string{`updated > "-5m"`},
				}, func() {
					mockutils.WriteBody(w, `
					{
					  "startAt": 0,
					  "issues": [{"fields":{}, "id": "0"}]
					}`)
				})
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return actual.Rows == expected.Rows
			},
			expected: &common.ReadResult{
				Rows: 1,
			},
			expectedErrs: nil, // there must be no errors.
		},
		{
			name: "Next page is propagated in query params",
			input: common.ReadParams{
				NextPage: "17",
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.RespondToQueryParameters(w, r, url.Values{
					"startAt": []string{"17"},
				}, func() {
					mockutils.WriteBody(w, `
					{
					  "startAt": 17,
					  "issues": [
						{"fields":{}, "id": "0"},
						{"fields":{}, "id": "1"},
						{"fields":{}, "id": "2"}
					]}`)
				})
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return actual.Rows == expected.Rows
			},
			expected: &common.ReadResult{
				Rows: 3,
			},
			expectedErrs: nil, // there must be no errors
		},
		{
			name: "Successful list of rows",
			input: common.ReadParams{
				Fields: []string{"id", "summary"},
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseIssuesFirstPage)
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
					mockutils.ReadResultComparator.SubsetRaw(actual, expected) &&
					actual.NextPage.String() == expected.NextPage.String() &&
					actual.Rows == expected.Rows &&
					actual.Done == expected.Done
			},
			expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":      "10001",
						"summary": "Another one",
					},
					Raw: map[string]any{
						"id":                       "10001",
						"description":              nil,
						"created":                  "2024-07-22T22:41:48.474+0300",
						"statuscategorychangedate": "2024-07-22T22:41:48.686+0300",
					},
				}, {
					Fields: map[string]any{
						"id":      "10000",
						"summary": "First Issue on Jira",
					},
					Raw: map[string]any{
						"id":                       "10000",
						"description":              nil,
						"created":                  "2024-07-22T22:41:35.069+0300",
						"statuscategorychangedate": "2024-07-22T22:41:35.326+0300",
					},
				}},
				NextPage: "2",
				Done:     false,
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
				WithWorkspace("test-workspace"),
				WithModule(ModuleJira),
				WithMetadata(map[string]string{
					"cloudId": "ebc887b2-7e61-4059-ab35-71f15cc16e12", // any value will work for the test
				}),
			)
			if err != nil {
				t.Fatalf("%s: error in test while constructing connector %v", tt.name, err)
			}

			// for testing we want to redirect calls to our mock server
			connector.setBaseURL(tt.server.URL)

			if err != nil {
				t.Fatalf("%s: failed to setup auth metadata connector %v", tt.name, err)
			}

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

func TestReadWithoutMetadata(t *testing.T) {
	t.Parallel()

	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
		WithWorkspace("test-workspace"),
		WithModule(ModuleJira),
	)
	if err != nil {
		t.Fatal("failed to create connector")
	}

	_, err = connector.Read(context.Background(), common.ReadParams{})
	if !errors.Is(err, ErrMissingCloudId) {
		t.Fatalf("expected Read method to complain about missing cloud id")
	}
}

func nextPageComparator(actual *common.ReadResult, expected *common.ReadResult) bool {
	return actual.NextPage.String() == expected.NextPage.String() &&
		actual.Rows == expected.Rows &&
		actual.Done == expected.Done
}
