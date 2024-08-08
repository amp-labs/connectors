package testroutines

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// Read is a test suite useful for testing connectors.ReadConnector interface.
type Read struct {
	Name  string
	Input common.ReadParams
	// dependencies
	Server *httptest.Server
	// custom comparison
	Comparator func(serverURL string, actual, expected *common.ReadResult) bool
	// output
	Expected     *common.ReadResult
	ExpectedErrs []error
}

func (r Read) getOutline() suiteOutline[common.ReadResult] {
	// It is better to create suiteOutline on a fly then store it directly on Read struct.
	// Nested objects make the test ugly. Better this function then all test files.
	return suiteOutline[common.ReadResult]{
		Name:         r.Name,
		Server:       r.Server,
		Comparator:   r.Comparator,
		Expected:     r.Expected,
		ExpectedErrs: r.ExpectedErrs,
	}
}

// Run provides a procedure to test connectors.ReadConnector
func (r Read) Run(t *testing.T, builder ConnectorBuilder[connectors.ReadConnector]) {
	defer r.Server.Close()

	conn, err := builder()
	if err != nil {
		t.Fatalf("%s: error in test while constructing connector %v", r.Name, err)
	}

	output, err := conn.Read(context.Background(), r.Input)
	r.getOutline().Validate(t, err, output)
}
