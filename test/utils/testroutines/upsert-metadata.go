package testroutines

import (
	"context"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

type (
	upsertMetadataType = TestCase[*common.UpsertMetadataParams, *common.UpsertMetadataResult]
	// TestCaseUpsertMetadata is a test suite useful for testing connectors.UpsertMetadataConnector interface.
	TestCaseUpsertMetadata upsertMetadataType
)

// Run provides a procedure to test connectors.UpsertMetadataConnector.
func (m TestCaseUpsertMetadata) Run(t *testing.T, builder ConnectorBuilder[connectors.UpsertMetadataConnector]) {
	m.RunWithContext(t, t.Context(), builder)
}

// RunWithContext provides a procedure to test connectors.UpsertMetadataConnector.
func (m TestCaseUpsertMetadata) RunWithContext(t *testing.T, ctx context.Context,
	builder ConnectorBuilder[connectors.UpsertMetadataConnector],
) {
	t.Helper()
	t.Cleanup(func() {
		upsertMetadataType(m).Close()
	})

	conn := builder.Build(t, m.Name)
	output, err := conn.UpsertMetadata(ctx, upsertMetadataType(m).PrepareInput())
	upsertMetadataType(m).Validate(t, err, output)
}
