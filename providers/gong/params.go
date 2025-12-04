package gong

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"golang.org/x/oauth2"
)

// Option is a function which mutates the connector configuration.
type Option = func(params *parameters)

type parameters struct {
	paramsbuilder.Client
}

func newParams(opts []Option) (*common.ConnectorParams, error) { // nolint:unused
	oldParams, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	return &common.ConnectorParams{
		AuthenticatedClient: oldParams.Client.Caller.Client,
	}, nil
}

func (p parameters) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
	)
}

func WithClient(ctx context.Context, client *http.Client,
	config *oauth2.Config, token *oauth2.Token, opts ...common.OAuthOption,
) Option {
	return func(params *parameters) {
		params.WithOauthClient(ctx, client, config, token, opts...)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.WithAuthenticatedClient(client)
	}
}

func buildReadParams(url *urlbuilder.URL, config common.ReadParams) {
	if len(config.NextPage) != 0 { // not the first page, add a cursor
		url.WithQueryParam("cursor", config.NextPage.String())
	}

	if !config.Since.IsZero() {
		// This time format is RFC3339 but in UTC only.
		// See calls or users object for query parameter requirements.
		// https://gong.app.gong.io/settings/api/documentation#get-/v2/calls
		url.WithQueryParam("fromDateTime", datautils.Time.FormatRFC3339inUTC(config.Since))
	}
}

func buildReadBody(config common.ReadParams) map[string]any {
	filter := make(map[string]any)

	if !config.Since.IsZero() {
		filter["fromDateTime"] = datautils.Time.FormatRFC3339inUTC(config.Since)
	}

	if !config.Until.IsZero() {
		filter["toDateTime"] = datautils.Time.FormatRFC3339inUTC(config.Until)
	}

	body := map[string]any{
		"filter": filter,
	}

	if len(config.NextPage) != 0 {
		body["cursor"] = config.NextPage.String()
	}

	if config.ObjectName == objectNameCalls {
		// https://app.gong.io/settings/api/documentation#post-/v2/calls/extensive
		body["contentSelector"] = map[string]any{
			"context": "Extended",
			"exposedFields": map[string]any{
				"parties": true,
				"media": true,
			},
		}
	}

	return body
}
