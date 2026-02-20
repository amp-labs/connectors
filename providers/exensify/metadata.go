package exensify

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResult := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, obj := range objectNames {
		metadata, err := c.fetchObjectMetadata(ctx, obj)
		if err != nil {
			metadataResult.Errors[obj] = err
		} else {
			metadataResult.Result[obj] = *metadata
		}
	}

	return &metadataResult, nil
}

func (c *Connector) fetchObjectMetadata(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	body, err := buildReadBody(objectName)
	if err != nil {
		return nil, err
	}

	resp, err := c.executeRequest(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata for object %s: %w", objectName, err)
	}

	var result map[string]any

	if err = json.Unmarshal(common.GetResponseBodyOnce(resp), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if err = checkResponseCode(result); err != nil {
		return nil, err
	}

	records, ok := result[readObjectResponseIdentifier.Get(objectName)].([]any)
	if !ok {
		return nil, fmt.Errorf("failed to parse metadata response for object %s: %w", objectName, common.ErrMissingExpectedValues)
	}

	if len(records) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
	}

	for field, value := range firstRecord {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    inferValueTypeFromData(value),
			ProviderType: "", // not available
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}
