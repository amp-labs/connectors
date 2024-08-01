package testroutines

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// Metadata is a test suite useful for testing connectors.ObjectMetadataConnector interface.
type Metadata struct {
	Name  string
	Input []string
	// dependencies
	Server *httptest.Server
	// custom comparison
	Comparator func(serverURL string, actual, expected *common.ListObjectMetadataResult) bool
	// output
	Expected     *common.ListObjectMetadataResult
	ExpectedErrs []error
}

func (r Metadata) getOutline() suiteOutline[common.ListObjectMetadataResult] {
	return suiteOutline[common.ListObjectMetadataResult]{
		Name:         r.Name,
		Server:       r.Server,
		Comparator:   r.Comparator,
		Expected:     r.Expected,
		ExpectedErrs: r.ExpectedErrs,
	}
}

// MetadataConnector provides a procedure to test connectors.ObjectMetadataConnector
func (r routines) MetadataConnector(t *testing.T, testSuite Metadata,
	builder ConnectorBuilder[connectors.ObjectMetadataConnector],
) {
	defer testSuite.Server.Close()

	conn, err := builder()
	if err != nil {
		t.Fatalf("%s: error in test while constructing connector %v", testSuite.Name, err)
	}

	output, err := conn.ListObjectMetadata(context.Background(), testSuite.Input)
	testSuite.getOutline().Validate(t, err, output)
}
