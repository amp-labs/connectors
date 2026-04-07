package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

type (
	ReadByIdsType = TestCase[ReadByIdsParams, []common.ReadResultRow]
	// ReadByIds is a test suite useful for testing connectors.BatchRecordReaderConnector interface.
	ReadByIds ReadByIdsType
)

type ReadByIdsParams struct {
	ObjectName   string
	RecordIds    []string
	Fields       []string
	Associations []string
}

// Run provides a procedure to test connectors.BatchRecordReaderConnector
func (r ReadByIds) Run(t *testing.T, builder ConnectorBuilder[connectors.BatchRecordReaderConnector]) {
	t.Helper()
	t.Cleanup(func() {
		ReadByIdsType(r).Close()
	})

	conn := builder.Build(t, r.Name)
	output, err := conn.GetRecordsByIds(t.Context(),
		r.Input.ObjectName, r.Input.RecordIds,
		r.Input.Fields, r.Input.Associations,
	)
	ReadByIdsType(r).Validate(t, err, output)
}
