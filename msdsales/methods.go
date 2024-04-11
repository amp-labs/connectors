package msdsales

import (
	"context"
	"errors"
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/reqrepeater"
)

// FIXME these methods look repetitive
// FIXME arguments: Retry Strategy + HTTP Client.
func (c *Connector) get(ctx context.Context,
	url string, headers ...common.Header,
) (*common.JSONHTTPResponse, error) {
	retry := c.RetryStrategy.Start()

	for {
		rsp, err := c.Client.Get(ctx, url, headers...)
		err, ok := handleError(err)

		if ok {
			return rsp, nil
		}

		if !errors.Is(err, reqrepeater.ErrRetry) {
			// actual error from client
			return nil, err
		}
		// we can mitigate an error if we retry
		if retry.Completed() {
			return nil, err
		}
	}
}

func (c *Connector) post(ctx context.Context,
	url string, body any, headers ...common.Header,
) (*common.JSONHTTPResponse, error) {
	retry := c.RetryStrategy.Start()

	for {
		rsp, err := c.Client.Post(ctx, url, body, headers...)
		err, ok := handleError(err)

		if ok {
			return rsp, nil
		}

		if !errors.Is(err, reqrepeater.ErrRetry) {
			// actual error from client
			return nil, err
		}
		// we can mitigate an error if we retry
		if retry.Completed() {
			return nil, err
		}
	}
}

func (c *Connector) patch(ctx context.Context,
	url string, body any, headers ...common.Header,
) (*common.JSONHTTPResponse, error) {
	retry := c.RetryStrategy.Start()

	for {
		rsp, err := c.Client.Patch(ctx, url, body, headers...)
		err, ok := handleError(err)

		if ok {
			return rsp, nil
		}

		if !errors.Is(err, reqrepeater.ErrRetry) {
			// actual error from client
			return nil, err
		}
		// we can mitigate an error if we retry
		if retry.Completed() {
			return nil, err
		}
	}
}

func (c *Connector) delete(ctx context.Context,
	url string, headers ...common.Header,
) (*common.JSONHTTPResponse, error) {
	retry := c.RetryStrategy.Start()

	for {
		rsp, err := c.Client.Delete(ctx, url, headers...)
		err, ok := handleError(err)

		if ok {
			return rsp, nil
		}

		if !errors.Is(err, reqrepeater.ErrRetry) {
			// actual error from client
			return nil, err
		}
		// we can mitigate an error if we retry
		if retry.Completed() {
			return nil, err
		}
	}
}

func handleError(err error) (error, bool) {
	if err == nil {
		return nil, true
	}

	switch {
	case errors.Is(err, common.ErrAccessToken):
		slog.Warn("Access token invalid, retrying", "error", err)
		fallthrough
	case errors.Is(err, common.ErrRetryable):
		return reqrepeater.ErrRetry, false
	case errors.Is(err, common.ErrApiDisabled):
		fallthrough
	case errors.Is(err, common.ErrForbidden):
		fallthrough
	default:
		// Anything else is a permanent error
		return err, false
	}
}
