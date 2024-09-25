package testroutines

import (
	"context"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

type (
	WriteType = TestCase[common.WriteParams, *common.WriteResult]
	// Write is a test suite useful for testing connectors.WriteConnector interface.
	Write WriteType
)

// Run provides a procedure to test connectors.WriteConnector
func (w Write) Run(t *testing.T, builder ConnectorBuilder[connectors.WriteConnector]) {
	t.Helper()
	conn := builder.Build(t, w.Name)
	output, err := conn.Write(context.Background(), w.Input)
	WriteType(w).Validate(t, err, output)
}
