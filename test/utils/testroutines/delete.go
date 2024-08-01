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

func (r Delete) getOutline() suiteOutline[common.DeleteResult] {
	return suiteOutline[common.DeleteResult]{
		Name:         r.Name,
		Server:       r.Server,
		Comparator:   r.Comparator,
		Expected:     r.Expected,
		ExpectedErrs: r.ExpectedErrs,
	}
}

// DeleteConnector provides a procedure to test connectors.DeleteConnector
func (r routines) DeleteConnector(t *testing.T, testSuite Delete, builder ConnectorBuilder[connectors.DeleteConnector]) {
	defer testSuite.Server.Close()

	conn, err := builder()
	if err != nil {
		t.Fatalf("%s: error in test while constructing connector %v", testSuite.Name, err)
	}

	output, err := conn.Delete(context.Background(), testSuite.Input)
	testSuite.getOutline().Validate(t, err, output)
}
