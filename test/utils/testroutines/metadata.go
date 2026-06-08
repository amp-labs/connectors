package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

type (
	metadataType = TestCase[[]string, *common.ListObjectMetadataResult]
	// Metadata is a test suite useful for testing connectors.ObjectMetadataConnector interface.
	Metadata metadataType
)

// Run provides a procedure to test connectors.ObjectMetadataConnector
func (m Metadata) Run(t *testing.T, builder ConnectorBuilder[connectors.ObjectMetadataConnector]) {
	t.Helper()
	t.Cleanup(func() {
		metadataType(m).Close()
	})

	conn := builder.Build(t, m.Name)
	output, err := conn.ListObjectMetadata(t.Context(), metadataType(m).PrepareInput())
	metadataType(m).Validate(t, err, output)
}
