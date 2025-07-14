package calendly

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

var (
	ErrDiscoveryFailure = errors.New("failed to collect post authentication data")
)

func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	userURI, orgURI, err := c.retrieveUserAndOrgURIs(ctx)
	if err != nil {
		return nil, errors.Join(ErrDiscoveryFailure, err)
	}

	c.userURI = userURI

	catalogVars := AuthMetadataVars{
		UserURI:         userURI,
		OrganizationURI: orgURI,
	}

	return &common.PostAuthInfo{
		CatalogVars: catalogVars.AsMap(),
	}, nil
}

func (c *Connector) retrieveUserAndOrgURIs(ctx context.Context) (string, string, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "users", "me")
	if err != nil {
		return "", "", err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return "", "", err
	}

	body, ok := resp.Body()
	if !ok {
		return "", "", errors.New("failed to get response body")
	}

	resource, err := jsonquery.New(body).ObjectRequired("resource")
	if err != nil {
		return "", "", err
	}

	userURI, err := jsonquery.New(resource).StringRequired("uri")
	if err != nil {
		return "", "", err
	}

	orgURI, err := jsonquery.New(resource).StringRequired("current_organization")
	if err != nil {
		return "", "", err
	}

	return userURI, orgURI, nil
} 