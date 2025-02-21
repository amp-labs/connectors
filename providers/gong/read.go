package gong

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/gong/metadata"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedObjectsByRead[c.Module.ID].Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.getReadURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	var res *common.JSONHTTPResponse

	if postReadObjects.Has(config.ObjectName) {
		body := buildReadBody(config)
		res, err = c.Client.Post(ctx, url.String(), body)
	} else {
		buildReadParams(url, config)
		res, err = c.Client.Get(ctx, url.String())
	}

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

	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module.ID, config.ObjectName)

	return common.ParseResult(res,
		common.GetRecordsUnderJSONPath(responseFieldName),
		getNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}
