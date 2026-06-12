package zoominfo

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen
	t.Parallel()

	contactsResponse := testutils.DataFromFile(t, "read-contacts.json")
	industriesResponse := testutils.DataFromFile(t, "read-industries.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Search read without Since defaults the required date criterion to epoch",
			Input: common.ReadParams{
				ObjectName: objContacts,
				Fields:     connectors.Fields("firstName"),
				NextPage:   "2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/gtm/data/v1/contacts/search"),
					mockcond.QueryParam("page[number]", "2"),
					mockcond.QueryParam("page[size]", "100"),
					mockcond.Body(`{"data":{"type":"ContactSearch","attributes":{"lastUpdatedDateAfter":"1970-01-01T00:00:00Z"}}}`),
				},
				Then: mockserver.Response(http.StatusOK, contactsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"firstName": "Ada"},
						Raw: map[string]any{
							"id":        "1",
							"type":      "Contact",
							"firstName": "Ada",
							"lastName":  "Lovelace",
							"jobTitle":  "Engineer",
						},
					},
					{
						Fields: map[string]any{"firstName": "Alan"},
						Raw: map[string]any{
							"id":        "2",
							"type":      "Contact",
							"firstName": "Alan",
							"lastName":  "Turing",
							"jobTitle":  "Scientist",
						},
					},
				},
				NextPage: "3",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Incremental Since maps to lastUpdatedDateAfter for contacts",
			Input: common.ReadParams{
				ObjectName: objContacts,
				Fields:     connectors.Fields("firstName"),
				Since:      time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/gtm/data/v1/contacts/search"),
					mockcond.Body(`{"data":{"type":"ContactSearch","attributes":{"lastUpdatedDateAfter":"2026-06-01T00:00:00Z"}}}`),
				},
				Then: mockserver.Response(http.StatusOK, contactsResponse),
			}.Server(),
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 2, Done: false, NextPage: "3"},
			ExpectedErrs: nil,
		},
		{
			Name: "Since and Until map to pageDateMin/pageDateMax for news",
			Input: common.ReadParams{
				ObjectName: objNews,
				Fields:     connectors.Fields("title"),
				Since:      time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/gtm/data/v1/news/search"),
					mockcond.Body(
						`{"data":{"type":"NewsSearch","attributes":` +
							`{"pageDateMin":"2026-06-01T00:00:00Z","pageDateMax":"2026-06-30T00:00:00Z"}}}`,
					),
				},
				Then: mockserver.Response(http.StatusOK, industriesResponse),
			}.Server(),
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 1, Done: true},
			ExpectedErrs: nil,
		},
		{
			// Companies has no date field, so it sends empty criteria (its search
			// API accepts that and returns all records) — Since is ignored.
			Name: "Object without a date field sends empty criteria",
			Input: common.ReadParams{
				ObjectName: objCompanies,
				Fields:     connectors.Fields("name"),
				Since:      time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/gtm/data/v1/companies/search"),
					mockcond.Body(`{"data":{"type":"CompanySearch","attributes":{}}}`),
				},
				Then: mockserver.Response(http.StatusOK, industriesResponse),
			}.Server(),
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 1, Done: true},
			ExpectedErrs: nil,
		},
		{
			// audience-folders is paginated but NOT incremental: its documented
			// filter[updatedAfter] is broken server-side, so Since must be ignored
			// (no filter[updatedAfter] is sent) rather than 400 every read.
			Name: "Paginated GET object ignores Since (no updated-since filter)",
			Input: common.ReadParams{
				ObjectName: "audience-folders",
				Fields:     connectors.Fields("name"),
				Since:      time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/gtm/studio/v1/folders"),
					mockcond.QueryParam("page[size]", "100"),
					mockcond.QueryParamsMissing("filter[updatedAfter]"),
				},
				Then: mockserver.Response(http.StatusOK, industriesResponse),
			}.Server(),
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 1, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Lookup object reads a single unpaginated page",
			Input: common.ReadParams{ObjectName: objIndustries, Fields: connectors.Fields("name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/gtm/data/v1/lookup/industries"),
				Then:  mockserver.Response(http.StatusOK, industriesResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"name": "Software"},
						Raw: map[string]any{
							"id":   "software",
							"type": "Industry",
							"name": "Software",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}
