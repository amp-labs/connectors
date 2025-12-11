package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

type (
	BatchWriteType = TestCase[*common.BatchWriteParam, *common.BatchWriteResult]
	// BatchWrite is a test suite useful for testing connectors.BatchWriteConnector interface.
	BatchWrite BatchWriteType
)

// Run provides a procedure to test connectors.BatchWriteConnector
func (m BatchWrite) Run(t *testing.T, builder ConnectorBuilder[connectors.BatchWriteConnector]) {
	t.Helper()
	t.Cleanup(func() {
		BatchWriteType(m).Close()
	})

	conn := builder.Build(t, m.Name)
	output, err := conn.BatchWrite(t.Context(), m.Input)
	BatchWriteType(m).Validate(t, err, output)
}
