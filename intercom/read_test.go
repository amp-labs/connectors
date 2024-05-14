package intercom

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

	responseErrorFormat := mockutils.DataFromFile(t, "page-req-too-large.json")
	responseNotesFirstPage := mockutils.DataFromFile(t, "read-notes-1-first-page.json")
	responseContactsFirstPage := mockutils.DataFromFile(t, "read-contacts-1-first-page.json")

	// TODO deal with other sample files

	// x := mockutils.DataFromFile(t, "read-contacts-3-last-page.json")
	// x := mockutils.DataFromFile(t, "read-notes-2-last-page.json")
	// x := mockutils.DataFromFile(t, "read-conversations.json")

	// TODO think about test cases

	// list => data [popular]
	// something.list => apply plural form to `something` [common]
	// event.summary => events (add as exception?) [unique]
	// Download content data export: returns None in payload [unique]

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
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(responseErrorFormat)
			})),
			expectedErrs: []error{
				common.ErrBadRequest, errors.New("parameter_invalid[Per Page is too big]"), // nolint:goerr113
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
				  "type": "list",
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
			name: "Next page URL is resolved, when provided with a string",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseNotesFirstPage)
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return actual.NextPage.String() == expected.NextPage.String()
			},
			expected: &common.ReadResult{
				NextPage: "https://api.intercom.io/contacts/6643703ffae7834d1792fd30/notes?per_page=2&page=2",
			},
			expectedErrs: nil,
		},
		{
			name: "Next page URL is inferred, when provided with an object",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseContactsFirstPage)
			})),
			comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				expectedNextPage := strings.ReplaceAll(expected.NextPage.String(), "{{testServerURL}}", baseURL)
				return actual.NextPage.String() == expectedNextPage // nolint:nlreturn
			},
			expected: &common.ReadResult{
				NextPage: "{{testServerURL}}?per_page=50&starting_after=" +
					"WzE3MTU2OTU2NzkwMDAsIjY2NDM3MDNmZmFlNzgzNGQxNzkyZmQzMCIsMl0=",
			},
			expectedErrs: nil,
		},
		//{
		//	name: "Successful read with 25 entries, checking one row",
		//	server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//		w.Header().Set("Content-Type", "application/json")
		//		w.WriteHeader(http.StatusOK)
		//		_, _ = w.Write(responseListPeople)
		//	})),
		//	comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
		//		return mockutils.ReadResultComparator.SubsetRaw(actual, expected) &&
		//			actual.Done == expected.Done &&
		//			actual.Rows == expected.Rows
		//	},
		//	expected: &common.ReadResult{
		//		Rows: 25,
		//		// We are only interested to validate only first Read Row!
		//		Data: []common.ReadResultRow{{
		//			Fields: map[string]any{},
		//			Raw: map[string]any{
		//				"first_name":             "Lynnelle",
		//				"email_address":          "losbourn29@paypal.com",
		//				"full_email_address":     "\"Lynnelle new\" <losbourn29@paypal.com>",
		//				"person_company_website": "http://paypal.com",
		//			},
		//		}},
		//		Done: false,
		//	},
		//	expectedErrs: nil,
		// },
		//{
		//	name: "Successful read with chosen fields",
		//	input: common.ReadParams{
		//		Fields: []string{"email_address", "person_company_website"},
		//	},
		//	server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//		w.Header().Set("Content-Type", "application/json")
		//		w.WriteHeader(http.StatusOK)
		//		_, _ = w.Write(responseListPeople)
		//	})),
		//	comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
		//		return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
		//			mockutils.ReadResultComparator.SubsetRaw(actual, expected)
		//	},
		//	expected: &common.ReadResult{
		//		Data: []common.ReadResultRow{{
		//			Fields: map[string]any{
		//				"email_address":          "losbourn29@paypal.com",
		//				"person_company_website": "http://paypal.com",
		//			},
		//			Raw: map[string]any{
		//				"first_name":             "Lynnelle",
		//				"email_address":          "losbourn29@paypal.com",
		//				"full_email_address":     "\"Lynnelle new\" <losbourn29@paypal.com>",
		//				"person_company_website": "http://paypal.com",
		//			},
		//		}},
		//	},
		//	expectedErrs: nil,
		//},
		//{
		//	name: "Listing Users without pagination payload",
		//	input: common.ReadParams{
		//		Fields: []string{"email", "guid"},
		//	},
		//	server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//		w.Header().Set("Content-Type", "application/json")
		//		w.WriteHeader(http.StatusOK)
		//		_, _ = w.Write(responseListUsers)
		//	})),
		//	comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
		//		return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
		//			mockutils.ReadResultComparator.SubsetRaw(actual, expected)
		//	},
		//	expected: &common.ReadResult{
		//		Rows: 1,
		//		Data: []common.ReadResultRow{{
		//			Fields: map[string]any{
		//				"guid":  "0863ed13-7120-479b-8650-206a3679e2fb",
		//				"email": "somebody@withampersand.com",
		//			},
		//			Raw: map[string]any{
		//				"name":       "Int User",
		//				"first_name": "Int",
		//				"last_name":  "User",
		//			},
		//		}},
		//		NextPage: "",
		//		Done:     false,
		//	},
		//	expectedErrs: nil,
		//},
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
