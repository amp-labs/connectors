package subscriber

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers/connectwise/internal/core"
)

const messageVersion = "3.0.0"

// Strategy implements subscription lifecycle operations.
//
// It embeds SubscriptionInputOutput to provide type-safe handling of
// subscription input/output while conforming to the non-generic connectors.SubscribeConnector interface.
type Strategy struct {
	components.SubscriptionInputOutput[Request, Result]

	base *core.Base
}

// NewStrategy constructs a Strategy with required dependencies.
func NewStrategy(base *core.Base) *Strategy {
	return &Strategy{base: base}
}

// Request defines the subscription request payload for the webhook provider.
type Request struct {
	// WebhookURL is the endpoint where webhook messages will be delivered.
	WebhookURL string
}

// Result represents the subscription result payload.
type Result struct {
	// ObjectWebhooks maps object names to their associated webhook resources.
	ObjectWebhooks map[common.ObjectName]SubscriptionResource
}
