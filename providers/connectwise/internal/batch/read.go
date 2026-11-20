package batch

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/parallelfetch"
	"github.com/amp-labs/connectors/providers/connectwise/internal/core"
)

// maxConcurrency is the number of goroutines that can be spawned to run multiple requests in parallel.
// The number 3 is chosen at random.
const maxConcurrency = 3

// Result is the outcome of a batch read. The identifiers are split across several requests,
// and a failed request does not discard the records returned by the others.
type Result[B any] struct {
	// Records are the records returned by the requests that succeeded.
	Records []B
	// Failures are the requests that failed, in the order they were built.
	Failures []Failure
}

// Failure is a request that failed, with the identifiers it asked for.
type Failure struct {
	Identifiers []string
	Err         error
}

// Err joins the errors of the failed requests. It is nil when every request succeeded.
func (r Result[B]) Err() error {
	errs := make([]error, 0, len(r.Failures))
	for _, failure := range r.Failures {
		errs = append(errs, failure.Err)
	}

	return errors.Join(errs...)
}

// Read performs a batch read for the requested object type and identifiers.
// The returned error is set only when the requests could not be built; errors of the requests
// themselves are in Result.Failures.
func Read[B any](ctx context.Context,
	base *core.Base,
	objectName string,
	identifiers []string,
) (*Result[B], error) {
	baseURL, err := base.GetUrl(objectName)
	if err != nil {
		return nil, err
	}

	urls, err := withIdentifiers(baseURL, identifiers, maxUrlSize)
	if err != nil {
		return nil, err
	}

	tasks := make([]parallelfetch.Task[int, readResponse[B]], len(urls))
	for index, wrapper := range urls {
		tasks[index] = func(ctx context.Context) (taskID int, data *readResponse[B], err error) {
			result := make(readResponse[B], 0)
			url := wrapper.URL

			// Paginated read until no `next` links is present in the header.
			for url != "" {
				res, err := base.JSONHTTPClient().Get(ctx, url, base.ClientIdHeader())
				if err != nil {
					return index, nil, err
				}

				// Parse the response to obtain any API errors.
				apiResponse, err := common.UnmarshalJSON[readResponse[B]](res)
				if err != nil {
					return index, nil, err
				}

				result = append(result, *apiResponse...)

				// Repeat for the next page if any.
				url = httpkit.HeaderLink(res, "next")
			}

			return index, &result, nil
		}
	}

	response := parallelfetch.Execute(ctx, tasks, maxConcurrency)

	result := &Result[B]{
		Records: make([]B, 0),
	}

	for index, wrapper := range urls {
		if err, failed := response.Errors[index]; failed {
			result.Failures = append(result.Failures, Failure{
				Identifiers: wrapper.identifiers,
				Err:         err,
			})

			continue
		}

		result.Records = append(result.Records, response.Records[index]...)
	}

	return result, nil
}

// readResponse mirrors the ConnectWise read response which is array at the root level.
// The generic is accepted to allow .
type readResponse[B any] []B
