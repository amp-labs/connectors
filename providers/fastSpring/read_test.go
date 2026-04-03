package fastspring

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

// TestRead duplicates default limit ("1000") and event "days" ("30") as string literals instead of
// importing read.go constants so changing those defaults forces an explicit test update in review.

func TestRead(t *testing.T) { // nolint:funlen
	t.Parallel()

	firstPage := []byte(`{
		"accounts": [
			{"id": "acc_1", "account": "Account One"},
			{"id": "acc_2", "account": "Account Two"}
		],
		"nextPage": 2
			}`)

	lastPage := []byte(`{
			"accounts": [
				{"id": "acc_3", "account": "Account Three"}
			],
			"nextPage": 0
	}`)

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "At least one field is requested",
			Input: common.ReadParams{
				ObjectName: "accounts",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "Read accounts uses default pagination and maps nextPage to page query",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     connectors.Fields("id", "account"),
				PageSize:   0, // unset → connector uses default limit (1000 in read.go at time of writing)
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/accounts"),
							mockcond.QueryParam("limit", "1000"),
							mockcond.QueryParam("page", "1"),
						},
						Then: mockserver.Response(http.StatusOK, firstPage),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     2,
				NextPage: testroutines.URLTestServer + "/accounts?limit=1000&page=2",
				Done:     false,
			},
		},
		{
			Name: "Read accounts accepts NextPage token and returns last page",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     datautils.NewStringSet("id"),
				NextPage:   testroutines.URLTestServer + "/accounts?limit=1000&page=2",
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/accounts"),
							mockcond.QueryParam("limit", "1000"),
							mockcond.QueryParam("page", "2"),
						},
						Then: mockserver.Response(http.StatusOK, lastPage),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "acc_3"},
					Raw: map[string]any{
						"id":      "acc_3",
						"account": "Account Three",
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Read processed events maps Since and Until to begin and end query params",
			Input: common.ReadParams{
				ObjectName: "events-processed",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/events/processed"),
							mockcond.QueryParam("days", "30"),
							mockcond.QueryParam("begin", "2025-01-01"),
							mockcond.QueryParam("end", "2025-01-31"),
							mockcond.QueryParam("limit", "1000"),
							mockcond.QueryParam("page", "1"),
						},
						Then: mockserver.Response(http.StatusOK, []byte(`{"events":[],"nextPage":0}`)),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				NextPage: "",
				Done:     true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestReadConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestReadConnector(baseURL string) (*Connector, error) {
	ctx := context.Background()

	client, err := common.NewBasicAuthHTTPClient(ctx, "test-user", "test-password",
		common.WithHeaderClient(mockutils.NewClient()),
	)
	if err != nil {
		return nil, err
	}

	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		return nil, err
	}

	conn.SetUnitTestBaseURL(baseURL)

	return conn, nil
}
