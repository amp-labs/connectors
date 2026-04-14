package testroutines

import (
	"reflect"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

type (
	PostAuthInfoType = TestCase[any, *common.PostAuthInfo]
	// PostAuthInfo is a test suite useful for testing connectors.AuthMetadataConnector interface.
	PostAuthInfo PostAuthInfoType
)

// Run provides a procedure to test connectors.AuthMetadataConnector
func (r PostAuthInfo) Run(t *testing.T, builder ConnectorBuilder[connectors.AuthMetadataConnector]) {
	t.Helper()
	t.Cleanup(func() {
		PostAuthInfoType(r).Close()
	})

	r.Comparator = func(serverURL string, actual, expected *common.PostAuthInfo) bool {
		if actual.ProviderWorkspaceRef != expected.ProviderWorkspaceRef {
			return false
		}

		if !reflect.DeepEqual(actual.CatalogVars, expected.CatalogVars) {
			return false
		}

		if actual.RawResponse == nil && expected.RawResponse != nil {
			return false
		}

		if actual.RawResponse != nil && expected.RawResponse == nil {
			return false
		}

		if actual.RawResponse != nil && expected != nil {
			if actual.RawResponse.Code != expected.RawResponse.Code {
				return false
			}

			if expected.RawResponse.Headers != nil {
				t.Fatalf("RawResoponse.Headers is not supported by PostAuthInfo TestCase")

				return false
			}
		}

		return true
	}

	conn := builder.Build(t, r.Name)
	output, err := conn.GetPostAuthInfo(t.Context())
	PostAuthInfoType(r).Validate(t, err, output)
}
