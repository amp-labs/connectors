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

// Run provides a procedure to test connectors.WriteConnector
func (w Write) Run(t *testing.T, builder ConnectorBuilder[connectors.WriteConnector]) {
	defer w.Server.Close()

	conn, err := builder()
	if err != nil {
		t.Fatalf("%s: error in test while constructing connector %v", w.Name, err)
	}

	output, err := conn.Write(context.Background(), w.Input)
	w.getOutline().Validate(t, err, output)
}
