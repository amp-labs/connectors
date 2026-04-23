package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

type (
	SearchType = TestCase[common.SearchParams, *common.SearchResult]
	// Search is a test suite useful for testing connectors.SearchConnector interface.
	Search SearchType
)

// Run provides a procedure to test connectors.SearchConnector
func (r Search) Run(t *testing.T, builder ConnectorBuilder[connectors.SearchConnector]) {
	t.Helper()
	t.Cleanup(func() {
		SearchType(r).Close()
	})

	conn := builder.Build(t, r.Name)
	searchParams := prepareSearchParams(r.Server.URL, r.Input)
	output, err := conn.Search(t.Context(), searchParams)
	SearchType(r).Validate(t, err, output)
}

// This enables tests where we want to specify NextPage. Since we are dealing with mock-server
// NextPage token may include URLTestServer key.
// Example:
//
//	common.SearchParams{
//		ObjectName: "tags",
//		NextPage:   testroutines.URLTestServer + "/v1/tags?limit=100&skip=100",
//	}
func prepareSearchParams(serverURL string, params common.SearchParams) *common.SearchParams {
	params.NextPage = common.NextPageToken(
		resolveTestServerURL(params.NextPage.String(), serverURL),
	)

	return &params
}
