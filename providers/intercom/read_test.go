package intercom

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

var testApiVersionHeader = http.Header{} // nolint:gochecknoglobals

func init() {
	testApiVersionHeader.Add("Intercom-Version", "2.11")
}

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseErrorFormat := testutils.DataFromFile(t, "page-req-too-large.json")
	responseContactsFirstPage := testutils.DataFromFile(t, "read-contacts-1-first-page.json")
	responseContactsSecondPage := testutils.DataFromFile(t, "read-contacts-2-second-page.json")
	responseContactsThirdPage := testutils.DataFromFile(t, "read-contacts-3-last-page.json")
	responseReadConversations := testutils.DataFromFile(t, "read-conversations.json")
	responseNotesFirstPage := testutils.DataFromFile(t, "read-notes-1-first-page.json")
	responseNotesSecondPage := testutils.DataFromFile(t, "read-notes-2-last-page.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unsupported object name",
			Input:        common.ReadParams{ObjectName: "butterflies", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Mime response header expected",
			Input:        common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseErrorFormat),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest, errors.New("parameter_invalid[Per Page is too big]"), // nolint:goerr113
			},
		},
		{
			Name:  "Incorrect key in payload",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{
					"garbage": {}
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrKeyNotFound},
		},
		{
			Name:  "Incorrect data type in payload",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{
					"data": {}
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrNotArray},
		},
		{
			Name:  "Next page cursor may be missing in payload",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `
				{
				  "type": "list",
				  "data": []
				}`),
			}.Server(),
			Expected: &common.ReadResult{
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "API version header is passed as server request",
			Input: common.ReadParams{ObjectName: "articles", Fields: connectors.Fields("id")},
			// notes is not supported for now, but its payload is good for testing
			Server: mockserver.Reactive{
				Setup:     mockserver.ContentJSON(),
				Condition: mockcond.Header(testApiVersionHeader),
				OnSuccess: mockserver.Response(http.StatusOK, responseNotesSecondPage),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				// response doesn't matter much, as soon as we don't have errors we are good
				return actual.Done == expected.Done
			},
			Expected:     &common.ReadResult{Done: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Next page URL is resolved, when provided with a string",
			Input: common.ReadParams{ObjectName: "articles", Fields: connectors.Fields("id")},
			// notes is not supported for now, but its payload is good for testing
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseNotesFirstPage),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return actual.NextPage.String() == expected.NextPage.String()
			},
			Expected: &common.ReadResult{
				NextPage: "https://api.intercom.io/contacts/6643703ffae7834d1792fd30/notes?per_page=2&page=2",
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Next page URL is inferred, when provided with an object",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseContactsFirstPage),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				expectedNextPage := strings.ReplaceAll(expected.NextPage.String(), "{{testServerURL}}", baseURL)
				return actual.NextPage.String() == expectedNextPage // nolint:nlreturn
			},
			Expected: &common.ReadResult{
				NextPage: "{{testServerURL}}/contacts?per_page=60&starting_after=" +
					"WzE3MTU2OTU2NzkwMDAsIjY2NDM3MDNmZmFlNzgzNGQxNzkyZmQzMCIsMl0=",
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Next page URL is empty, when provided with null object",
			Input: common.ReadParams{ObjectName: "articles", Fields: connectors.Fields("id")},
			// notes is not supported for now, but its payload is good for testing
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseNotesSecondPage),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return actual.NextPage.String() == expected.NextPage.String() &&
					actual.Done == expected.Done
			},
			Expected:     &common.ReadResult{NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Next page URL is empty, when provided with missing object",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseContactsThirdPage),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return actual.NextPage.String() == expected.NextPage.String() &&
					actual.Done == expected.Done
			},
			Expected:     &common.ReadResult{NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read with chosen fields",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email", "name"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseContactsSecondPage),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				expectedNextPage := strings.ReplaceAll(expected.NextPage.String(), "{{testServerURL}}", baseURL)
				// custom comparison focuses on subset of fields to keep the test short
				return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
					mockutils.ReadResultComparator.SubsetRaw(actual, expected) &&
					actual.NextPage.String() == expectedNextPage &&
					actual.Done == expected.Done
			},
			Expected: &common.ReadResult{
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name":  "Patrick",
						"email": "patrick@gmail.com",
					},
					Raw: map[string]any{
						"type":       "contact",
						"id":         "66439b947bb095a681f7fd9e",
						"role":       "user",
						"email":      "patrick@gmail.com",
						"phone":      nil,
						"name":       "Patrick",
						"created_at": float64(1715706772),
						"updated_at": float64(1715706939),
					},
				}},
				NextPage: "{{testServerURL}}/contacts?per_page=60&starting_after=" +
					"Wy0xLCI2NjQzOWI5NDdiYjA5NWE2ODFmN2ZkOWUiLDNd",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read of named list",
			Input: common.ReadParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("state"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseReadConversations),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
					actual.NextPage.String() == expected.NextPage.String() &&
					actual.Done == expected.Done &&
					actual.Rows == expected.Rows
			},
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"state": "closed",
					},
				}, {
					Fields: map[string]any{
						"state": "open",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
