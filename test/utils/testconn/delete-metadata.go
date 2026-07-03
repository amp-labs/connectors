package testconn

import (
	"context"
	"testing"

	"github.com/amp-labs/connectors/common"
)

type (
	deleteMetadataType = TestCase[*common.DeleteMetadataParams, *common.DeleteMetadataResult]
	// TestCaseDeleteMetadata is a test suite useful for testing connectors.DeleteMetadataConnector interface.
	TestCaseDeleteMetadata deleteMetadataType
)

// Run provides a procedure to test connectors.DeleteMetadataConnector.
func (m TestCaseDeleteMetadata) Run(t *testing.T, builder ConnectorBuilder[TestableMetadataDeleter]) {
	m.RunWithContext(t, t.Context(), builder)
}

// RunWithContext provides a procedure to test connectors.DeleteMetadataConnector.
func (m TestCaseDeleteMetadata) RunWithContext(t *testing.T, ctx context.Context,
	builder ConnectorBuilder[TestableMetadataDeleter],
) {
	t.Helper()
	t.Cleanup(func() {
		deleteMetadataType(m).Close()
	})

	conn := builder.Build(t, m.Name)
	output, err := conn.DeleteMetadata(ctx, deleteMetadataType(m).PrepareInput())
	deleteMetadataType(m).Validate(t, err, output)
}
