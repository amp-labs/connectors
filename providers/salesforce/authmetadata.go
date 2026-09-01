package salesforce

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// PostAuthCatalogVarUsername is the catalog-var key under which GetPostAuthInfo
// reports the connected user's Salesforce username. Callers (e.g. amp-labs/server)
// persist it on the connection's provider metadata and can later feed it into
// FlowSubscriptionConfig.IntegrationUsername without a fresh identity lookup.
const PostAuthCatalogVarUsername = "username"

var errUserinfoMissingUsername = errors.New("userinfo response carries no preferred_username")

// GetPostAuthInfo returns authentication metadata fetched from Salesforce right
// after a connection is established (see the PostAuthentication items in the
// provider catalog). It currently reports the connected user's username.
//
// Best-effort by design: the userinfo endpoint is only reachable by tokens
// minted with an identity scope (id/openid/profile/full — permission sets play
// no role), so a scope-limited token gets a 403 here. That must not fail
// connection creation, so any lookup error degrades to an empty PostAuthInfo
// with a warning. Downstream, a missing username simply means the flow-based
// Subscribe path falls back to its own userinfo lookup or the caller-supplied
// IntegrationUsername.
func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	resp, username, err := c.getUserinfo(ctx)
	if err != nil {
		logging.Logger(ctx).WarnContext(ctx,
			"could not resolve the connected user's username post-auth; "+
				"continuing without it (token likely lacks an identity scope)",
			"provider", c.Provider(),
			"error", err,
		)

		return &common.PostAuthInfo{}, nil
	}

	return &common.PostAuthInfo{
		RawResponse: resp,
		CatalogVars: &map[string]string{
			PostAuthCatalogVarUsername: username,
		},
	}, nil
}

// GetCurrentUsername returns the Salesforce username of the user the connector
// is authenticated as, via the OpenID Connect userinfo endpoint. Requires the
// token to carry an identity scope (id/openid/profile/full).
//
// Official docs with example request/response (see the preferred_username
// response parameter — "Username of the queried user"):
// https://help.salesforce.com/s/articleView?id=sf.remoteaccess_using_userinfo_endpoint.htm
func (c *Connector) GetCurrentUsername(ctx context.Context) (string, error) {
	_, username, err := c.getUserinfo(ctx)

	return username, err
}

// getUserinfo calls /services/oauth2/userinfo and extracts preferred_username.
func (c *Connector) getUserinfo(ctx context.Context) (*common.JSONHTTPResponse, string, error) {
	url, err := urlbuilder.New(c.getModuleURL(), "services/oauth2/userinfo")
	if err != nil {
		return nil, "", err
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch userinfo: %w", err)
	}

	type userinfo struct {
		PreferredUsername string `json:"preferred_username"`
	}

	info, err := common.UnmarshalJSON[userinfo](resp)
	if err != nil {
		return nil, "", err
	}

	if info == nil || info.PreferredUsername == "" {
		return nil, "", errUserinfoMissingUsername
	}

	return resp, info.PreferredUsername, nil
}
