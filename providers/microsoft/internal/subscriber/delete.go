package subscriber

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
)

func (s Strategy) removeSubscriptionsByIds(
	ctx context.Context, ids []string,
) (*batch.Result[any], error) {
	batchParams, err := s.paramsForBatchRemoveSubscriptionsByIds(ids)
	if err != nil {
		return nil, err
	}

	bundledResponse := batch.Execute[any](ctx, s.batchStrategy, batchParams)

	return bundledResponse, nil
}

func (s Strategy) paramsForBatchRemoveSubscriptionsByIds(ids []string) (*batch.Params, error) {
	batchParams := &batch.Params{}

	for _, identifier := range ids {
		url, err := s.getSubscriptionURL()
		if err != nil {
			return nil, err
		}

		url.AddPath(identifier)

		// RequestID is Subscription identifier.
		batchParams.WithRequest(batch.RequestID(identifier), http.MethodDelete, url, nil, map[string]any{
			"Content-Type": "application/json",
		})
	}

	return batchParams, nil
}
