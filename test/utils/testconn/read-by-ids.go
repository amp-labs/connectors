package testconn

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

type (
	readByIdsType = TestCase[ReadByIdsParams, []common.ReadResultRow]
	// TestCaseGetRecordsByIds is a test suite useful for testing connectors.BatchRecordReaderConnector interface.
	TestCaseGetRecordsByIds readByIdsType
)

type ReadByIdsParams struct {
	ObjectName   string
	RecordIds    []string
	Fields       []string
	Associations []string
}

// Run provides a procedure to test connectors.BatchRecordReaderConnector
func (r TestCaseGetRecordsByIds) Run(t *testing.T, builder ConnectorBuilder[TestableBatchReader]) {
	t.Helper()
	t.Cleanup(func() {
		readByIdsType(r).Close()
	})

	conn := builder.Build(t, r.Name)
	input := readByIdsType(r).PrepareInput()
	output, err := conn.GetRecordsByIds(t.Context(),
		input.ObjectName, input.RecordIds,
		input.Fields, input.Associations,
	)

	// Only the fetched rows are compared here. Per-id failures are asserted by
	// the connectors that populate them, not by this shared happy-path runner.
	var rows []common.ReadResultRow
	if output != nil {
		rows = output.Rows
	}

	readByIdsType(r).Validate(t, err, rows)
}
