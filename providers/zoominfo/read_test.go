package zoominfo

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/spyzhov/ajson"
)

func TestRead(t *testing.T) { // nolint:funlen
	t.Parallel()

	contactsResponse := testutils.DataFromFile(t, "read-contacts.json")
	contactsFinalPageResponse := testutils.DataFromFile(t, "read-contacts-final-page.json")
	industriesResponse := testutils.DataFromFile(t, "read-industries.json")

	tests := []testconn.TestCaseRead{
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
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/gtm/data/v1/contacts/search"),
					mockcond.QueryParam("page[number]", "2"),
					mockcond.QueryParam("page[size]", "2"),
					mockcond.Body(`{"data":{"type":"ContactSearch","attributes":{"lastUpdatedDateAfter":"1970-01-01T00:00:00Z"}}}`),
				},
				Then: mockserver.Response(http.StatusOK, contactsResponse),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
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
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/gtm/data/v1/contacts/search"),
					mockcond.Body(`{"data":{"type":"ContactSearch","attributes":{"lastUpdatedDateAfter":"2026-06-01T00:00:00Z"}}}`),
				},
				Then: mockserver.Response(http.StatusOK, contactsResponse),
			}.Server(),
			Comparator:   testconn.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 2, Done: false, NextPage: "3"},
			ExpectedErrs: nil,
		},
		{
			// Regression test for the PFAPI0004 overrun. The final page carries no
			// links.next, so the read must finish here. meta.page.total is 2 (the
			// record count of this page, which is what ZoomInfo actually reports)
			// against meta.page.number 1 — the old number-vs-total comparison read
			// that as "more pages follow" and asked for page 2, which the API
			// rejects with "Page number (page) requested is greater than the
			// available results".
			Name: "Final page without links.next ends the read",
			Input: common.ReadParams{
				ObjectName: objContacts,
				Fields:     connectors.Fields("firstName"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/gtm/data/v1/contacts/search"),
				Then:  mockserver.Response(http.StatusOK, contactsFinalPageResponse),
			}.Server(),
			Comparator:   testconn.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 2, Done: true, NextPage: ""},
			ExpectedErrs: nil,
		},
		{
			// Defence in depth: a page shorter than the requested page[size] ends the
			// read even though this fixture does advertise links.next.
			Name: "Page shorter than page[size] ends the read",
			Input: common.ReadParams{
				ObjectName: objContacts,
				Fields:     connectors.Fields("firstName"),
				PageSize:   100,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/gtm/data/v1/contacts/search"),
				Then:  mockserver.Response(http.StatusOK, contactsResponse),
			}.Server(),
			Comparator:   testconn.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 2, Done: true, NextPage: ""},
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
			Comparator:   testconn.ComparatorPagination,
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
			Comparator:   testconn.ComparatorPagination,
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
			Comparator:   testconn.ComparatorPagination,
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
			Comparator: testconn.ComparatorPagination,
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

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

// TestNextPageFromLinks pins the next-page extraction to link values copied verbatim
// from live ZoomInfo responses. The two surfaces serialise "links" differently — the
// Data API emits strings prefixed with "uri=" and a relative path, Studio and Copilot
// emit absolute URLs — so both forms are covered here.
func TestNextPageFromLinks(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []struct {
		name     string
		pageSize int
		body     string
		expected string
		wantErr  error
	}{
		{
			// Verbatim links from POST /gtm/data/v1/contacts/search page 2 of 100.
			name:     "Data API next link carries the uri= prefix",
			pageSize: 2,
			body: `{"data":[{"id":"1"},{"id":"2"}],"links":{
				"first":"uri=/data/v1/contacts/search?page[number]=1&page[size]=100",
				"prev":"uri=/data/v1/contacts/search?page[number]=1&page[size]=100",
				"next":"uri=/data/v1/contacts/search?page[number]=3&page[size]=100",
				"last":"uri=/data/v1/contacts/search?page[number]=100&page[size]=100"}}`,
			expected: "3",
		},
		{
			// Studio links are absolute URLs and carry extra query params.
			name:     "Studio next link is an absolute URL",
			pageSize: 2,
			body: `{"data":[{"id":"1"},{"id":"2"}],"links":{
				"first":"https://api.zoominfo.com/gtm/studio/v1/audiences?page[number]=1&page[size]=100&sort=-updatedAt",
				"next":"https://api.zoominfo.com/gtm/studio/v1/audiences?page[number]=2&page[size]=100&sort=-updatedAt",
				"last":"https://api.zoominfo.com/gtm/studio/v1/audiences?page[number]=9&page[size]=100&sort=-updatedAt"}}`,
			expected: "2",
		},
		{
			// The final page omits links.next. meta.page.total repeats the record
			// count of this page, which must not be mistaken for a page count.
			name:     "Final page omits next",
			pageSize: 2,
			body: `{"data":[{"id":"1"},{"id":"2"}],"meta":{"page":{"number":1,"total":2},"totalResults":2},
				"links":{"first":"uri=/data/v1/contacts/search?page[number]=1&page[size]=2",
				"last":"uri=/data/v1/contacts/search?page[number]=1&page[size]=2"}}`,
			expected: "",
		},
		{
			name:     "Unpaginated endpoint omits links entirely",
			pageSize: 100,
			body:     `{"data":[{"id":"software"}]}`,
			expected: "",
		},
		{
			name:     "Short page ends the read despite a next link",
			pageSize: 100,
			body: `{"data":[{"id":"1"}],"links":{
				"next":"uri=/data/v1/contacts/search?page[number]=2&page[size]=100"}}`,
			expected: "",
		},
		{
			name:     "Unusable next link is an error, never a silent stop",
			pageSize: 1,
			body:     `{"data":[{"id":"1"}],"links":{"next":"uri=/data/v1/contacts/search?cursor=abc"}}`,
			wantErr:  common.ErrNextPageInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			node, err := ajson.Unmarshal([]byte(tt.body))
			if err != nil {
				t.Fatalf("bad test fixture: %v", err)
			}

			token, err := nextPageFromLinks(tt.pageSize)(node)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if token != tt.expected {
				t.Errorf("expected next page %q, got %q", tt.expected, token)
			}
		})
	}
}
