package aircall

import (
	"errors"
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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseUsers := testutils.DataFromFile(t, "read/users/response.json")
	responseUsersPaginated := testutils.DataFromFile(t, "read/users/paginated-response.json")
	responseCallsEmpty := testutils.DataFromFile(t, "read/calls/empty-response.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "users"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "Read users successfully",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "name", "email"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/users"),
					mockcond.QueryParam("per_page", "50"),
				},
				Then: mockserver.Response(http.StatusOK, responseUsers),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    float64(1784786),
						"name":  "Constantin Koval",
						"email": "constantin@withampersand.com",
					},
					Raw: map[string]any{
						"id":                  float64(1784786),
						"direct_link":         "https://api.aircall.io/v1/users/1784786",
						"name":                "Constantin Koval",
						"email":               "constantin@withampersand.com",
						"available":           false,
						"availability_status": "available",
						"created_at":          "2025-11-21T16:58:07Z",
						"language":            "en-US",
						"time_zone":           "Etc/UTC",
						"wrap_up_time":        float64(0),
						"state":               "always_opened",
						"extension":           "001",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read users with pagination",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/users"),
					mockcond.QueryParam("per_page", "50"),
				},
				Then: mockserver.Response(http.StatusOK, responseUsersPaginated),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(1784786),
						"name": "Constantin Koval",
					},
					Raw: map[string]any{
						"id":   float64(1784786),
						"name": "Constantin Koval",
					},
				}, {
					Fields: map[string]any{
						"id":   float64(1784787),
						"name": "John Doe",
					},
					Raw: map[string]any{
						"id":   float64(1784787),
						"name": "John Doe",
					},
				}},
				NextPage: "https://api.aircall.io/v1/users?per_page=2&page=2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read teams successfully",
			Input: common.ReadParams{
				ObjectName: "teams",
				Fields:     connectors.Fields("id", "name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/teams"),
					mockcond.QueryParam("per_page", "50"),
				},
				Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "read/teams/response.json")),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(123),
						"name": "Support",
					},
					Raw: map[string]any{
						"id":   float64(123),
						"name": "Support",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read tags successfully",
			Input: common.ReadParams{
				ObjectName: "tags",
				Fields:     connectors.Fields("id", "name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/tags"),
					mockcond.QueryParam("per_page", "50"),
				},
				Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "read/tags/response.json")),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(10),
						"name": "VIP",
					},
					Raw: map[string]any{
						"id":   float64(10),
						"name": "VIP",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read contacts successfully",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "first_name", "last_name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/contacts"),
					mockcond.QueryParam("per_page", "50"),
				},
				Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "read/contacts/response.json")),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":         float64(2410757),
						"first_name": "John",
						"last_name":  "Doe",
					},
					Raw: map[string]any{
						"id":         float64(2410757),
						"first_name": "John",
						"last_name":  "Doe",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read numbers successfully",
			Input: common.ReadParams{
				ObjectName: "numbers",
				Fields:     connectors.Fields("id", "name", "digits"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/numbers"),
					mockcond.QueryParam("per_page", "50"),
				},
				Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "read/numbers/response.json")),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":     float64(456),
						"name":   "Main Line",
						"digits": "+15551234567",
					},
					Raw: map[string]any{
						"id":     float64(456),
						"name":   "Main Line",
						"digits": "+15551234567",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read calls returns empty result",
			Input: common.ReadParams{
				ObjectName: "calls",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/calls"),
					mockcond.QueryParam("per_page", "50"),
				},
				Then: mockserver.Response(http.StatusOK, responseCallsEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Unauthorized error",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnauthorized, []byte(`{"error": "Unauthorized"}`)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrAccessToken,
				errors.New("Unauthorized"), //nolint:goerr113
			},
		},
		{
			Name: "Not found error",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, []byte(`{"error": "Not Found"}`)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrNotFound,
			},
		},
		{
			Name: "Bad request error",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, []byte(`{"error": "Invalid parameter"}`)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrCaller,
				errors.New("Invalid parameter"), //nolint:goerr113
			},
		},
		{
			Name: "Internal server error",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusInternalServerError, []byte(`{"error": "Internal Server Error"}`)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrServer,
				errors.New("Internal Server Error"), //nolint:goerr113
			},
		},
		{
			Name: "Incremental sync with date range",
			Input: common.ReadParams{
				ObjectName: "calls",
				Fields:     connectors.Fields("id", "created_at"),
				Since:      time.Unix(1700000000, 0), // 2023-11-14
				Until:      time.Unix(1700086400, 0), // 2023-11-15
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/calls"),
					mockcond.QueryParam("per_page", "50"),
					mockcond.QueryParam("from", "1700000000"),
					mockcond.QueryParam("to", "1700086400"),
				},
				Then: mockserver.Response(http.StatusOK, responseCallsEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read with custom PageSize",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "name"),
				PageSize:   10,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/users"),
					mockcond.QueryParam("per_page", "10"),
				},
				Then: mockserver.Response(http.StatusOK, responseUsers),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(1784786),
						"name": "Constantin Koval",
					},
					Raw: map[string]any{
						"id":   float64(1784786),
						"name": "Constantin Koval",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read with PageSize exceeding maximum caps at 50",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id"),
				PageSize:   100,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/users"),
					mockcond.QueryParam("per_page", "50"), // Should be capped at 50
				},
				Then: mockserver.Response(http.StatusOK, responseUsers),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": float64(1784786),
					},
					Raw: map[string]any{
						"id": float64(1784786),
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Date filtering not applied to teams",
			Input: common.ReadParams{
				ObjectName: "teams",
				Fields:     connectors.Fields("id"),
				Since:      time.Unix(1700000000, 0),
				Until:      time.Unix(1700086400, 0),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/teams"),
					mockcond.QueryParam("per_page", "50"),
					// Should NOT have from/to params since teams doesn't support date filtering
				},
				Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "read/teams/response.json")),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": float64(123),
					},
					Raw: map[string]any{
						"id": float64(123),
					},
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

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
