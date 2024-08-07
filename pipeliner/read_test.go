package pipeliner

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
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/go-test/deep"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseNotFound := testutils.DataFromFile(t, "resource-not-found.json")
	responseProfilesFirstPage := testutils.DataFromFile(t, "read-profiles-1-first-page.json")
	responseProfilesSecondPage := testutils.DataFromFile(t, "read-profiles-2-second-page.json")
	responseProfilesLastPage := testutils.DataFromFile(t, "read-profiles-3-last-page.json")

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
			name:  "Correct error message is understood from JSON response",
			input: common.ReadParams{ObjectName: "articles"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(responseNotFound)
			})),
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New("The requested URL was not found on the server"), // nolint:goerr113
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
				mockutils.WriteBody(w, `
				{
				  "success": true,
				  "total": 0,
				  "data": []
				}`)
			})),
			expected: &common.ReadResult{
				Data: []common.ReadResultRow{},
				Done: true,
			},
			expectedErrs: nil,
		},
		{
			name: "Next page URL is inferred, when provided with an object",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseProfilesFirstPage)
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return actual.NextPage.String() == expected.NextPage.String() // nolint:nlreturn
			},
			expected: &common.ReadResult{
				NextPage: "WyIwMDAwMDAwMC0wMDAwLTAwMDEtMDAwMS0wMDAwMDAwMDhlOTciXQ==",
			},
			expectedErrs: nil,
		},
		{
			name: "Next page URL is empty, when provided with null object",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseProfilesLastPage)
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return actual.NextPage.String() == expected.NextPage.String() &&
					actual.Done == expected.Done
			},
			expected:     &common.ReadResult{NextPage: "", Done: true},
			expectedErrs: nil,
		},
		{
			name: "Successful read with chosen fields",
			input: common.ReadParams{
				Fields: []string{"name", "owner_id"},
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseProfilesSecondPage)
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
						"name":     "Lang_DefaultProfileAllUsers",
						"owner_id": "00000000-0000-0000-0000-000000008e97",
					},
					Raw: map[string]any{
						"name":       "Lang_DefaultProfileAllUsers",
						"owner_id":   "00000000-0000-0000-0000-000000008e97",
						"use_lang":   true,
						"entity":     float64(3),
						"is_deleted": false,
					},
				}, {
					Fields: map[string]any{
						"name":     "Lang_DefaultProfileMy",
						"owner_id": "00000000-0000-0000-0000-000000008e97",
					},
					Raw: map[string]any{
						"name":       "Lang_DefaultProfileMy",
						"owner_id":   "00000000-0000-0000-0000-000000008e97",
						"use_lang":   true,
						"entity":     float64(3),
						"is_deleted": false,
					},
				}},
				NextPage: "WyIwMDAwMDAwMC0wMDAwLTAwMDMtMDAwMS0wMDAwMDAwMDhlOTciXQ==",
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
