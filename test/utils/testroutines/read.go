package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

type (
	ReadType = TestCase[common.ReadParams, *common.ReadResult]
	// Read is a test suite useful for testing connectors.ReadConnector interface.
	Read ReadType
)

// Run provides a procedure to test connectors.ReadConnector
func (r Read) Run(t *testing.T, builder ConnectorBuilder[connectors.ReadConnector]) {
	t.Helper()
	t.Cleanup(func() {
		ReadType(r).Close()
	})

	conn := builder.Build(t, r.Name)
	readParams := prepareReadParams(r.Server.URL, r.Input)
	output, err := conn.Read(t.Context(), readParams)
	ReadType(r).Validate(t, err, output)
}

// This enables tests where we want to specify NextPage. Since we are dealing with mock-server
// NextPage token may include URLTestServer key.
// Example:
//
//	common.ReadParams{
//		ObjectName: "tags",
//		NextPage:   testroutines.URLTestServer + "/v1/tags?limit=100&skip=100",
//	}
func prepareReadParams(serverURL string, config common.ReadParams) common.ReadParams {
	config.NextPage = common.NextPageToken(
		resolveTestServerURL(config.NextPage.String(), serverURL),
	)

	return config
}
