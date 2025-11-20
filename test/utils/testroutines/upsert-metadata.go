package testroutines

import (
	"context"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

type (
	UpsertMetadataType = TestCase[*common.UpsertMetadataParams, *common.UpsertMetadataResult]
	// UpsertMetadata is a test suite useful for testing connectors.UpsertMetadataConnector interface.
	UpsertMetadata UpsertMetadataType
)

// Run provides a procedure to test connectors.UpsertMetadataConnector.
func (m UpsertMetadata) Run(t *testing.T, builder ConnectorBuilder[connectors.UpsertMetadataConnector]) {
	m.RunWithContext(t, t.Context(), builder)
}

// RunWithContext provides a procedure to test connectors.UpsertMetadataConnector.
func (m UpsertMetadata) RunWithContext(t *testing.T, ctx context.Context,
	builder ConnectorBuilder[connectors.UpsertMetadataConnector],
) {
	t.Helper()
	t.Cleanup(func() {
		UpsertMetadataType(m).Close()
	})

	conn := builder.Build(t, m.Name)
	output, err := conn.UpsertMetadata(ctx, m.Input)
	UpsertMetadataType(m).Validate(t, err, output)
}
