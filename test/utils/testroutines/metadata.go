package testroutines

import (
	"context"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

type (
	MetadataType = TestCase[[]string, *common.ListObjectMetadataResult]
	// Metadata is a test suite useful for testing connectors.ObjectMetadataConnector interface.
	Metadata MetadataType
)

// Run provides a procedure to test connectors.ObjectMetadataConnector
func (m Metadata) Run(t *testing.T, builder ConnectorBuilder[connectors.ObjectMetadataConnector]) {
	t.Helper()
	conn := builder.Build(t, m.Name)
	output, err := conn.ListObjectMetadata(context.Background(), m.Input)
	MetadataType(m).Validate(t, err, output)
}
