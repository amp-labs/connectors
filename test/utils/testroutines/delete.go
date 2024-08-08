package testroutines

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// Delete is a test suite useful for testing connectors.DeleteConnector interface.
type Delete struct {
	Name  string
	Input common.DeleteParams
	// dependencies
	Server *httptest.Server
	// custom comparison
	Comparator func(serverURL string, actual, expected *common.DeleteResult) bool
	// output
	Expected     *common.DeleteResult
	ExpectedErrs []error
}

func (d Delete) getOutline() suiteOutline[common.DeleteResult] {
	return suiteOutline[common.DeleteResult]{
		Name:         d.Name,
		Server:       d.Server,
		Comparator:   d.Comparator,
		Expected:     d.Expected,
		ExpectedErrs: d.ExpectedErrs,
	}
}

// Run provides a procedure to test connectors.DeleteConnector
func (d Delete) Run(t *testing.T, builder ConnectorBuilder[connectors.DeleteConnector]) {
	defer d.Server.Close()

	conn, err := builder()
	if err != nil {
		t.Fatalf("%s: error in test while constructing connector %v", d.Name, err)
	}

	output, err := conn.Delete(context.Background(), d.Input)
	d.getOutline().Validate(t, err, output)
}
