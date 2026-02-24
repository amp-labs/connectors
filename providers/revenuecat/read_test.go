package revenuecat

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

func TestRead(t *testing.T) {
	t.Parallel()

	firstPageWithRelativeNext := []byte(`{
	  "object":"list",
	  "items":[{"id":"prod_1","object":"product"},{"id":"prod_2","object":"product"}],
	  "next_page":"/v2/projects/proj_123/products?starting_after=prod_2",
	  "url":"/v2/projects/proj_123/products"
	}`)

	lastPage := []byte(`{
	  "object":"list",
	  "items":[{"id":"prod_3","object":"product"}],
	  "url":"/v2/projects/proj_123/products"
	}`)

	pageWithCreatedAtAndNext := []byte(`{
	  "object":"list",
	  "items":[
	    {"id":"prod_new","object":"product","created_at":2000},
	    {"id":"prod_old","object":"product","created_at":1000}
	  ],
	  "next_page":"/v2/projects/proj_123/products?starting_after=prod_old",
	  "url":"/v2/projects/proj_123/products"
	}`)

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "products"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Read products uses default limit",
			Input: common.ReadParams{ObjectName: "products", Fields: datautils.NewStringSet("id")},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v2/projects/proj_123/products"),
							mockcond.QueryParam("limit", defaultPageSize),
						},
						Then: mockserver.Response(http.StatusOK, lastPage),
					},
				},
				Default: mockserver.ResponseString(500, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "prod_3"},
					Raw:    map[string]any{"id": "prod_3"},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read products uses PageSize as limit",
			Input: common.ReadParams{ObjectName: "products", Fields: datautils.NewStringSet("id"), PageSize: 7},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v2/projects/proj_123/products"),
							mockcond.QueryParam("limit", "7"),
						},
						Then: mockserver.Response(http.StatusOK, lastPage),
					},
				},
				Default: mockserver.ResponseString(500, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read products returns absolute next page URL for relative next_page",
			Input: common.ReadParams{ObjectName: "products", Fields: datautils.NewStringSet("id")},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v2/projects/proj_123/products"),
							mockcond.QueryParam("limit", defaultPageSize),
						},
						Then: mockserver.Response(http.StatusOK, firstPageWithRelativeNext),
					},
				},
				Default: mockserver.ResponseString(500, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     2,
				NextPage: testroutines.URLTestServer + "/v2/projects/proj_123/products?starting_after=prod_2",
				Done:     false,
			},
		},
		{
			Name: "Read products supports connector-side incremental filtering (created_at) and can stop early",
			Input: common.ReadParams{
				ObjectName: "products",
				Fields:     datautils.NewStringSet("id"),
				Since:      time.UnixMilli(1500),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v2/projects/proj_123/products"),
							mockcond.QueryParam("limit", defaultPageSize),
						},
						Then: mockserver.Response(http.StatusOK, pageWithCreatedAtAndNext),
					},
				},
				Default: mockserver.ResponseString(500, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "prod_new"},
					Raw:    map[string]any{"id": "prod_new"},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read products returns absolute next page token",
			Input: common.ReadParams{ObjectName: "products", Fields: datautils.NewStringSet("id")},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v2/projects/proj_123/products"),
							mockcond.QueryParam("limit", defaultPageSize),
						},
						Then: func(w http.ResponseWriter, r *http.Request) {
							nextPage := "http://" + r.Host + "/v2/projects/proj_123/products?starting_after=prod_1"
							firstPage := []byte(`{
							  "object":"list",
							  "items":[{"id":"prod_1","object":"product"}],
							  "next_page":"` + nextPage + `",
							  "url":"/v2/projects/proj_123/products"
							}`)
							mockserver.Response(http.StatusOK, firstPage)(w, r)
						},
					},
				},
				Default: mockserver.ResponseString(500, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: testroutines.URLTestServer + "/v2/projects/proj_123/products?starting_after=prod_1",
				Done:     false,
			},
		},
		{
			Name: "Read products accepts absolute NextPage URL",
			Input: common.ReadParams{
				ObjectName: "products",
				Fields:     datautils.NewStringSet("id"),
				NextPage:   testroutines.URLTestServer + "/v2/projects/proj_123/products?starting_after=prod_1",
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v2/projects/proj_123/products"),
							mockcond.QueryParam("starting_after", "prod_1"),
						},
						Then: mockserver.Response(http.StatusOK, lastPage),
					},
				},
				Default: mockserver.ResponseString(500, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "",
				Done:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestReadConnector(tt.Server.URL, "proj_123")
			})
		})
	}
}

func constructTestReadConnector(baseURL, projectID string) (*Connector, error) {
	ctx := context.Background()

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, "Authorization", "Bearer test",
		common.WithHeaderClient(mockutils.NewClient()),
	)
	if err != nil {
		return nil, err
	}

	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Metadata: map[string]string{
			"project_id": projectID,
		},
	})
	if err != nil {
		return nil, err
	}

	conn.SetUnitTestBaseURL(baseURL)
	return conn, nil
}
