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

	getRecords := func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayRequired(responseFieldName)
	}

	return common.ParseResult(res,
		getRecords,
		getNextRecordsURL,
		common.MakeMarshaledDataFunc(flattenRecords),
		config.Fields,
	)
}

// flattenCallsMetaData is a custom transformer for calls objects that:
// 1. Flattens metaData fields to the top level
// 2. Preserves other top-level fields (context, parties, content, etc.)
// 3. Removes the metaData wrapper after flattening
func flattenRecords(node *ajson.Node) (map[string]any, error) {

	record, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	// Extract metaData if it exists
	metaDataNode, err := jsonquery.New(node).ObjectOptional("metaData")
	if err != nil {
		return nil, err
	}

	// If metaData exists, flatten it to the top level
	if metaDataNode != nil {
		metaData, err := jsonquery.Convertor.ObjectToMap(metaDataNode)
		if err != nil {
			return nil, err
		}

		// Remove the metaData wrapper
		delete(record, "metaData")

		// Add all metaData fields to the top level
		maps.Copy(record, metaData)
	}

	return record, nil
}
