package search

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/connectwise/internal/core"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestSearch(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseContacts := testutils.DataFromFile(t, "contact-jorge-smith.json")

	tests := []testconn.TestCaseSearch{
		{
			Name:         "Object name must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one search filter is requested",
			Input:        common.SearchParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingSearchFilters},
		},
		{
			Name: "Searching contacts by equality condition",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("AMPERSAND-defaultPhone", "AMPERSAND-defaultPhoneId", "firstName"),
				Filter:     connectors.SearchFilter{}.FilterBy("firstName", "eq", "Jorge"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v4_6_release/apis/3.0/company/contacts"),
					mockcond.QueryParam("conditions", `firstName = "Jorge"`),
					mockcond.QueryParam("pageSize", "1000"),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "1",
					Fields: map[string]any{
						"ampersand-defaultphone":   "+13344455",
						"ampersand-defaultphoneid": "22",
						"firstname":                "Jorge",
					},
					Raw: map[string]any{"lastName": "Smith"}},
				},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Searching contacts by inclusion condition",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("AMPERSAND-defaultPhone"),
				Filter:     connectors.SearchFilter{}.FilterBy("id", "in", []string{"1", "2", "3"}),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v4_6_release/apis/3.0/company/contacts"),
					mockcond.QueryParam("conditions", `id in (1,2,3)`),
					mockcond.QueryParam("pageSize", "1000"),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id:     "1",
					Fields: map[string]any{"ampersand-defaultphone": "+13344455"},
					Raw:    map[string]any{"lastName": "Smith"}},
				},
				NextPage: "",
				Done:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableSearcher, error) {
				return constructTestStrategy(tt.Server)
			})
		})
	}
}

func constructTestStrategy(server *httptest.Server) (*Strategy, error) {
	base, err := core.NewBase(common.ConnectorParams{
		AuthenticatedClient: server.Client(),
		Metadata:            map[string]string{"clientId": "test-client-id"},
	})
	if err != nil {
		return nil, err
	}

	base.SetUnitTestMockServerBaseUrl(server.URL)

	return NewStrategy(base), nil
}
