package calendly

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	userInfoEndpoint = "users/me"
)

var (
	ErrUnexpectedUserURI = errors.New("malformed response: user URI missing or invalid type")
	ErrUnexpectedOrgURI  = errors.New("malformed response: organization URI missing or invalid type")
)

type UserResponse struct {
	Resource map[string]any `json:"resource"`
}

func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	userURI, orgURI, rawResp, err := c.retrieveURIs(ctx)
	if err != nil {
		return nil, err
	}

	c.userURI = userURI
	c.orgURI = orgURI

	cv := map[string]string{
		"userURI":         userURI,
		"organizationURI": orgURI,
	}

	return &common.PostAuthInfo{
		CatalogVars: &cv,
		RawResponse: rawResp,
	}, nil
}

// retrieveURIs fetches the authenticated user's URI and their current organization URI
// from the user info endpoint. These URIs are used to identify the user's context
// since a user can only belong to one organization at a time.
//
// Returns:
//   - userURI: The unique identifier URI for the authenticated user
//   - orgURI: The unique identifier URI for the user's current organization
//   - error: Any error that occurred during the request or parsing
func (c *Connector) retrieveURIs(ctx context.Context) (string, string, *common.JSONHTTPResponse, error) {
	// Build the full user info endpoint URL
	fullURL, err := urlbuilder.New(c.ModuleInfo().BaseURL, userInfoEndpoint)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to build user info URL: %w", err)
	}

	resp, err := c.JSONHTTPClient().Get(ctx, fullURL.String())
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to fetch user info: %w", err)
	}

	res, err := common.UnmarshalJSON[UserResponse](resp)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	userURI, ok := res.Resource["uri"].(string)
	if !ok {
		return "", "", nil, ErrUnexpectedUserURI
	}

	orgURI, ok := res.Resource["current_organization"].(string)
	if !ok {
		return "", "", nil, ErrUnexpectedOrgURI
	}

	return userURI, orgURI, resp, nil
}
