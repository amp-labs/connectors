package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

type (
	updateSubscriptionType = TestCase[UpdateSubscriptionParams, *common.SubscriptionResult]
	// TestCaseUpdateSubscription is a test suite useful for testing part of connectors.SubscribeConnector interface.
	TestCaseUpdateSubscription updateSubscriptionType
)

type UpdateSubscriptionParams struct {
	Params         common.SubscribeParams
	PreviousResult *common.SubscriptionResult
}

// Run provides a procedure to test connectors.SubscribeConnector
func (s TestCaseUpdateSubscription) Run(t *testing.T, builder ConnectorBuilder[TestableSubscriptionUpdater]) {
	t.Helper()
	t.Cleanup(func() {
		updateSubscriptionType(s).Close()
	})

	conn := builder.Build(t, s.Name)
	input := updateSubscriptionType(s).PrepareInput()
	output, err := conn.UpdateSubscription(t.Context(), s.Input.Params, input.PreviousResult)
	updateSubscriptionType(s).Validate(t, err, output)
}
