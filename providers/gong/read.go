package gong

import (
	"context"
	"errors"
	"maps"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/gong/metadata"
	"github.com/spyzhov/ajson"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedObjectsByRead[c.Module.ID].Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	// Handle flows specially since it requires dynamic flowOwnerEmail query param
	if config.ObjectName == objectNameFlows {
		return c.readFlows(ctx, config)
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

	if config.ObjectName == objectNameCalls {
		return common.ParseResult(res,
			getRecords(responseFieldName),
			getNextRecordsURL,
			common.MakeMarshaledDataFunc(flattenMetaData),
			config.Fields,
		)
	}

	return common.ParseResult(res,
		common.ExtractRecordsFromPath(responseFieldName),
		getNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}

// flattenMetaData transforms a single call record node by flattening metaData fields.
// This only affects the Fields output, Raw will maintain the original nested structure.
func flattenMetaData(node *ajson.Node) (map[string]any, error) {
	record, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	metaDataNode, err := jsonquery.New(node).ObjectOptional("metaData")
	if err != nil {
		return nil, err
	}

	// if metaData is not present, return the record as is
	if metaDataNode == nil {
		return record, nil
	}

	metaData, err := jsonquery.Convertor.ObjectToMap(metaDataNode)
	if err != nil {
		return nil, err
	}

	delete(record, "metaData")
	maps.Copy(record, metaData)

	return record, nil
}
