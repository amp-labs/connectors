package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

type (
	createSubscriptionType = TestCase[common.SubscribeParams, *common.SubscriptionResult]
	// TestCaseSubscribe is a test suite useful for testing part of connectors.SubscribeConnector interface.
	TestCaseSubscribe createSubscriptionType
)

// Run provides a procedure to test connectors.SubscribeConnector
func (s TestCaseSubscribe) Run(t *testing.T, builder ConnectorBuilder[TestableSubscriptionCreator]) {
	t.Helper()
	t.Cleanup(func() {
		createSubscriptionType(s).Close()
	})

	conn := builder.Build(t, s.Name)
	output, err := conn.Subscribe(t.Context(), createSubscriptionType(s).PrepareInput())
	createSubscriptionType(s).Validate(t, err, output)
}
