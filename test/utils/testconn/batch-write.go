package testconn

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

type (
	batchWriteType = TestCase[*common.BatchWriteParam, *common.BatchWriteResult]
	// TestCaseBatchWrite is a test suite useful for testing connectors.BatchWriteConnector interface.
	TestCaseBatchWrite batchWriteType
)

// Run provides a procedure to test connectors.BatchWriteConnector
func (m TestCaseBatchWrite) Run(t *testing.T, builder ConnectorBuilder[TestableBatchWriter]) {
	t.Helper()
	t.Cleanup(func() {
		batchWriteType(m).Close()
	})

	conn := builder.Build(t, m.Name)
	output, err := conn.BatchWrite(t.Context(), batchWriteType(m).PrepareInput())
	batchWriteType(m).Validate(t, err, output)
}
