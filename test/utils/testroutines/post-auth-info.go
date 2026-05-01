package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
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

	r.Comparator = func(serverURL string, actual, expected *common.PostAuthInfo) *testutils.CompareResult {
		result := testutils.NewCompareResult()

		result.Assert("ProviderWorkspaceRef", expected.ProviderWorkspaceRef, actual.ProviderWorkspaceRef)
		result.Assert("CatalogVars", expected.CatalogVars, actual.CatalogVars)

		if actual.RawResponse == nil && expected.RawResponse != nil {
			result.AddDiff("RawResponse is nil, expected non-nil")
		}

		if actual.RawResponse != nil && expected.RawResponse == nil {
			result.AddDiff("RawResponse is non-nil, expected nil")
		}

		if actual.RawResponse != nil && expected != nil {
			result.Assert("RawResponse.Code", expected.RawResponse.Code, actual.RawResponse.Code)

			if expected.RawResponse.Headers != nil {
				result.AddDiff("RawResponse.Headers is not supported by PostAuthInfo TestCase")
			}
		}

		return result
	}

	conn := builder.Build(t, r.Name)
	output, err := conn.GetPostAuthInfo(t.Context())
	PostAuthInfoType(r).Validate(t, err, output)
}
