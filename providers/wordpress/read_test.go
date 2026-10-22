package wordpress

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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorInvalidParams := testutils.DataFromFile(t, "read/invalid-params.json")
	responsePostsFirstPage := testutils.DataFromFile(t, "read/posts/1-first-page.json")
	responsePostsLastPage := testutils.DataFromFile(t, "read/posts/2-last-page.json")

	tests := []testconn.TestCaseRead{
		{
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "posts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorInvalidParams),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("rest_invalid_param: Invalid parameter(s): page"),
			},
		},
		{
			Name: "Read parties first page incrementally",
			Input: common.ReadParams{
				ObjectName: "posts",
				Fields:     connectors.Fields("date_gmt", "$['title']['rendered']"),
				Since: time.Date(2024, 9, 19, 4, 30, 45, 600,
					time.FixedZone("UTC-8", -8*60*60)),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/wp/v2/sites/test-workspace/posts"),
					mockcond.QueryParam("modified_after", "2024-09-19T12:30:45Z"),
					mockcond.QueryParam("page", "1"),
					mockcond.QueryParam("per_page", "100"),
				},
				Then: mockserver.ResponseChainedFuncs(
					mockserver.Header("Link",
						`<https://public-api.wordpress.com/wp/v2/sites/258069112/posts?page=2&per_page=1>; rel="next"`),
					mockserver.Response(http.StatusOK, responsePostsFirstPage),
				),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"date_gmt": "2026-10-19T19:31:46",
						"title": map[string]any{
							"rendered": "This is a third&nbsp;Post",
						},
					},
					Raw: map[string]any{"slug": "this-is-a-third-post"},
					Id:  "12",
				}},
				NextPage: "https://public-api.wordpress.com/wp/v2/sites/258069112/posts?page=2&per_page=1",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read parties second page which is without next page token",
			Input: common.ReadParams{
				ObjectName: "posts",
				Fields:     connectors.Fields("date_gmt", "$['title']['rendered']"),
				NextPage:   testconn.URLTestServer + "/wp/v2/sites/258069112/posts?page=2&per_page=1",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/wp/v2/sites/258069112/posts"),
					mockcond.QueryParam("page", "2"),
					mockcond.QueryParam("per_page", "1"),
				},
				Then: mockserver.ResponseChainedFuncs(
					mockserver.Header("Link",
						`<https://public-api.wordpress.com/wp/v2/sites/258069112/posts?page=1&per_page=2>; rel="prev"`),
					mockserver.Response(http.StatusOK, responsePostsLastPage),
				),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"date_gmt": "2026-10-19T18:26:26",
						"title": map[string]any{
							"rendered": "Hello World!",
						},
					},
					Raw: map[string]any{"slug": "hello-world"},
					Id:  "3",
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
