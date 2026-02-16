package revenuecat

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
)

func TestRead_Products_PaginatesUsingNextPage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	firstPage := []byte(`{
	  "object":"list",
	  "items":[{"id":"prod_1","object":"product"},{"id":"prod_2","object":"product"}],
	  "next_page":"/v2/projects/proj_123/products?starting_after=prod_2",
	  "url":"/v2/projects/proj_123/products"
	}`)

	secondPage := []byte(`{
	  "object":"list",
	  "items":[{"id":"prod_3","object":"product"}],
	  "url":"/v2/projects/proj_123/products"
	}`)

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v2/projects/proj_123/products"),
					mockcond.QueryParam("limit", defaultPageSize),
				},
				Then: mockserver.Response(200, firstPage),
			},
			{
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v2/projects/proj_123/products"),
					mockcond.QueryParam("starting_after", "prod_2"),
				},
				Then: mockserver.Response(200, secondPage),
			},
		},
		Default: mockserver.ResponseString(500, `{"error":"unexpected request"}`),
	}.Server()
	t.Cleanup(server.Close)

	conn := mustTestConnector(t, server.URL, "proj_123")

	out1, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "products",
		Fields:     datautils.NewStringSet("id"),
	})
	if err != nil {
		t.Fatalf("read first page error: %v", err)
	}
	if out1.NextPage.String() == "" {
		t.Fatalf("expected next page token to be set")
	}

	out2, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "products",
		Fields:     datautils.NewStringSet("id"),
		NextPage:   out1.NextPage,
	})
	if err != nil {
		t.Fatalf("read second page error: %v", err)
	}
	if out2.NextPage.String() != "" {
		t.Fatalf("expected no next page token on last page, got: %s", out2.NextPage.String())
	}
}

func TestRead_Products_UsesPageSizeAsLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	body := []byte(`{
	  "object":"list",
	  "items":[{"id":"prod_1","object":"product"}],
	  "url":"/v2/projects/proj_123/products"
	}`)

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v2/projects/proj_123/products"),
					mockcond.QueryParam("limit", "7"),
				},
				Then: mockserver.Response(200, body),
			},
		},
		Default: mockserver.ResponseString(500, `{"error":"unexpected request"}`),
	}.Server()
	t.Cleanup(server.Close)

	conn := mustTestConnector(t, server.URL, "proj_123")

	_, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "products",
		Fields:     datautils.NewStringSet("id"),
		PageSize:   7,
	})
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
}

func TestRead_Products_AllowsAbsoluteNextPageURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	secondPage := []byte(`{
	  "object":"list",
	  "items":[{"id":"prod_2","object":"product"}],
	  "url":"/v2/projects/proj_123/products"
	}`)

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v2/projects/proj_123/products"),
					mockcond.QueryParam("limit", defaultPageSize),
				},
				Then: func(w http.ResponseWriter, r *http.Request) {
					// Provide an absolute next_page URL.
					nextPage := fmt.Sprintf("http://%s/v2/projects/proj_123/products?starting_after=prod_1", r.Host)
					firstPage := []byte(`{
					  "object":"list",
					  "items":[{"id":"prod_1","object":"product"}],
					  "next_page":"` + nextPage + `",
					  "url":"/v2/projects/proj_123/products"
					}`)
					mockserver.Response(200, firstPage)(w, r)
				},
			},
			{
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v2/projects/proj_123/products"),
					mockcond.QueryParam("starting_after", "prod_1"),
				},
				Then: mockserver.Response(200, secondPage),
			},
		},
		Default: mockserver.ResponseString(500, `{"error":"unexpected request"}`),
	}.Server()
	t.Cleanup(server.Close)

	conn := mustTestConnector(t, server.URL, "proj_123")

	out1, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "products",
		Fields:     datautils.NewStringSet("id"),
	})
	if err != nil {
		t.Fatalf("read first page error: %v", err)
	}
	if out1.NextPage.String() == "" {
		t.Fatalf("expected next page token to be set")
	}

	_, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "products",
		Fields:     datautils.NewStringSet("id"),
		NextPage:   out1.NextPage,
	})
	if err != nil {
		t.Fatalf("read second page error: %v", err)
	}
}

func mustTestConnector(t *testing.T, baseURL, projectID string) *Connector {
	t.Helper()

	ctx := context.Background()

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, "Authorization", "Bearer test")
	if err != nil {
		t.Fatalf("auth client error: %v", err)
	}

	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Metadata: map[string]string{
			"project_id": projectID,
		},
	})
	if err != nil {
		t.Fatalf("connector init error: %v", err)
	}

	conn.SetUnitTestBaseURL(baseURL)

	return conn
}
