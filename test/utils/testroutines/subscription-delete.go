package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

type (
	DeleteSubscriptionType = TestCase[common.SubscriptionResult, None]
	// DeleteSubscription is a test suite useful for testing part of connectors.SubscribeConnector interface.
	DeleteSubscription DeleteSubscriptionType
)

type DeleteSubscriptionParams struct {
	Params         common.SubscribeParams
	PreviousResult *common.SubscriptionResult
}

// Run provides a procedure to test connectors.SubscribeConnector
func (s DeleteSubscription) Run(t *testing.T, builder ConnectorBuilder[components.SubscriptionRemover]) {
	t.Helper()
	t.Cleanup(func() {
		DeleteSubscriptionType(s).Close()
	})

	s.Expected = None{}

	conn := builder.Build(t, s.Name)
	err := conn.DeleteSubscription(t.Context(), s.Input)
	DeleteSubscriptionType(s).Validate(t, err, None{})
}
