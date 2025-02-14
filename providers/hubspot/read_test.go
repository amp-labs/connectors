package hubspot

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseContacts := testutils.DataFromFile(t, "read-contacts-objects-api.json")
	responseListsFirst := testutils.DataFromFile(t, "read-lists-1-first-page.json")
	responseListsLast := testutils.DataFromFile(t, "read-lists-2-second-page.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
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
			Name: "Contacts uses object API endpoint",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/crm/v3/objects/contacts"),
				Then:  mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"email": "a@example.com",
						},
						Id: "1",
						Raw: map[string]any{
							"id": "1",
							"properties": map[string]any{
								"createdate":       "2023-10-26T17:55:48.301Z",
								"email":            "a@example.com",
								"lastmodifieddate": "2024-12-24T17:31:54.727Z",
							},
							"createdAt": "2023-10-26T17:55:48.301Z",
							"updatedAt": "2024-12-24T17:31:54.727Z",
							"archived":  false,
						},
					}, {
						Fields: map[string]any{
							"email": "b@example.com",
						},
						Id: "51",
						Raw: map[string]any{
							"id": "51",
							"properties": map[string]any{
								"createdate":       "2023-10-26T17:55:48.691Z",
								"email":            "b@example.com",
								"lastmodifieddate": "2023-12-13T22:45:30.353Z",
							},
							"createdAt": "2023-10-26T17:55:48.691Z",
							"updatedAt": "2023-12-13T22:45:30.353Z",
							"archived":  false,
						},
					}, {
						Fields: map[string]any{
							"email": "c@example.com",
						},
						Id: "101",
						Raw: map[string]any{
							"id": "101",
							"properties": map[string]any{
								"createdate":       "2023-12-13T22:20:02.649Z",
								"email":            "c@example.com",
								"lastmodifieddate": "2023-12-13T22:20:05.498Z",
							},
							"createdAt": "2023-12-13T22:20:02.649Z",
							"updatedAt": "2023-12-13T22:20:05.498Z",
							"archived":  false,
						},
					},
				},
				NextPage: "https://api.hubapi.com/crm/v3/objects/contacts?limit=100&properties=listId%2Cname&after=394", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Lists first page is done via search",
			Input: common.ReadParams{
				ObjectName: "lists",
				Fields:     connectors.Fields("processingType"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/crm/v3/lists/search"),
				Then:  mockserver.Response(http.StatusOK, responseListsFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"processingtype": "DYNAMIC",
					},
					Raw: map[string]any{
						// "listId": "3",
						"name": "Test List",
					},
				}, {
					Fields: map[string]any{
						"processingtype": "SNAPSHOT",
					},
					Raw: map[string]any{
						// "listId": "4",
						"name": "Test static company list",
					},
				}},
				NextPage: "2", // Next page token is in fact an offset.
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Lists next page sends offset in payload",
			Input: common.ReadParams{
				ObjectName: "lists",
				Fields:     connectors.Fields("name"),
				NextPage:   "2", // Move offset 2 records ahead to get next page.
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.PathSuffix("/crm/v3/lists/search"),
					mockcond.Body(`{
						"offset": 2,
						"count": 100
					}`),
				},
				Then: mockserver.Response(http.StatusOK, responseListsLast),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				Data:     []common.ReadResultRow{},
				NextPage: "", // empty next page is inferred from response
				Done:     true,
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
