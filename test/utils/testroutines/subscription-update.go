package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

type (
	UpdateSubscriptionType = TestCase[UpdateSubscriptionParams, *common.SubscriptionResult]
	// UpdateSubscription is a test suite useful for testing part of connectors.SubscribeConnector interface.
	UpdateSubscription UpdateSubscriptionType
)

type UpdateSubscriptionParams struct {
	Params         common.SubscribeParams
	PreviousResult *common.SubscriptionResult
}

// Run provides a procedure to test connectors.SubscribeConnector
func (s UpdateSubscription) Run(t *testing.T, builder ConnectorBuilder[components.SubscriptionUpdater]) {
	t.Helper()
	t.Cleanup(func() {
		UpdateSubscriptionType(s).Close()
	})

	conn := builder.Build(t, s.Name)
	output, err := conn.UpdateSubscription(t.Context(), s.Input.Params, s.Input.PreviousResult)
	UpdateSubscriptionType(s).Validate(t, err, output)
}
