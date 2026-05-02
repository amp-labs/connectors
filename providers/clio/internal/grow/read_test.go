package grow

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) {
	t.Parallel()

	responseContactsEmpty := testutils.DataFromFile(t, "read-contacts-empty.json")
	responseContactsSinglePage := testutils.DataFromFile(t, "read-contacts-single-page.json")
	firstPageFixture := testutils.DataFromFile(t, "read-contacts-first-page.json")
	responseUsersSingleRecord := testutils.DataFromFile(t, "read-users-single-record.json")

	tests := []testroutines.Read{
		{
			Name: "Read contacts empty list",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/grow/contacts"),
					mockcond.QueryParam("limit", "1"),
					mockcond.QueryParam("order", "id(asc)"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactsEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read contacts first page",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/grow/contacts"),
					mockcond.QueryParam("limit", "1"),
					mockcond.QueryParam("order", "id(asc)"),
				},
				Then: mockserver.Response(http.StatusOK, firstPageFixture),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(123),
						"name": "John Doe",
					},
					Raw: map[string]any{
						"id":   float64(123),
						"name": "John Doe",
					},
				}},
				NextPage: "https://api.clio.com/grow/contacts?page_token=abc123",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read contacts with updated_since",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name"),
				PageSize:   1,
				Since:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/grow/contacts"),
					mockcond.QueryParam("limit", "1"),
					mockcond.QueryParam("order", "id(asc)"),
					mockcond.QueryParam("updated_since", "2026-04-01T00:00:00Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactsSinglePage),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(123),
						"name": "John Doe",
					},
					Raw: map[string]any{
						"id":   float64(123),
						"name": "John Doe",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read contacts next page via NextPage",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name"),
				PageSize:   1,
				NextPage:   testroutines.URLTestServer + "/grow/contacts?page_token=abc123",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/grow/contacts"),
					mockcond.QueryParam("page_token", "abc123"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactsSinglePage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(123),
						"name": "John Doe",
					},
					Raw: map[string]any{
						"id":   float64(123),
						"name": "John Doe",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read users single record required fields no pagination",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "first_name", "last_name", "email"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/grow/users"),
					mockcond.QueryParam("limit", "1"),
					mockcond.QueryParam("order", "id(asc)"),
					mockcond.Permute(
						func(fields []string) mockcond.Condition {
							return mockcond.QueryParam("fields", strings.Join(fields, ","))
						},
						"id", "first_name", "last_name", "email",
					),
				},
				Then: mockserver.Response(http.StatusOK, responseUsersSingleRecord),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{

						"first_name": "Jane",
						"last_name":  "Doe",
						"email":      "jane.doe@example.com",
					},
					Raw: map[string]any{

						"first_name": "Jane",
						"last_name":  "Doe",
						"email":      "jane.doe@example.com",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructReadTestAdapter(tt.Server.URL)
			})
		})
	}
}

func constructReadTestAdapter(serverURL string) (*Adapter, error) {
	adapter, err := NewAdapter(common.ConnectorParams{
		Module:              providers.ModuleClioGrow,
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "api.clio.com",
		Metadata: map[string]string{
			"region": "",
		},
	})
	if err != nil {
		return nil, err
	}

	adapter.SetBaseURL(mockutils.ReplaceURLOrigin(adapter.HTTPClient().Base, serverURL))

	return adapter, nil
}
