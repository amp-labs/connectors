package smartlead

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// Delete removes object. As of now only removal of Campaigns is allowed.
func (c *Connector) Delete(ctx context.Context, config common.DeleteParams) (*common.DeleteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !supportedObjectsByDelete.Has(config.ObjectName) {
		// Removing campaign is the only to be supported at this time.
		// https://api.smartlead.ai/reference/delete-campaign
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.AddPath(config.RecordId)

	// 200 OK is expected
	_, err = c.JSON.Delete(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, arg)
}
