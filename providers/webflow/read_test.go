package webflow

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

const testSitePath = "/v2/sites/test-site-id"

func TestRead(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseSites := testutils.DataFromFile(t, "read/sites.json")
	responsePagesPage1 := testutils.DataFromFile(t, "read/pages-page1.json")
	responsePagesPage2 := testutils.DataFromFile(t, "read/pages-page2.json")
	responsePagesEmpty := testutils.DataFromFile(t, "read/pages-empty.json")
	responseAssets := testutils.DataFromFile(t, "read/assets.json")
	responseCommentsPage1 := testutils.DataFromFile(t, "read/comments-page1.json")
	responseProducts := testutils.DataFromFile(t, "read/products.json")
	responseOrders := testutils.DataFromFile(t, "read/orders.json")
	responseForbidden := testutils.DataFromFile(t, "read/error-forbidden.json")

	since := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: objectPages},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "items", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Sites are account scoped and not paginated",
			Input: common.ReadParams{ObjectName: objectSites, Fields: connectors.Fields("id", "displayName")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v2/sites"),
					mockcond.QueryParamsMissing(limitParam, offsetParam),
				},
				Then: mockserver.Response(http.StatusOK, responseSites),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":          "6ac4a4402b7737869ac6f61d",
						"displayname": "Joseph's Five-Star Site",
					},
					Raw: map[string]any{
						"id":          "6ac4a4402b7737869ac6f61d",
						"displayName": "Joseph's Five-Star Site",
						"timeZone":    "America/Los_Angeles",
					},
					Id: "6ac4a4402b7737869ac6f61d",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: objectPages, Fields: connectors.Fields("id", "title")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path(testSitePath + "/pages"),
					mockcond.QueryParam(limitParam, "100"),
					mockcond.QueryParam(offsetParam, "0"),
				},
				Then: mockserver.Response(http.StatusOK, responsePagesEmpty),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "First page of pages advances offset while below total",
			Input: common.ReadParams{
				ObjectName: objectPages,
				Fields:     connectors.Fields("id", "title", "archived"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path(testSitePath + "/pages"),
					mockcond.QueryParam(limitParam, "1"),
					mockcond.QueryParam(offsetParam, "0"),
				},
				Then: mockserver.Response(http.StatusOK, responsePagesPage1),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":       "page-1",
						"title":    "Home",
						"archived": false,
					},
					Raw: map[string]any{
						"id":   "page-1",
						"slug": "index",
					},
					Id: "page-1",
				}},
				NextPage: testconn.URLTestServer + testSitePath + "/pages?limit=1&offset=1",
				Done:     false,
			},
		},
		{
			Name: "Last page of pages via NextPage",
			Input: common.ReadParams{
				ObjectName: objectPages,
				Fields:     connectors.Fields("id", "title"),
				NextPage:   testconn.URLTestServer + testSitePath + "/pages?limit=1&offset=1",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path(testSitePath + "/pages"),
					mockcond.QueryParam(limitParam, "1"),
					mockcond.QueryParam(offsetParam, "1"),
				},
				Then: mockserver.Response(http.StatusOK, responsePagesPage2),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "page-2", "title": "About"},
					Raw:    map[string]any{"id": "page-2", "draft": true},
					Id:     "page-2",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Since is ignored for lists that cannot be sorted",
			Input: common.ReadParams{
				ObjectName: objectAssets,
				Fields:     connectors.Fields("id", "displayName"),
				Since:      since,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path(testSitePath + "/assets"),
					mockcond.QueryParam(limitParam, "100"),
					mockcond.QueryParam(offsetParam, "0"),
					mockcond.QueryParamsMissing(sortByParam, sortOrderParam),
				},
				Then: mockserver.Response(http.StatusOK, responseAssets),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"id": "asset-new", "displayname": "fans.jpg"},
						Raw:    map[string]any{"id": "asset-new", "contentType": "image/jpeg"},
						Id:     "asset-new",
					},
					{
						Fields: map[string]any{"id": "asset-old", "displayname": "logo.svg"},
						Raw:    map[string]any{"id": "asset-old", "contentType": "image/svg+xml"},
						Id:     "asset-old",
					},
				},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Comments are sorted newest first and stop paging at Since",
			Input: common.ReadParams{
				ObjectName: objectComments,
				Fields:     connectors.Fields("id", "content"),
				PageSize:   2,
				Since:      since,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path(testSitePath + "/comments"),
					mockcond.QueryParam(limitParam, "2"),
					mockcond.QueryParam(offsetParam, "0"),
					mockcond.QueryParam(sortByParam, fieldLastUpdated),
					mockcond.QueryParam(sortOrderParam, sortDescending),
				},
				Then: mockserver.Response(http.StatusOK, responseCommentsPage1),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "comment-new", "content": "Please update the hero image"},
					Raw:    map[string]any{"id": "comment-new", "isResolved": false},
					Id:     "comment-new",
				}},
				// 50 comments exist, but the second record is already older than Since.
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Comments without Since keep paging",
			Input: common.ReadParams{
				ObjectName: objectComments,
				Fields:     connectors.Fields("id"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path(testSitePath + "/comments"),
					mockcond.QueryParam(limitParam, "2"),
					mockcond.QueryParam(offsetParam, "0"),
				},
				Then: mockserver.Response(http.StatusOK, responseCommentsPage1),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{Fields: map[string]any{"id": "comment-new"}, Raw: map[string]any{"id": "comment-new"}, Id: "comment-new"},
					{Fields: map[string]any{"id": "comment-old"}, Raw: map[string]any{"id": "comment-old"}, Id: "comment-old"},
				},
				NextPage: testconn.URLTestServer + testSitePath +
					"/comments?limit=2&offset=2&sortBy=lastUpdated&sortOrder=desc",
				Done: false,
			},
		},
		{
			Name: "Products flatten product fields next to skus and keep the raw envelope",
			Input: common.ReadParams{
				ObjectName: objectProducts,
				Fields:     connectors.Fields("id", "fieldData", "lastUpdated", "skus"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path(testSitePath + "/products"),
				},
				Then: mockserver.Response(http.StatusOK, responseProducts),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":          "product-1",
						"fielddata":   map[string]any{"name": "Five-Star Hat", "slug": "five-star-hat"},
						"lastupdated": "2026-10-06T08:00:00.000Z",
						"skus": []any{map[string]any{
							"id":        "sku-1",
							"fieldData": map[string]any{"name": "Five-Star Hat", "price": map[string]any{"value": float64(2500), "unit": "USD"}},
						}},
					},
					Raw: map[string]any{
						"product": map[string]any{
							"id":          "product-1",
							"cmsLocaleId": "653ad57de882f528b32e810e",
							"isArchived":  false,
							"isDraft":     false,
							"createdOn":   "2026-09-10T08:00:00.000Z",
							"lastUpdated": "2026-10-06T08:00:00.000Z",
							"fieldData":   map[string]any{"name": "Five-Star Hat", "slug": "five-star-hat"},
						},
					},
					Id: "product-1",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Orders use orderId as the record id",
			Input: common.ReadParams{ObjectName: objectOrders, Fields: connectors.Fields("orderId", "status")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path(testSitePath + "/orders"),
				},
				Then: mockserver.Response(http.StatusOK, responseOrders),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"orderid": "a1b2-c3d4", "status": "unfulfilled"},
					Raw:    map[string]any{"orderId": "a1b2-c3d4"},
					Id:     "a1b2-c3d4",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Plan gated endpoint surfaces forbidden error",
			Input: common.ReadParams{ObjectName: objectRedirects, Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusForbidden, responseForbidden),
			}.Server(),
			ExpectedErrs: []error{common.ErrForbidden},
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
