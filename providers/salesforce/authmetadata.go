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
// FlowConfig.IntegrationUsername without a fresh identity lookup.
const PostAuthCatalogVarUsername = "username"

var errUserinfoMissingUsername = errors.New("userinfo response carries no preferred_username")

// GetPostAuthInfo is the connect-time hook the server calls right after OAuth.
// It GETs /services/oauth2/userinfo and returns preferred_username as
// CatalogVars["username"] for the server to persist on the connection.
// Failures are swallowed (empty result + warning): a token without an
// identity scope must not block creating the connection. Subscribe then
// either uses that stored username as IntegrationUsername or falls back to
// GetConnectedUsername.
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

// GetConnectedUsername is the subscribe-time lookup used only when
// IntegrationUsername was not passed (no stored post-auth username). Same
// userinfo endpoint as GetPostAuthInfo, but errors on failure — an outbound
// message cannot deploy without an integration user.
// Requires an identity scope (id/openid/profile/full).
// https://help.salesforce.com/s/articleView?id=sf.remoteaccess_using_userinfo_endpoint.htm
func (c *Connector) GetConnectedUsername(ctx context.Context) (string, error) {
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
