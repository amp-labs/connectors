package confluence

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

	errorInvalidPath := testutils.DataFromFile(t, "confluence/read/err-invalid-path.json")
	errorBadPageSize := testutils.DataFromFile(t, "confluence/read/err-page-size.json")
	responseBlogpostsFirstPage := testutils.DataFromFile(t, "confluence/read/blogposts/1-first-page.json")
	responseBlogpostsLastPage := testutils.DataFromFile(t, "confluence/read/blogposts/2-last-page.json")
	responseBlogpostsManyRecords := testutils.DataFromFile(t, "confluence/read/blogposts/many-records.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "blogposts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Error message page size out of range is parsed",
			Input: common.ReadParams{ObjectName: "blogposts", Fields: connectors.Fields("body")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorBadPageSize),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Provided size {300} for 'limit' is greater than the max allowed: 250"), // nolint:goerr113
			},
		},
		{
			Name:  "Error message invalid path is parsed",
			Input: common.ReadParams{ObjectName: "blogposts", Fields: connectors.Fields("body")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorInvalidPath),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("No message available"), // nolint:goerr113
			},
		},
		{
			Name:  "First page has next page reference",
			Input: common.ReadParams{ObjectName: "blogposts", Fields: connectors.Fields("title")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/ex/confluence/ebc887b2-7e61-4059-ab35-71f15cc16e12/wiki/api/v2/blogposts"),
					mockcond.QueryParam("limit", "250"),
				},
				Then: mockserver.ResponseChainedFuncs(
					mockserver.Header("Link",
						`</wiki/api/v2/blogposts?limit=2&cursor=eyJpZCI6Ijc1MzY2OSIsImNvbnRlbnRPcmRlciI6ImlkIiwiY29udGVudE9yZGVyVmFsdWUiOjc1MzY2OX0=>; rel="next", <https://withampersand-team-oqo0hkaj.atlassian.net/wiki>; rel="base"`), // nolint:lll
					mockserver.Response(http.StatusOK, responseBlogpostsFirstPage),
				),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"title": "First blog post ever!",
					},
					Raw: map[string]any{
						"id":        "688129",
						"spaceId":   "131189",
						"createdAt": "2025-06-26T20:39:16.596Z",
					},
				}, {
					Fields: map[string]any{
						"title": "Second blog post!",
					},
					Raw: map[string]any{
						"id":        "753669",
						"spaceId":   "131189",
						"createdAt": "2025-06-26T20:39:37.501Z",
					},
				}},
				NextPage: testconn.URLTestServer + "/ex/confluence/ebc887b2-7e61-4059-ab35-71f15cc16e12/wiki/api/v2/blogposts?limit=2&cursor=eyJpZCI6Ijc1MzY2OSIsImNvbnRlbnRPcmRlciI6ImlkIiwiY29udGVudE9yZGVyVmFsdWUiOjc1MzY2OX0=", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read with chosen fields using next page token",
			Input: common.ReadParams{
				ObjectName: "blogposts",
				NextPage: testconn.URLTestServer + "/ex/confluence/ebc887b2-7e61-4059-ab35-71f15cc16e12/wiki/api/v2/blogposts/wiki/api/v2/blogposts" + // nolint:lll
					"?limit=2&cursor=eyJpZCI6Ijc1MzY2OSIsImNvbnRlbnRPcmRlciI6ImlkIiwiY29udGVudE9yZGVyVmFsdWUiOjc1MzY2OX0=", // nolint:lll
				Fields: connectors.Fields("title"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/ex/confluence/ebc887b2-7e61-4059-ab35-71f15cc16e12/wiki/api/v2/blogposts/wiki/api/v2/blogposts"), // nolint:lll
					mockcond.QueryParam("limit", "2"),
					mockcond.QueryParam("cursor", "eyJpZCI6Ijc1MzY2OSIsImNvbnRlbnRPcmRlciI6ImlkIiwiY29udGVudE9yZGVyVmFsdWUiOjc1MzY2OX0="), // nolint:lll
				},
				Then: mockserver.ResponseChainedFuncs(
					mockserver.Header("Link",
						`<https://withampersand-team-oqo0hkaj.atlassian.net/wiki>; rel="base"`),
					mockserver.Response(http.StatusOK, responseBlogpostsLastPage),
				),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"title": "Even a third blog post, on a roll!",
					},
					Raw: map[string]any{
						"id":        "819206",
						"spaceId":   "131189",
						"createdAt": "2025-06-26T20:39:48.350Z",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Blogposts with connector side filtering",
			Input: common.ReadParams{
				ObjectName: "blogposts",
				Fields:     connectors.Fields("title"),
				Since:      time.Date(2026, time.February, 25, 5, 17, 45, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/ex/confluence/ebc887b2-7e61-4059-ab35-71f15cc16e12/wiki/api/v2/blogposts"),
					mockcond.QueryParam("limit", "250"),
					mockcond.QueryParam("sort", "-created-date"),
				},
				Then: mockserver.Response(http.StatusOK, responseBlogpostsManyRecords),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"title": "4th post"},
					Raw:    map[string]any{"id": "33390597"},
				}, {
					Fields: map[string]any{"title": "3rd post"},
					Raw:    map[string]any{"id": "33456129"},
				}, {
					Fields: map[string]any{"title": "2nd post"},
					Raw:    map[string]any{"id": "33292310"},
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
				return constructTestAdapter(tt.Server)
			})
		})
	}
}
