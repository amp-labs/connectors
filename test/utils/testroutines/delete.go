package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

type (
	deleteType = TestCase[common.DeleteParams, *common.DeleteResult]
	// TestCaseDelete is a test suite useful for testing connectors.DeleteConnector interface.
	TestCaseDelete deleteType
)

// Run provides a procedure to test connectors.DeleteConnector
func (d TestCaseDelete) Run(t *testing.T, builder ConnectorBuilder[TestableDeleter]) {
	t.Helper()
	t.Cleanup(func() {
		deleteType(d).Close()
	})

	conn := builder.Build(t, d.Name)
	output, err := conn.Delete(t.Context(), deleteType(d).PrepareInput())
	deleteType(d).Validate(t, err, output)
}
