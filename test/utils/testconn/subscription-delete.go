package testconn

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

type (
	deleteSubscriptionType = TestCase[common.SubscriptionResult, None]
	// TestCaseDeleteSubscription is a test suite useful for testing part of connectors.SubscribeConnector interface.
	TestCaseDeleteSubscription deleteSubscriptionType
)

type DeleteSubscriptionParams struct {
	Params         common.SubscribeParams
	PreviousResult *common.SubscriptionResult
}

// Run provides a procedure to test connectors.SubscribeConnector
func (s TestCaseDeleteSubscription) Run(t *testing.T, builder ConnectorBuilder[TestableSubscriptionRemover]) {
	t.Helper()
	t.Cleanup(func() {
		deleteSubscriptionType(s).Close()
	})

	s.Expected = None{}

	conn := builder.Build(t, s.Name)
	err := conn.DeleteSubscription(t.Context(), deleteSubscriptionType(s).PrepareInput())
	deleteSubscriptionType(s).Validate(t, err, None{})
}
