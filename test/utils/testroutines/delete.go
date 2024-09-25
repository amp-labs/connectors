package testroutines

import (
	"context"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

type (
	DeleteType = TestCase[common.DeleteParams, *common.DeleteResult]
	// Delete is a test suite useful for testing connectors.DeleteConnector interface.
	Delete DeleteType
)

// Run provides a procedure to test connectors.DeleteConnector
func (d Delete) Run(t *testing.T, builder ConnectorBuilder[connectors.DeleteConnector]) {
	t.Helper()
	conn := builder.Build(t, d.Name)
	output, err := conn.Delete(context.Background(), d.Input)
	DeleteType(d).Validate(t, err, output)
}
