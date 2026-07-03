package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

type (
	writeType = TestCase[common.WriteParams, *common.WriteResult]
	// TestCaseWrite is a test suite useful for testing connectors.WriteConnector interface.
	TestCaseWrite writeType
)

// Run provides a procedure to test connectors.WriteConnector
func (w TestCaseWrite) Run(t *testing.T, builder ConnectorBuilder[TestableWriter]) {
	t.Helper()
	t.Cleanup(func() {
		writeType(w).Close()
	})

	conn := builder.Build(t, w.Name)
	output, err := conn.Write(t.Context(), writeType(w).PrepareInput())
	writeType(w).Validate(t, err, output)
}
