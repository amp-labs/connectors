package subscriber

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (s Strategy) RunScheduledMaintenance(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	return s.UpdateSubscription(ctx, params, previousResult)
}
