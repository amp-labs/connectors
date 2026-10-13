package lightfield

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseUnauthorized := testutils.DataFromFile(t, "error-unauthorized.json")
	responseAccountsFirstPage := testutils.DataFromFile(t, "accounts-first-page.json")
	responseAccountsLastPage := testutils.DataFromFile(t, "accounts-last-page.json")
	responseMembers := testutils.DataFromFile(t, "members.json")
	responseAccountsKeyClash := testutils.DataFromFile(t, "accounts-key-clash.json")

	versionHeader := http.Header{"Lightfield-Version": []string{"2026-03-01"}}

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "accounts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "notes", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnauthorized, responseUnauthorized),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrAccessToken,
				testutils.StringError("Invalid API key."),
			},
		},
		{
			Name: "First page of accounts is flattened and links to the next offset",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     connectors.Fields("$name", "icp-score", "$owner", "updatedAt"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/accounts"),
					mockcond.QueryParam("limit", "2"),
					mockcond.QueryParam("offset", "0"),
					mockcond.Header(versionHeader),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id: "acc_2",
					Fields: map[string]any{
						"$name":     "Globex",
						"icp-score": float64(7),
						"$owner":    []any{"usr_1"},
						"updatedat": "2026-09-20T08:15:00.000Z",
					},
					Raw: map[string]any{
						"httpLink": "https://crm.lightfield.app/crm/accounts/acc_2",
					},
				}, {
					Id: "acc_1",
					Fields: map[string]any{
						"$name":     "Acme Corp",
						"icp-score": nil,
						"$owner":    []any{},
						"updatedat": nil,
					},
					Raw: map[string]any{
						"externalId": "sf-001",
					},
				}},
				NextPage: testconn.URLTestServer + "/v1/accounts?limit=2&offset=2",
				Done:     false,
			},
		},
		{
			Name: "Short page of accounts is the last one",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     connectors.Fields("$name"),
				NextPage:   testconn.URLTestServer + "/v1/accounts?limit=2&offset=2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/accounts"),
					mockcond.QueryParam("limit", "2"),
					mockcond.QueryParam("offset", "2"),
					mockcond.Header(versionHeader),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsLastPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id:     "acc_0",
					Fields: map[string]any{"$name": "Initech"},
					Raw:    map[string]any{"id": "acc_0"},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Incremental read of accounts sends updatedAt filters and returns every record it gets",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     connectors.Fields("$name"),
				PageSize:   2,
				Since:      time.Date(2026, 9, 1, 0, 0, 0, 123_000_000, time.UTC),
				Until:      time.Date(2026, 9, 21, 0, 0, 0, 456_000_000, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/accounts"),
					mockcond.QueryParam("limit", "2"),
					mockcond.QueryParam("offset", "0"),
					mockcond.QueryParam("$updatedAt[greaterThan]", "2026-09-01T00:00:00.123Z"),
					mockcond.QueryParam("$updatedAt[lessThanOrEqual]", "2026-09-21T00:00:00.456Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				// The provider filters; acc_1 (never updated) is not dropped by the connector.
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id:     "acc_2",
					Fields: map[string]any{"$name": "Globex"},
					Raw:    map[string]any{"id": "acc_2"},
				}, {
					Id:     "acc_1",
					Fields: map[string]any{"$name": "Acme Corp"},
					Raw:    map[string]any{"id": "acc_1"},
				}},
				NextPage: testconn.URLTestServer + "/v1/accounts?" +
					"%24updatedAt%5BgreaterThan%5D=2026-09-01T00%3A00%3A00.123Z&" +
					"%24updatedAt%5BlessThanOrEqual%5D=2026-09-21T00%3A00%3A00.456Z&limit=2&offset=2",
				Done: false,
			},
		},
		{
			Name: "Since alone is sent as an exclusive updatedAt bound with milliseconds",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 10, 12, 22, 7, 48, 174_000_000, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/contacts"),
					mockcond.QueryParam("$updatedAt[greaterThan]", "2026-10-12T22:07:48.174Z"),
					mockcond.QueryParamsMissing("$updatedAt[lessThanOrEqual]"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsLastPage),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 1, NextPage: "", Done: true},
		},
		{
			Name: "Until alone is sent as an inclusive updatedAt bound with milliseconds",
			Input: common.ReadParams{
				ObjectName: "opportunities",
				Fields:     connectors.Fields("id"),
				Until:      time.Date(2026, 10, 12, 22, 7, 48, 174_000_000, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/opportunities"),
					mockcond.QueryParam("$updatedAt[lessThanOrEqual]", "2026-10-12T22:07:48.174Z"),
					mockcond.QueryParamsMissing("$updatedAt[greaterThan]"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsLastPage),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 1, NextPage: "", Done: true},
		},
		{
			Name:  "Record properties win over custom fields of the same name, and fields over relationships",
			Input: common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("updatedAt", "partners", "$name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/accounts"),
				Then:  mockserver.Response(http.StatusOK, responseAccountsKeyClash),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "acc_9",
					Fields: map[string]any{
						"updatedat": "2026-09-25T16:45:00.000Z",
						"partners":  "field value",
						"$name":     "Hooli",
					},
					// The shadowed custom field and relationship stay in Raw.
					Raw: map[string]any{
						"fields": map[string]any{
							"$name":     map[string]any{"value": "Hooli", "valueType": "TEXT"},
							"updatedAt": map[string]any{"value": "custom value", "valueType": "TEXT"},
							"partners":  map[string]any{"value": "field value", "valueType": "TEXT"},
						},
						"relationships": map[string]any{
							"partners": map[string]any{
								"cardinality": "HAS_MANY",
								"objectType":  "account",
								"values":      []any{"acc_1"},
							},
						},
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Incremental read of members sends no filters and filters records locally",
			Input: common.ReadParams{
				ObjectName: "members",
				Fields:     connectors.Fields("$email"),
				Since:      time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/members"),
					mockcond.QueryParam("offset", "0"),
					mockcond.QueryParamsMissing("$updatedAt[greaterThan]", "$updatedAt[lessThanOrEqual]"),
				},
				Then: mockserver.Response(http.StatusOK, responseMembers),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				// mem_1 was last updated on 2026-02-01, before Since.
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Member updated exactly at Since is excluded",
			Input: common.ReadParams{
				ObjectName: "members",
				Fields:     connectors.Fields("$email"),
				Since:      time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/members"),
				Then:  mockserver.Response(http.StatusOK, responseMembers),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, NextPage: "", Done: true},
		},
		{
			Name: "Member updated exactly at Until is included",
			Input: common.ReadParams{
				ObjectName: "members",
				Fields:     connectors.Fields("$email"),
				Until:      time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/members"),
					mockcond.QueryParamsMissing("$updatedAt[greaterThan]", "$updatedAt[lessThanOrEqual]"),
				},
				Then: mockserver.Response(http.StatusOK, responseMembers),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id:     "mem_1",
					Fields: map[string]any{"$email": "jane@example.com"},
					Raw:    map[string]any{"id": "mem_1"},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Members are flattened and page size is capped at 25",
			Input: common.ReadParams{
				ObjectName: "members",
				Fields:     connectors.Fields("$email", "$name", "$role", "$user"),
				PageSize:   100,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/members"),
					mockcond.QueryParam("limit", "25"),
					mockcond.QueryParam("offset", "0"),
					mockcond.Header(versionHeader),
				},
				Then: mockserver.Response(http.StatusOK, responseMembers),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "mem_1",
					Fields: map[string]any{
						"$email": "jane@example.com",
						"$name": map[string]any{
							"firstName": "Jane",
							"lastName":  "Doe",
						},
						"$role": "admin",
						"$user": []any{"usr_1"},
					},
					// Raw keeps the provider's nested shape.
					Raw: map[string]any{
						"fields": map[string]any{
							"$email": map[string]any{"value": "jane@example.com", "valueType": "EMAIL"},
							"$name": map[string]any{
								"value":     map[string]any{"firstName": "Jane", "lastName": "Doe"},
								"valueType": "FULL_NAME",
							},
							"$profileImage": map[string]any{"value": nil, "valueType": "URL"},
							"$role":         map[string]any{"value": "admin", "valueType": "TEXT"},
						},
					},
				}},
				NextPage: "",
				Done:     true,
			},
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
