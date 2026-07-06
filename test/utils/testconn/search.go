package testconn

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

type (
	searchType = TestCase[common.SearchParams, *common.SearchResult]
	// TestCaseSearch is a test suite useful for testing connectors.SearchConnector interface.
	TestCaseSearch searchType
)

// Run provides a procedure to test connectors.SearchConnector
func (s TestCaseSearch) Run(t *testing.T, builder ConnectorBuilder[TestableSearcher]) {
	t.Helper()
	t.Cleanup(func() {
		searchType(s).Close()
	})

	conn := builder.Build(t, s.Name)
	searchParams := prepareSearchParams(s.Server.URL, searchType(s).PrepareInput())
	output, err := conn.Search(t.Context(), searchParams)
	searchType(s).Validate(t, err, output)
}

// This enables tests where we want to specify NextPage. Since we are dealing with mock-server
// NextPage token may include URLTestServer key.
// Example:
//
//	common.SearchParams{
//		ObjectName: "tags",
//		NextPage:   testconn.URLTestServer + "/v1/tags?limit=100&skip=100",
//	}
func prepareSearchParams(serverURL string, params common.SearchParams) *common.SearchParams {
	params.NextPage = common.NextPageToken(
		ResolveTestServerURL(params.NextPage.String(), serverURL),
	)

	return &params
}
