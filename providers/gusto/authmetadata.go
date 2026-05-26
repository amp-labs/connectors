package gusto

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

type tokenInfoResponse struct {
	Resource struct {
		Type string `json:"type"`
		UUID string `json:"uuid"`
	} `json:"resource"`
}

func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	companyID, err := c.retrieveCompanyID(ctx)
	if err != nil {
		return nil, err
	}

	if companyID == "" {
		return nil, common.ErrMissingExpectedValues
	}

	c.companyId = companyID

	return &common.PostAuthInfo{
		CatalogVars: AuthMetadataVars{
			CompanyId: companyID,
		}.AsMap(),
		RawResponse: nil,
	}, nil
}

func (c *Connector) retrieveCompanyID(ctx context.Context) (string, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "/v1/token_info")
	if err != nil {
		return "", err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return "", err
	}

	if _, ok := resp.Body(); !ok {
		return "", common.ErrEmptyJSONHTTPResponse
	}

	data, err := common.UnmarshalJSON[tokenInfoResponse](resp)
	if err != nil {
		return "", common.ErrFailedToUnmarshalBody
	}

	return data.Resource.UUID, nil
}
