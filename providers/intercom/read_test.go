package intercom

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
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
	requestSearchConversations := testutils.DataFromFile(t, "read-search-conversations-request.json")
	responseSearchConversations := testutils.DataFromFile(t, "read-search-conversations.json")
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
			Expected:     &common.ReadResult{Done: true, Data: []common.ReadResultRow{}},
			ExpectedErrs: nil,
		},
		{
			Name:  "API version header is passed as server request",
			Input: common.ReadParams{ObjectName: "articles", Fields: connectors.Fields("id")},
			// notes is not supported for now, but its payload is good for testing
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Header(testApiVersionHeader),
				Then:  mockserver.Response(http.StatusOK, responseNotesSecondPage),
			}.Server(),
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 1, NextPage: "", Done: true},
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
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     2,
				NextPage: "https://api.intercom.io/contacts/6643703ffae7834d1792fd30/notes?per_page=2&page=2",
				Done:     false,
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
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				NextPage: testroutines.URLTestServer + "/contacts?per_page=60&starting_after=" +
					"WzE3MTU2OTU2NzkwMDAsIjY2NDM3MDNmZmFlNzgzNGQxNzkyZmQzMCIsMl0=",
				Done: false,
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
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 1, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Next page URL is empty, when provided with missing object",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseContactsThirdPage),
			}.Server(),
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 1, NextPage: "", Done: true},
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
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
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
				NextPage: testroutines.URLTestServer + "/contacts?per_page=60&starting_after=" +
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
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"state": "closed",
					},
					Raw: map[string]any{
						"state": "closed",
					},
				}, {
					Fields: map[string]any{
						"state": "open",
					},
					Raw: map[string]any{
						"state": "open",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Incremental read of conversations via search",
			Input: common.ReadParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("id", "state", "title"),
				Since:      time.Unix(1726674883, 0),
			},
			// notes is not supported for now, but its payload is good for testing
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.BodyBytes(requestSearchConversations),
				Then:  mockserver.Response(http.StatusOK, responseSearchConversations),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    "5",
						"state": "open",
						"title": "What is return policy?",
					},
					Raw: map[string]any{
						"ai_agent_participated": false,
						"created_at":            float64(1726752048),
						"updated_at":            float64(1726752145),
					},
				}},
				NextPage: testroutines.URLTestServer + "/conversations/search?starting_after=WzE3MjY3NTIxNDUwMDAsNSwyXQ==",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
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
