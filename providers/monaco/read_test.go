package monaco

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseContacts := testutils.DataFromFile(t, "read/contacts.json")
	responseContactsLastPage := testutils.DataFromFile(t, "read/contacts-last-page.json")
	responseContactsEmpty := testutils.DataFromFile(t, "read/contacts-empty.json")
	responseTags := testutils.DataFromFile(t, "read/tags.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: objectContacts},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "butterflies", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrResolvingURLPathForObject},
		},
		{
			Name: "Contacts list is a POST carrying page and page_size",
			Input: common.ReadParams{
				ObjectName: objectContacts,
				Fields:     connectors.Fields("id", "email"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/contacts/list"),
					mockcond.Body(`{"page":1,"page_size":100}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "con_abc123", "email": "jane@acme.com"},
					Raw:    map[string]any{"first_name": "Jane", "title": "VP Engineering"},
					Id:     "con_abc123",
				}, {
					Fields: map[string]any{"id": "con_xyz789", "email": "john@globex.com"},
					Raw:    map[string]any{"first_name": "John", "do_not_contact": true},
					Id:     "con_xyz789",
				}},
				// page 1 of 2 -> the token is simply the next page number.
				NextPage: "2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "NextPage token is fed back as the page number",
			Input: common.ReadParams{
				ObjectName: objectContacts,
				Fields:     connectors.Fields("id"),
				NextPage:   "2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/contacts/list"),
					mockcond.Body(`{"page":2,"page_size":100}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContactsLastPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "con_lmn456"},
					Raw:    map[string]any{"email": "ada@acme.com"},
					Id:     "con_lmn456",
				}},
				// page == total_pages, so the read is finished.
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Empty page reports done",
			Input: common.ReadParams{
				ObjectName: objectContacts,
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/contacts/list"),
				Then:  mockserver.Response(http.StatusOK, responseContactsEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Since and Until become updated_at filter rules",
			Input: common.ReadParams{
				ObjectName: objectContacts,
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2025, time.June, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/contacts/list"),
					mockcond.Body(`{
						"page": 1,
						"page_size": 100,
						"filters": [
							{"field":"updated_at","condition":"greater_than","value":"2025-06-01T00:00:00Z"},
							{"field":"updated_at","condition":"less_than","value":"2025-07-01T00:00:00Z"}
						]
					}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContactsEmpty),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Audiences ignores Since because its list request has no filters",
			Input: common.ReadParams{
				ObjectName: objectAudiences,
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2025, time.June, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/audiences/list"),
					// No filters key -- pushing one down would be rejected.
					mockcond.Body(`{"page":1,"page_size":100}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContactsEmpty),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "PageSize is clamped to Monaco's maximum of 500",
			Input: common.ReadParams{
				ObjectName: objectContacts,
				Fields:     connectors.Fields("id"),
				PageSize:   5000,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Body(`{"page":1,"page_size":500}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContactsEmpty),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Tags is a GET on the slash-terminated route and never paginates",
			Input: common.ReadParams{
				ObjectName: objectTags,
				Fields:     connectors.Fields("id", "name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					// The trailing slash is load-bearing: /v1/tags answers 307.
					mockcond.Path("/v1/tags/"),
				},
				Then: mockserver.Response(http.StatusOK, responseTags),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "tag_001", "name": "Interested"},
					Raw:    map[string]any{"color": "#22c55e"},
					Id:     "tag_001",
				}, {
					Fields: map[string]any{"id": "tag_002", "name": "Decision Maker"},
					Raw:    map[string]any{"color": "#3b82f6"},
					Id:     "tag_002",
				}},
				// No pagination block in the response.
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Sequence templates keep the unslashed route",
			Input: common.ReadParams{
				ObjectName: objectSequenceTemplates,
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					// Mirror image of tags: here the slashed form is the redirect.
					mockcond.Path("/v1/sequence-templates"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{"data":[]}`)),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			// Records are extracted with required semantics. A response missing
			// `data` violates Monaco's schema, and reporting it as an empty page
			// would let a sync finish early and silently.
			Name: "Missing data envelope is an error, not an empty page",
			Input: common.ReadParams{
				ObjectName: objectContacts,
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/contacts/list"),
				Then: mockserver.Response(http.StatusOK,
					[]byte(`{"pagination":{"page":1,"page_size":100,"total_count":0,"total_pages":1},"meta":{}}`)),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrKeyNotFound},
		},
		{
			Name: "Malformed NextPage is rejected",
			Input: common.ReadParams{
				ObjectName: objectContacts,
				Fields:     connectors.Fields("id"),
				NextPage:   "not-a-page",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrInvalidNextPage},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
