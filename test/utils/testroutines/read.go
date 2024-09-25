package testroutines

import (
	"context"
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
	conn := builder.Build(t, r.Name)
	output, err := conn.Read(context.Background(), r.Input)
	ReadType(r).Validate(t, err, output)
}
