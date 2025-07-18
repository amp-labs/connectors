package snapchatads

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

var (
	ErrFailedToGetOrganizationId = errors.New("failed to get organization id")
	ErrDiscoveryFailure          = errors.New("failed to collect post authentication data")
)

func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	organizationId, err := c.retrieveOrganizationId(ctx)
	if err != nil {
		return nil, errors.Join(ErrDiscoveryFailure, err)
	}

	c.organizationId = organizationId

	return &common.PostAuthInfo{
		CatalogVars: AuthMetadataVars{
			OrganizationId: organizationId,
		}.AsMap(),
		RawResponse: nil,
	}, nil
}

func (c *Connector) retrieveOrganizationId(ctx context.Context) (string, error) {
	ctx = logging.With(ctx, "connector", "snapchatAds")

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "me")
	if err != nil {
		return "", err
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return "", err
	}

	body, ok := resp.Body()
	if !ok {
		return "", ErrFailedToGetOrganizationId
	}

	objectResponse, err := jsonquery.New(body).ObjectRequired("me")
	if err != nil {
		return "", err
	}

	organizationId, err := jsonquery.New(objectResponse).StringRequired("organization_id")
	if err != nil {
		return "", err
	}

	return organizationId, nil
}
