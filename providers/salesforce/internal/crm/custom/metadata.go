package custom

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// UpsertMetadata creates or updates definition of a custom field.
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_upsertMetadata.htm
func (a *Adapter) UpsertMetadata(
	ctx context.Context, params *common.UpsertMetadataParams,
) (*common.UpsertMetadataResult, error) {
	return a.upsertCustomFields(ctx, params)
}

func (a *Adapter) upsertCustomFields(
	ctx context.Context, params *common.UpsertMetadataParams,
) (*common.UpsertMetadataResult, error) {
	payload, err := NewCustomFieldsPayload(params)
	if err != nil {
		return nil, err
	}

	response, err := performMetadataAPICall[UpsertMetadataResponse](ctx, a, payload)
	if err != nil {
		return nil, err
	}

	result, err := transformResponseToResult(response)
	if err != nil {
		return nil, err
	}

	return result, nil
}
