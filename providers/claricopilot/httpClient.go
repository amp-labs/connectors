package claricopilot

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// NewClariCopilotAuthHTTPClient returns a new http client with
// dual header authentication.
func NewClariCopilotAuthHTTPClient( //nolint:ireturn
	ctx context.Context,
	headerAPIKeyName, headerAPIKeyValue, headerPasswordName, headerPasswordValue string,
	opts ...common.HeaderAuthClientOption,
) (common.AuthenticatedHTTPClient, error) {
	headers := []common.Header{
		{
			Key:   headerAPIKeyName,
			Value: headerAPIKeyValue,
		},
		{
			Key:   headerPasswordName,
			Value: headerPasswordValue,
		},
	}
	// Use the existing header auth client with multiple headers
	return common.NewHeaderAuthHTTPClient(ctx, append(opts, common.WithHeaders(headers...))...)
}
