package salesforce

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"gotest.tools/v3/assert"
)

func soqlFilter(condition any) *common.RawFilter {
	return &common.RawFilter{Type: core.RawFilterTypeSOQL, Filter: condition}
}

func TestConnectorValidateRawFilter(t *testing.T) {
	t.Parallel()

	server := mockserver.Dummy()
	defer server.Close()

	crm, err := constructTestConnector(server.URL)
	assert.NilError(t, err)

	pardot, err := constructTestConnectorAccountEngagement(server.URL)
	assert.NilError(t, err)

	assert.NilError(t, crm.ValidateRawFilter(*soqlFilter("IsDeleted = true OR IsDeleted = false")))

	for name, err := range map[string]error{
		// The type declares the filter's syntax. Salesforce only runs SOQL, so a filter written in any
		// other syntax (here "sql") must not be sent to it as if it were SOQL.
		"unsupported type":  crm.ValidateRawFilter(common.RawFilter{Type: "sql", Filter: "IsDeleted = true"}),
		"non-string filter": crm.ValidateRawFilter(*soqlFilter(42)),
		"empty filter":      crm.ValidateRawFilter(*soqlFilter("")),
		// Account Engagement (Pardot) reads don't use SOQL and never apply a raw filter.
		// Accepting one would silently read unfiltered records, so it is rejected up front.
		"account engagement": pardot.ValidateRawFilter(*soqlFilter("IsDeleted = true")),
	} {
		assert.Assert(t, errors.Is(err, common.ErrInvalidRawFilter), "%s: got error %v", name, err)
	}
}

func TestValidateRawFilterParentheses(t *testing.T) {
	t.Parallel()

	balanced := []string{
		"IsDeleted = true OR IsDeleted = false",
		"Status = 'Qualified' AND (Country = 'US' OR Country = 'CA')",
		"Id IN (SELECT ContactId FROM CampaignMember WHERE Status = 'Sent')",
		"Name = 'unmatched ) inside a literal'",
		`Name = 'O\'Brien (' AND Title != null`,
	}
	for _, condition := range balanced {
		_, err := common.RawFilterAs[core.SOQLFilter](*soqlFilter(condition))
		assert.NilError(t, err, condition)
	}

	unbalanced := []string{
		"IsDeleted = true) OR (Id != null",
		"(IsDeleted = true",
		"IsDeleted = true)",
		"Name = 'x') OR (Id != null",
	}
	for _, condition := range unbalanced {
		_, err := common.RawFilterAs[core.SOQLFilter](*soqlFilter(condition))
		assert.Assert(t, errors.Is(err, common.ErrInvalidRawFilter), "%s: got error %v", condition, err)
	}
}

func TestMakeSOQLRawFilterIsGroupedAfterTimeWindow(t *testing.T) {
	t.Parallel()

	soql, err := makeSOQL(common.ReadParams{
		ObjectName: "Account",
		Fields:     datautils.NewSet("Name"),
		Since:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Until:      time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		RawFilter:  soqlFilter("IsDeleted = true OR IsDeleted = false"),
	}, defaultTimestampColumn)
	assert.NilError(t, err)

	assert.Equal(t, soql.String(), "SELECT Id,Name FROM Account WHERE "+
		"SystemModstamp > 2024-01-15T00:00:00Z AND SystemModstamp <= 2024-06-15T00:00:00Z AND "+
		"(IsDeleted = true OR IsDeleted = false)")
}

func TestQueryEndpoint(t *testing.T) {
	t.Parallel()

	assert.Equal(t, core.QueryEndpoint(false), "query")
	assert.Equal(t, core.QueryEndpoint(true), "queryAll")
}

func TestMakeSOQLRejectsInvalidRawFilter(t *testing.T) {
	t.Parallel()

	// Every SOQL path (read, bulk read, records by ID) builds through makeSOQL, so all of them reject it.
	_, err := makeSOQL(common.ReadParams{
		ObjectName: "Account",
		Fields:     datautils.NewSet("Name"),
		RawFilter:  soqlFilter("IsDeleted = true) OR (Id != null"),
	}, defaultTimestampColumn)
	assert.Assert(t, errors.Is(err, common.ErrInvalidRawFilter), "got error %v", err)
}

func TestReadRawFilter(t *testing.T) { //nolint:funlen
	t.Parallel()

	contactsDeletedAndLive := []byte(`{
		"totalSize": 2,
		"done": true,
		"records": [
			{"attributes": {"type": "Contact"}, "Id": "003A", "LastName": "Gone", "IsDeleted": true},
			{"attributes": {"type": "Contact"}, "Id": "003B", "LastName": "Here", "IsDeleted": false}
		]
	}`)
	contactsWithoutIsDeleted := []byte(`{
		"totalSize": 1,
		"done": true,
		"records": [{"attributes": {"type": "Contact"}, "Id": "003A", "LastName": "Gone"}]
	}`)
	contactsSecondPage := []byte(`{
		"totalSize": 3,
		"done": true,
		"records": [{"attributes": {"type": "Contact"}, "Id": "003C", "LastName": "Late", "IsDeleted": true}]
	}`)

	tests := []testconn.TestCaseRead{
		{
			Name: "Raw filter of another type is rejected before any request",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("LastName"),
				RawFilter:  &common.RawFilter{Type: "sql", Filter: "IsDeleted = true"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrInvalidRawFilter},
		},
		{
			Name: "Raw filter reads queryAll; IsDeleted is delivered when selected",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("LastName", "IsDeleted"),
				RawFilter:  soqlFilter("IsDeleted = true OR IsDeleted = false"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/services/data/v60.0/queryAll"),
					mockcond.Permute(
						queryParam("SELECT %v FROM contacts WHERE (IsDeleted = true OR IsDeleted = false)"),
						"Id", "LastName", "IsDeleted",
					),
				},
				Then: mockserver.Response(http.StatusOK, contactsDeletedAndLive),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id:     "003A",
					Fields: map[string]any{"lastname": "Gone", "isdeleted": true},
				}, {
					Id:     "003B",
					Fields: map[string]any{"lastname": "Here", "isdeleted": false},
				}},
				Done: true,
			},
		},
		{
			Name: "Raw filter selects only the requested fields",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("LastName"),
				RawFilter:  soqlFilter("IsDeleted = true"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/services/data/v60.0/queryAll"),
					mockcond.Permute(queryParam("SELECT %v FROM contacts WHERE (IsDeleted = true)"), "Id", "LastName"),
				},
				Then: mockserver.Response(http.StatusOK, contactsWithoutIsDeleted),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id:     "003A",
					Fields: map[string]any{"lastname": "Gone"},
				}},
				Done: true,
			},
		},
		{
			Name: "Next page follows the queryAll URL",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("LastName", "IsDeleted"),
				RawFilter:  soqlFilter("IsDeleted = true OR IsDeleted = false"),
				NextPage:   "/services/data/v60.0/queryAll/01gA-2000",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/services/data/v60.0/queryAll/01gA-2000"),
				Then:  mockserver.Response(http.StatusOK, contactsSecondPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id:     "003C",
					Fields: map[string]any{"lastname": "Late", "isdeleted": true},
				}},
				Done: true,
			},
		},
		{
			Name: "Read without raw filter keeps the query endpoint",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("LastName"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/services/data/v60.0/query"),
					mockcond.Permute(queryParam("SELECT %v FROM contacts"), "Id", "LastName"),
				},
				Then: mockserver.Response(http.StatusOK, contactsWithoutIsDeleted),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 1, Done: true},
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

func TestSearchRawFilter(t *testing.T) { //nolint:funlen
	t.Parallel()

	contactsDeletedAndLive := []byte(`{
		"totalSize": 2,
		"done": true,
		"records": [
			{"attributes": {"type": "Contact"}, "Id": "003A", "LastName": "Gone", "IsDeleted": true},
			{"attributes": {"type": "Contact"}, "Id": "003B", "LastName": "Here", "IsDeleted": false}
		]
	}`)

	tests := []testconn.TestCaseSearch{
		{
			Name: "Raw filter of another type is rejected before any request",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("LastName"),
				RawFilter:  &common.RawFilter{Type: "sql", Filter: "IsDeleted = true"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrInvalidRawFilter},
		},
		{
			Name: "Raw filter and field filters cannot be combined",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("LastName"),
				Filter:     connectors.SearchFilter{}.FilterBy("LastName", common.FilterOperatorEQ, "Gone"),
				RawFilter:  soqlFilter("IsDeleted = true"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrSearchFiltersCombined},
		},
		{
			Name: "Raw filter searches queryAll",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("LastName", "IsDeleted"),
				RawFilter:  soqlFilter("IsDeleted = true OR IsDeleted = false"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/services/data/v60.0/queryAll"),
					mockcond.Permute(
						queryParam("SELECT %v FROM contacts WHERE (IsDeleted = true OR IsDeleted = false)"),
						"Id", "LastName", "IsDeleted",
					),
				},
				Then: mockserver.Response(http.StatusOK, contactsDeletedAndLive),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id:     "003A",
					Fields: map[string]any{"lastname": "Gone", "isdeleted": true},
				}, {
					Id:     "003B",
					Fields: map[string]any{"lastname": "Here", "isdeleted": false},
				}},
				Done: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableSearcher, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
