package servicedeskplus

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
			objectMetadata.Result[object] = *metadata // nolint:wsl_v5
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
	url, err := a.getAPIURL(objectName)
	if err != nil {
		return nil, err
	}

	response, err := a.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return a.parseMetadataResponse(objectName, response)
}

func (a *Adapter) parseMetadataResponse(objectName string, resp *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return nil, err
	}

	if response == nil || len(*response) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	data, ok := (*response)[objectName].([]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the data response field data to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	firstRecord, ok := data[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
	}

	objectMetadata := common.ObjectMetadata{
		DisplayName: objectName,
		Fields:      make(common.FieldsMetadata),
		FieldsMap:   make(map[string]string),
	}

	for field, value := range firstRecord {
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
