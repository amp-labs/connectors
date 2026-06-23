package subscriber

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/parallelfetch"
)

// DeleteSubscription removes all remote subscriptions for objects specified in previousResult.
func (s Strategy) DeleteSubscription(
	ctx context.Context,
	previousResult common.SubscriptionResult,
) error {
	subscriptionResult, err := s.TypedSubscriptionResult(previousResult)
	if err != nil {
		return err
	}

	webhooksToRemove := make([]int, len(subscriptionResult.ObjectWebhooks))

	index := 0
	for _, webhook := range subscriptionResult.ObjectWebhooks {
		webhooksToRemove[index] = webhook.ID
		index += 1
	}

	if len(webhooksToRemove) == 0 {
		return nil
	}

	result := s.removeSubscriptionsByIDs(ctx, webhooksToRemove)
	if len(result.Errors) != 0 {
		return errors.Join(result.Errors.Values()...)
	}

	return nil
}

func (s Strategy) removeSubscriptionsByIDs(
	ctx context.Context, identifiers []int,
) parallelfetch.Result[int, any] {
	tasks := make([]parallelfetch.Task[int, any], len(identifiers))
	for index, identifier := range identifiers {
		tasks[index] = func(ctx context.Context) (taskID int, data *any, err error) {
			url, err := s.getSubscriptionURL()
			if err != nil {
				return identifier, nil, err
			}

			url.AddPath(strconv.Itoa(identifier))

			resp, err := s.client.Delete(ctx, url.String(), s.clientIdHeader())
			if err != nil {
				return identifier, nil, err
			}

			if !httpkit.Status2xx(resp.Code) {
				return identifier, nil, fmt.Errorf("failed to remove subscription with id=%d", taskID) // nolint:err113
			}

			return identifier, new(any), nil
		}
	}

	return parallelfetch.Execute(ctx, tasks, -1)
}
