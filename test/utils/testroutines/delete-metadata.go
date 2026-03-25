package testroutines

import (
	"context"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

type (
	DeleteMetadataType = TestCase[*common.DeleteMetadataParams, *common.DeleteMetadataResult]
	// DeleteMetadata is a test suite useful for testing connectors.DeleteMetadataConnector interface.
	DeleteMetadata DeleteMetadataType
)

// Run provides a procedure to test connectors.DeleteMetadataConnector.
func (m DeleteMetadata) Run(t *testing.T, builder ConnectorBuilder[connectors.DeleteMetadataConnector]) {
	m.RunWithContext(t, t.Context(), builder)
}

// RunWithContext provides a procedure to test connectors.DeleteMetadataConnector.
func (m DeleteMetadata) RunWithContext(t *testing.T, ctx context.Context,
	builder ConnectorBuilder[connectors.DeleteMetadataConnector],
) {
	t.Helper()
	t.Cleanup(func() {
		DeleteMetadataType(m).Close()
	})

	conn := builder.Build(t, m.Name)
	output, err := conn.DeleteMetadata(ctx, m.Input)
	DeleteMetadataType(m).Validate(t, err, output)
}
