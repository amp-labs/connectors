package expensify

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/simultaneously"
)

func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	var mutex sync.Mutex

	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResult := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	// Tasks to be executed simultaneously.
	callbacks := make([]simultaneously.Job, 0, len(objectNames))

	for _, object := range objectNames {

		obj := object
		callbacks = append(callbacks, func(ctx context.Context) error {
			metadata, err := c.fetchObjectMetadata(ctx, obj)
			if err != nil {
				mutex.Lock()
				metadataResult.Errors[obj] = err
				mutex.Unlock()
				return nil //nolint:nilerr // intentionally collecting errors in map, not failing fast
			}

			mutex.Lock()
			metadataResult.Result[obj] = *metadata
			mutex.Unlock()

			return nil

		})

	}

	// This will block until all callbacks are done.
	if err := simultaneously.DoCtx(ctx, -1, callbacks...); err != nil {
		return nil, err
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

	//nolint:bodyclose
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

	//nolint:varnamelen
	records, ok := result[readObjectResponseIdentifier.Get(objectName)].([]any)
	if !ok {
		return nil, fmt.Errorf("failed to parse metadata response for object %s: %w",
			objectName, common.ErrMissingExpectedValues)
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
