package deps

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// SubscriptionResultLister lists the stored subscription results for an installation.
type SubscriptionResultLister interface {
	// ListSubscriptionResults returns the stored subscription results for the given
	// installation. emptyResult is a factory producing a fresh container with the provider-
	// specific .Result type populated as a zero value, which the implementation hydrates from
	// storage (mirroring the server's typed-container subscription loading).
	ListSubscriptionResults(
		ctx context.Context,
		installationID string,
		emptyResult func() *common.SubscriptionResult,
	) ([]*common.SubscriptionResult, error)
}
