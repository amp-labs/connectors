package pipeliner

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseNotFound := testutils.DataFromFile(t, "resource-not-found.json")
	responseProfilesFirstPage := testutils.DataFromFile(t, "read/profiles/1-first-page.json")
	responseProfilesSecondPage := testutils.DataFromFile(t, "read/profiles/2-second-page.json")
	responseProfilesLastPage := testutils.DataFromFile(t, "read/profiles/3-last-page.json")
	responseAccounts := testutils.DataFromFile(t, "read/accounts-with-associations.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "Profiles"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "Profiles", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("The requested URL was not found on the server"),
			},
		},
		{
			Name:  "Incorrect key in payload",
			Input: common.ReadParams{ObjectName: "Profiles", Fields: connectors.Fields("id")},
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
			Input: common.ReadParams{ObjectName: "Profiles", Fields: connectors.Fields("id")},
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
			Input: common.ReadParams{ObjectName: "Profiles", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `
				{
				  "success": true,
				  "total": 0,
				  "data": []
				}`),
			}.Server(),
			Expected:     &common.ReadResult{Done: true, Data: []common.ReadResultRow{}},
			ExpectedErrs: nil,
		},
		{
			Name: "Next page URL is inferred, when provided with an object",
			Input: common.ReadParams{
				ObjectName: "Profiles",
				Fields:     connectors.Fields("id"),
				Since: time.Date(2024, 9, 19, 4, 30, 45, 600,
					time.FixedZone("UTC-8", -8*60*60)),
				PageSize: 77,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v100/rest/spaces/test-workspace/entities/Profiles"),
					mockcond.QueryParam("first", "77"),
					mockcond.QueryParam("order-by", "-modified"),
					mockcond.QueryParam("filter-op[modified]", "gte"),
					mockcond.QueryParam("filter[modified]", "2024-09-19T12:30:45Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseProfilesFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     2,
				NextPage: "WyIwMDAwMDAwMC0wMDAwLTAwMDEtMDAwMS0wMDAwMDAwMDhlOTciXQ==",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Next page URL is empty, when provided with null object",
			Input: common.ReadParams{ObjectName: "Profiles", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseProfilesLastPage),
			}.Server(),
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 0, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read with chosen fields",
			Input: common.ReadParams{
				ObjectName: "Profiles",
				Fields:     connectors.Fields("name", "owner_id"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v100/rest/spaces/test-workspace/entities/Profiles"),
				Then:  mockserver.Response(http.StatusOK, responseProfilesSecondPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name":     "Lang_DefaultProfileAllUsers",
						"owner_id": "00000000-0000-0000-0000-000000008e97",
					},
					Raw: map[string]any{
						"name":       "Lang_DefaultProfileAllUsers",
						"owner_id":   "00000000-0000-0000-0000-000000008e97",
						"use_lang":   true,
						"entity":     float64(3),
						"is_deleted": false,
					},
				}, {
					Fields: map[string]any{
						"name":     "Lang_DefaultProfileMy",
						"owner_id": "00000000-0000-0000-0000-000000008e97",
					},
					Raw: map[string]any{
						"name":       "Lang_DefaultProfileMy",
						"owner_id":   "00000000-0000-0000-0000-000000008e97",
						"use_lang":   true,
						"entity":     float64(3),
						"is_deleted": false,
					},
				}},
				NextPage: "WyIwMDAwMDAwMC0wMDAwLTAwMDMtMDAwMS0wMDAwMDAwMDhlOTciXQ==",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Accounts with associations",
			Input: common.ReadParams{
				ObjectName:        "Accounts",
				Fields:            connectors.Fields("id"),
				AssociatedObjects: []string{"industry", "customer_type"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v100/rest/spaces/test-workspace/entities/Accounts"),
					mockcond.QueryParam("first", "100"),
					mockcond.QueryParam("expand", "industry,customer_type"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccounts),
			}.Server(),
			Comparator: compareSubsetAssociations(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "0f4daaff-ffdd-4fff-9109-843e45e9609c",
					},
					Raw: map[string]any{
						"is_deleted": false,
						"modified":   "2026-02-05 00:13:56.718423+00:00",
						"created":    "2026-02-05 00:13:56.718423+00:00",
						"customer_type": map[string]any{
							"id":          "04444b3a-c669-03bc-2c49-bcd7f047d41a",
							"modified":    "2018-09-03 17:18:26.215642+00:00",
							"created":     "2018-09-03 17:18:26.215642+00:00",
							"data_set_id": "3c439b9a-f92a-0dd6-9ff9-74b90f6ae536",
							"option_name": "existing customer",
						},
						"industry": map[string]any{
							"id":          "3e046e52-00e4-097b-b2a4-b58af0381541",
							"modified":    "2018-09-03 17:18:26.215642+00:00",
							"created":     "2018-09-03 17:18:26.215642+00:00",
							"data_set_id": "d15960b8-c986-0a18-b587-67f0a78f827b",
							"option_name": "Agriculture",
						},
					},
					Associations: map[string][]common.Association{
						"customer_type": {{
							ObjectId:        "04444b3a-c669-03bc-2c49-bcd7f047d41a",
							AssociationType: "customer_type",
							Raw: map[string]any{
								"id":          "04444b3a-c669-03bc-2c49-bcd7f047d41a",
								"modified":    "2018-09-03 17:18:26.215642+00:00",
								"created":     "2018-09-03 17:18:26.215642+00:00",
								"data_set_id": "3c439b9a-f92a-0dd6-9ff9-74b90f6ae536",
								"option_name": "existing customer",
							},
						}},
						"industry": {{
							ObjectId:        "3e046e52-00e4-097b-b2a4-b58af0381541",
							AssociationType: "industry",
							Raw: map[string]any{
								"id":          "3e046e52-00e4-097b-b2a4-b58af0381541",
								"modified":    "2018-09-03 17:18:26.215642+00:00",
								"created":     "2018-09-03 17:18:26.215642+00:00",
								"data_set_id": "d15960b8-c986-0a18-b587-67f0a78f827b",
								"option_name": "Agriculture",
							},
						}},
					},
				}},
				NextPage: "WyIwZjRkYWFmZi1mZmRkLTRmZmYtOTEwOS04NDNlNDVlOTYwOWMiXQ==",
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

func compareSubsetAssociations() testroutines.Comparator[*common.ReadResult] {
	return func(serverURL string, actual *common.ReadResult, expected *common.ReadResult) bool {
		// Associations.
		for index, datum := range expected.Data {
			if !reflect.DeepEqual(datum.Associations, expected.Data[index].Associations) {
				return false
			}
		}

		// Usual subset comparison.
		return testroutines.ComparatorSubsetRead(serverURL, actual, expected)
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: mockutils.NewClient(),
			Workspace:           "test-workspace",
			Metadata:            map[string]string{},
		},
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetUnitTestBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
