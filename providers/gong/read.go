package gong

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedObjectsByRead[c.Module.ID].Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.NextPage) != 0 { // not the first page, add a cursor
		url.WithQueryParam("cursor", config.NextPage.String())
	}

	if !config.Since.IsZero() {
		// This time format is RFC3339 but in UTC only.
		// See calls or users object for query parameter requirements.
		// https://gong.app.gong.io/settings/api/documentation#get-/v2/calls
		url.WithQueryParam("fromDateTime", datautils.Time.FormatRFC3339inUTC(config.Since))
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return &common.ReadResult{
				Rows:     0,
				Data:     nil,
				NextPage: "",
				Done:     true,
			}, nil
		}

		return nil, err
	}

	return common.ParseResult(res,
		common.GetRecordsUnderJSONPath(config.ObjectName),
		getNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}
