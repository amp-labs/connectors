package subscriber

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
)

func (s Strategy) removeSubscriptionsByIDs(
	ctx context.Context, identifiers []SubscriptionID,
) (*batch.Result[any], error) {
	batchParams, err := s.paramsForBatchRemoveSubscriptionsByIDs(identifiers)
	if err != nil {
		return nil, err
	}

	bundledResponse := batch.Execute[any](ctx, s.batchStrategy, batchParams)

	return bundledResponse, nil
}

func (s Strategy) paramsForBatchRemoveSubscriptionsByIDs(identifiers []SubscriptionID) (*batch.Params, error) {
	batchParams := &batch.Params{}

	for _, identifier := range identifiers {
		url, err := s.getSubscriptionURL()
		if err != nil {
			return nil, err
		}

		url.AddPath(string(identifier))

		// RequestID is Subscription identifier.
		batchParams.WithRequest(batch.RequestID(identifier), http.MethodDelete, url, nil, map[string]any{
			"Content-Type": "application/json",
		})
	}

	return batchParams, nil
}
