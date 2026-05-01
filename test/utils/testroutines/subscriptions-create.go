package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

type (
	CreateSubscriptionType = TestCase[common.SubscribeParams, *common.SubscriptionResult]
	// CreateSubscription is a test suite useful for testing part of connectors.SubscribeConnector interface.
	CreateSubscription CreateSubscriptionType
)

// Run provides a procedure to test connectors.SubscribeConnector
func (s CreateSubscription) Run(t *testing.T, builder ConnectorBuilder[components.SubscriptionCreator]) {
	t.Helper()
	t.Cleanup(func() {
		CreateSubscriptionType(s).Close()
	})

	conn := builder.Build(t, s.Name)
	output, err := conn.Subscribe(t.Context(), s.Input)
	CreateSubscriptionType(s).Validate(t, err, output)
}
