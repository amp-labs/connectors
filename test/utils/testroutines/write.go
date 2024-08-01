package testroutines

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// Write is a test suite useful for testing connectors.WriteConnector interface.
type Write struct {
	Name  string
	Input common.WriteParams
	// dependencies
	Server *httptest.Server
	// custom comparison
	Comparator func(serverURL string, actual, expected *common.WriteResult) bool
	// output
	Expected     *common.WriteResult
	ExpectedErrs []error
}

func (w Write) getOutline() suiteOutline[connectors.WriteResult] {
	return suiteOutline[connectors.WriteResult]{
		Name:         w.Name,
		Server:       w.Server,
		Comparator:   w.Comparator,
		Expected:     w.Expected,
		ExpectedErrs: w.ExpectedErrs,
	}
}

// WriteConnector provides a procedure to test connectors.WriteConnector
func (routines) WriteConnector(t *testing.T, testSuite Write, builder ConnectorBuilder[connectors.WriteConnector]) {
	defer testSuite.Server.Close()

	conn, err := builder()
	if err != nil {
		t.Fatalf("%s: error in test while constructing connector %v", testSuite.Name, err)
	}

	output, err := conn.Write(context.Background(), testSuite.Input)
	testSuite.getOutline().Validate(t, err, output)
}
