package instantly

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedObjectsByRead.Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	urlPath, nodePath, err := matchObjectNameToEndpointPath(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config, urlPath)
	if err != nil {
		return nil, err
	}

	rsp, err := c.JSON.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		common.GetRecordsUnderJSONPath(nodePath),
		makeNextRecordsURL(url),
		common.GetMarshaledData,
		config.Fields,
	)
}

func matchObjectNameToEndpointPath(objectName string) (urlPath string, nodePath string, err error) {
	switch objectName {
	case objectNameCampaigns:
		// https://developer.instantly.ai/campaign-1/list-campaigns
		// Empty string of data location means the response is an array itself holding what we need.
		return "campaign/list", "", nil
	case objectNameAccounts:
		// https://developer.instantly.ai/account/list-accounts
		return "account/list", "accounts", nil
	case objectNameEmails:
		// https://developer.instantly.ai/unibox/emails-or-list
		return "unibox/emails", "data", nil
	case objectNameTags:
		// https://developer.instantly.ai/tags/list-tags
		return "custom-tag", "data", nil
	default:
		return "", "", common.ErrOperationNotSupportedForObject
	}
}

func (c *Connector) buildReadURL(config common.ReadParams, urlPath string) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return urlbuilder.New(config.NextPage.String())
	}

	// First page
	url, err := c.getURL(urlPath)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("skip", "0")
	url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

	return url, nil
}
