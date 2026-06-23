package batch

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/parallelfetch"
)

// maxConcurrency is the number of goroutines that can be spawned to run multiple requests in parallel.
// The number 3 is chosen at random.
const maxConcurrency = 3

// Read performs a batch read for the requested object type and identifiers.
func Read[B any](ctx context.Context,
	adapter *Adapter,
	objectName string,
	identifiers []string,
) ([]B, error) {
	baseURL, err := adapter.getURL(objectName)
	if err != nil {
		return nil, err
	}

	urls, err := withIdentifiers(baseURL, identifiers, maxURLSize)
	if err != nil {
		return nil, err
	}

	tasks := make([]parallelfetch.Task[int, readResponse[B]], len(urls))
	for index, wrapper := range urls {
		tasks[index] = func(ctx context.Context) (taskID int, data *readResponse[B], err error) {
			res, err := adapter.client.Get(ctx, wrapper.URL, adapter.clientIdHeader())
			if err != nil {
				return index, nil, err
			}

			// Parse the response to obtain any API errors.
			apiResponse, err := common.UnmarshalJSON[readResponse[B]](res)
			if err != nil {
				return index, nil, err
			}

			return index, apiResponse, nil
		}
	}

	result := parallelfetch.Execute(ctx, tasks, maxConcurrency)
	if len(result.Errors) != 0 {
		return nil, errors.Join(result.Errors.Values()...)
	}

	records := make([]B, 0)
	for _, list := range result.Records {
		records = append(records, list...)
	}

	return records, nil
}

// readResponse mirrors the ConnectWise read response which is array at the root level.
// The generic is accepted to allow .
type readResponse[B any] []B
