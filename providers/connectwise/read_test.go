package connectwise

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorNotFound := testutils.DataFromFile(t, "read/not-found.json")
	responseContacts := testutils.DataFromFile(t, "read/contacts.json")
	responseContactNoItems := testutils.DataFromFile(t, "read/contact-frank-fenrir.json")
	responseContactPartialItems := testutils.DataFromFile(t, "read/contact-jorge-smith.json")
	responseContactAllItems := testutils.DataFromFile(t, "read/contact-emma-diaz.json")
	responseCampaignsFirstPage := []byte(`[{"name": "Weekly Update #33", "id": 123}]`)
	responseExportsEmpty := []byte(`[]`)

	nextPageRaw := common.NextPageToken("https://sandbox-na.myconnectwise.net/v4_6_release/apis/3.0/company/contacts/?conditions=LastUpdated+%3e%3d+%5b2025-04-01T20%3a02%3a28Z%5d&pageSize=2&page=2")
	linkHeaderRaw := "<" + nextPageRaw.String() + ">; rel=\"next\""
	nextPageRelative := common.NextPageToken(testconn.URLTestServer + "/v4_6_release/apis/3.0/company/contacts/?conditions=LastUpdated+%3e%3d+%5b2025-04-01T20%3a02%3a28Z%5d&pageSize=2&page=2")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "messages"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Error response page not found",
			Input: common.ReadParams{ObjectName: "messages", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				testutils.StringError("cannot resolve URL path for given object name"),
			},
		},
		{
			Name: "Campaigns first page has a link to next",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v4_6_release/apis/3.0/marketing/campaigns"),
					mockcond.Header(http.Header{"ClientId": []string{"dummy-client-id"}}),
				},
				Then: mockserver.ResponseChainedFuncs(
					mockserver.Header("Link", linkHeaderRaw),
					mockserver.Response(http.StatusOK, responseCampaignsFirstPage),
				),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name": "Weekly Update #33",
					},
					Raw: map[string]any{
						"name": "Weekly Update #33",
						"id":   float64(123),
					},
				}},
				NextPage: nextPageRaw,
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read campaigns empty page",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v4_6_release/apis/3.0/marketing/campaigns"),
					mockcond.Header(http.Header{"ClientId": []string{"dummy-client-id"}}),
				},
				Then: mockserver.ResponseString(http.StatusOK, `[]`),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Read empty exports with null array",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v4_6_release/apis/3.0/marketing/campaigns"),
					mockcond.Header(http.Header{"ClientId": []string{"dummy-client-id"}}),
				},
				Then: mockserver.Response(http.StatusOK, responseExportsEmpty),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Contacts first page has a link to next",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("firstName", "lastName"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v4_6_release/apis/3.0/company/contacts"),
					mockcond.QueryParamsMissing("conditions"),
					mockcond.Header(http.Header{"ClientId": []string{"dummy-client-id"}}),
				},
				Then: mockserver.ResponseChainedFuncs(
					mockserver.Header("Link", linkHeaderRaw),
					mockserver.Response(http.StatusOK, responseContacts),
				),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"firstname": "Alex",
						"lastname":  "Morgan",
					},
					Raw: map[string]any{
						"id":        float64(31045),
						"firstName": "Alex",
						"lastName":  "Morgan",
					},
				}},
				NextPage: nextPageRaw,
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Contacts next page consumption",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields: connectors.Fields("firstName", "lastName",
					"customField22", // sync_status.
				),
				NextPage: nextPageRelative,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v4_6_release/apis/3.0/company/contacts/"),
					mockcond.QueryParam("pageSize", "2"),
					mockcond.QueryParam("page", "2"),
					mockcond.QueryParam("conditions", "LastUpdated >= [2025-04-01T20:02:28Z]"),
					mockcond.Header(http.Header{"ClientId": []string{"dummy-client-id"}}),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"firstname":     "Alex",
						"lastname":      "Morgan",
						"customfield22": "pending", // The `Fields` are always lower case
					},
					Raw: map[string]any{
						"title": "Operations Manager",
					},
					Id: "31045",
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Contact with no communication items",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields: connectors.Fields(
					"AMPERSAND-defaultEmail", "AMPERSAND-defaultFax", "AMPERSAND-defaultPhone",
				),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseContactNoItems),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"ampersand-defaultEmail": nil,
						"ampersand-defaultFax":   nil,
						"ampersand-defaultPhone": nil,
					},
					Raw: map[string]any{"id": float64(58001)},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Contact with some communication items of same phone type",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("AMPERSAND-defaultPhone", "AMPERSAND-defaultPhoneId"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseContactPartialItems),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"ampersand-defaultphone":   "+13344455",
						"ampersand-defaultphoneid": "22",
					},
					Raw: map[string]any{"id": float64(1)},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Contact with all communication items",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields: connectors.Fields(
					"AMPERSAND-defaultEmail", "AMPERSAND-defaultFax", "AMPERSAND-defaultPhone",
					"AMPERSAND-defaultEmailId", "AMPERSAND-defaultFaxId", "AMPERSAND-defaultPhoneId",
				),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseContactAllItems),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"ampersand-defaultemail":   "emma@test.com",
						"ampersand-defaultemailid": "14",
						"ampersand-defaultfax":     "6548568",
						"ampersand-defaultfaxid":   "20",
						"ampersand-defaultphone":   "+166155555",
						"ampersand-defaultphoneid": "27",
					},
					Raw: map[string]any{"id": float64(58118)},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}
