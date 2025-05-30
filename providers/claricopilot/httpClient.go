package claricopilot

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// NewClariCopilotAuthHTTPClient returns a new http client with
// dual header authentication.
func NewClariCopilotAuthHTTPClient( //nolint:ireturn
	ctx context.Context,
	header1Name, header1Value, header2Name, header2Value string,
	opts ...common.HeaderAuthClientOption,
) (common.AuthenticatedHTTPClient, error) {

	headers := []common.Header{
		{
			Key:   header1Name,
			Value: header1Value,
		},
		{
			Key:   header2Name,
			Value: header2Value,
		},
	}

	// Use the existing header auth client with multiple headers
	return common.NewHeaderAuthHTTPClient(ctx, append(opts, common.WithHeaders(headers...))...)
}
