package mailgun

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/mailgun/metadata"
)

// objectDeleteScopes lists the objects with an item-level DELETE endpoint,
// reached as <collection>/<RecordId>. RecordId is the object's natural key:
// address (suppressions, lists), value (whitelists), name (templates), id
// (routes, forwards) or webhook_id.
//
// messages has no delete; lists/members would need the parent list as a second
// identifier, which the connector API does not support, so it is excluded.
//
// API reference: https://documentation.mailgun.com/docs/mailgun/api-reference/intro/
//
//nolint:gochecknoglobals
var objectDeleteScopes = map[string]objectScope{
	// Domain-scoped.
	"bounces":      scopeDomain,
	"complaints":   scopeDomain,
	"unsubscribes": scopeDomain,
	"whitelists":   scopeDomain,
	"templates":    scopeDomain,

	// Account-scoped.
	"account/templates": scopeAccount,
	"lists":             scopeAccount,
	"routes":            scopeAccount,
	"forwards":          scopeAccount,
	"webhooks":          scopeAccount,
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	scope, ok := objectDeleteScopes[params.ObjectName]
	if !ok {
		return nil, common.ErrOperationNotSupportedForObject
	}

	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	if scope == scopeDomain {
		path, err = c.substituteDomain(path)
		if err != nil {
			return nil, err
		}
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	url.AddPath(params.RecordId)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if response.Code != http.StatusOK && response.Code != http.StatusAccepted &&
		response.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	return &common.DeleteResult{Success: true}, nil
}
