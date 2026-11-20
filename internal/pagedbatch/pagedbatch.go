// Package pagedbatch walks several paginated collections at once, such as the custom fields of every
// job on a page. The first page of every collection goes out as one request batch
// (JSONHTTPClient.DoBatch), then the next page of every collection that has one, and so on. The
// walk keeps the batch's concurrency limit, its stop on a 429, and the reuse of saved responses when
// the read is tried again.
package pagedbatch

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

// NextFunc handles one response of the collection at index walk. It returns the request for that
// collection's next page and true (see Next), or false when the collection is done (see Done). The
// response's Err is the request's own error, such as a 404; NextFunc decides whether it ends the
// collection or, by returning an error, the whole walk.
type NextFunc func(walk int, rsp common.BatchResponse) (common.BatchRequest, bool, error)

// Run sends first, one request per collection, as a batch, passes every response to next in order,
// and sends the requests next returns as the following batch, until no collection has a next page.
// It stops at the first error from the batch or from next.
func Run(
	ctx context.Context,
	client *common.JSONHTTPClient,
	first []common.BatchRequest,
	next NextFunc,
	opts ...common.BatchOption,
) error {
	pending := first

	walks := make([]int, len(first))
	for i := range walks {
		walks[i] = i
	}

	for len(pending) > 0 {
		responses, err := client.DoBatch(ctx, pending, opts...)
		if err != nil {
			return err
		}

		nextRequests := make([]common.BatchRequest, 0, len(pending))
		nextWalks := make([]int, 0, len(pending))

		for index, rsp := range responses {
			req, more, err := next(walks[index], rsp)
			if err != nil {
				return err
			}

			if more {
				nextRequests = append(nextRequests, req)
				nextWalks = append(nextWalks, walks[index])
			}
		}

		pending, walks = nextRequests, nextWalks
	}

	return nil
}

// Get returns a GET request for url.
func Get(url string) common.BatchRequest {
	return common.BatchRequest{Method: http.MethodGet, URL: url}
}

// Next is what a NextFunc returns when its collection's next page is a GET of url.
func Next(url string) (common.BatchRequest, bool, error) {
	return Get(url), true, nil
}

// Done is what a NextFunc returns when its collection has no more pages.
func Done() (common.BatchRequest, bool, error) {
	return common.BatchRequest{}, false, nil
}

// Fail is what a NextFunc returns to stop the whole walk with err.
func Fail(err error) (common.BatchRequest, bool, error) {
	return common.BatchRequest{}, false, err
}
