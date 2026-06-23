package subscriber

import (
	"context"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/parallelfetch"
)

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
