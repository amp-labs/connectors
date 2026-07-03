package testconn

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

type (
	metadataType = TestCase[[]string, *common.ListObjectMetadataResult]
	// TestCaseListObjectMetadata is a test suite useful for testing connectors.ObjectMetadataConnector interface.
	TestCaseListObjectMetadata metadataType
)

// Run provides a procedure to test connectors.ObjectMetadataConnector
func (m TestCaseListObjectMetadata) Run(t *testing.T, builder ConnectorBuilder[TestableMetadataReader]) {
	t.Helper()
	t.Cleanup(func() {
		metadataType(m).Close()
	})

	conn := builder.Build(t, m.Name)
	output, err := conn.ListObjectMetadata(t.Context(), metadataType(m).PrepareInput())
	metadataType(m).Validate(t, err, output)
}
