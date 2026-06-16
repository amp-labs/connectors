package mail

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/simultaneously"
)

func (a *Adapter) ListObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	var mu sync.Mutex //nolint: varnamelen

	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	objectMetadata := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata, len(objectNames)),
		Errors: make(map[string]error, len(objectNames)),
	}

	callbacks := make([]simultaneously.Job, 0, len(objectNames))

	for _, object := range objectNames {
		obj := object // capture loop variable

		callbacks = append(callbacks, func(ctx context.Context) error {
			metadata, err := a.retrieveSampleResponse(ctx, obj)
			if err != nil {
				mu.Lock()
				objectMetadata.Errors[obj] = err // nolint:wsl_v5
				mu.Unlock()

				return nil //nolint:nilerr // intentionally collecting errors in map, not failing fast
			}

			mu.Lock()
			objectMetadata.Result[obj] = *metadata // nolint:wsl_v5
			mu.Unlock()

			return nil
		})
	}

	if err := simultaneously.DoCtx(ctx, -1, callbacks...); err != nil {
		return nil, err
	}

	return &objectMetadata, nil
}

func (a *Adapter) retrieveSampleResponse(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {
	obj, err := lookupObject(objectName)
	if err != nil {
		return nil, err
	}

	url, err := a.getAPIURL(obj.path)
	if err != nil {
		return nil, err
	}

	response, err := a.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return parseMetadataResponse(objectName, obj, response)
}

func parseMetadataResponse(objectName string, obj objectDescriptor, resp *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	node, ok := resp.Body()
	if !ok {
		return nil, common.ErrMissingExpectedValues
	}

	records, err := extractRecordsFromKeyPath(obj.recordsPath)(node)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	objectMetadata := common.ObjectMetadata{
		DisplayName: objectName,
		Fields:      make(common.FieldsMetadata),
		FieldsMap:   make(map[string]string),
	}

	for field, value := range records[0] {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName: field,
			ValueType:   inferValue(value),
		}
	}

	return &objectMetadata, nil
}

func inferValue(value any) common.ValueType {
	v := reflect.ValueOf(value)

	switch v.Kind() { //nolint: exhaustive
	case reflect.String:
		return common.ValueTypeString
	case reflect.Float64:
		return common.ValueTypeFloat
	case reflect.Bool:
		return common.ValueTypeBoolean
	case reflect.Slice:
		return common.ValueTypeOther
	case reflect.Map:
		return common.ValueTypeOther
	default:
		return common.ValueTypeOther
	}
}
