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

func (m Metadata) getOutline() suiteOutline[common.ListObjectMetadataResult] {
	return suiteOutline[common.ListObjectMetadataResult]{
		Name:         m.Name,
		Server:       m.Server,
		Comparator:   m.Comparator,
		Expected:     m.Expected,
		ExpectedErrs: m.ExpectedErrs,
	}
}

// Run provides a procedure to test connectors.ObjectMetadataConnector
func (m Metadata) Run(t *testing.T,
	builder ConnectorBuilder[connectors.ObjectMetadataConnector],
) {
	defer m.Server.Close()

	conn, err := builder()
	if err != nil {
		t.Fatalf("%s: error in test while constructing connector %v", m.Name, err)
	}

	output, err := conn.ListObjectMetadata(context.Background(), m.Input)
	m.getOutline().Validate(t, err, output)
}
