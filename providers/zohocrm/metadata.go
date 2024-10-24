package zohocrm

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

// restMetadataEndpoint is the resource for retrieving metadata details.
const restMetadataEndpoint = "settings/fields"

// apiKeyField is the key holding the metadata field name.
const apiKeyField = "api_name"

type metadataFields struct {
	Fields []map[string]any `json:"fields"`
}

func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	objMetadata := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, obj := range objectNames {
		url, err := c.getAPIURL(restMetadataEndpoint)
		if err != nil {
			return nil, err
		}

		capObj := naming.CapitalizeFirstLetterEveryWord(obj)

		// setting this, returns both used and unused fields
		url.WithQueryParam("type", "all")
		url.WithQueryParam("module", capObj)

		resp, err := c.Client.Get(ctx, url.String())
		if err != nil {
			objMetadata.Errors[obj] = err

			continue
		}

		metadata, err := metadataMapper(resp)
		if err != nil {
			return nil, err
		}

		metadata.DisplayName = capObj
		objMetadata.Result[obj] = *metadata
	}

	return &objMetadata, nil
}

func metadataMapper(resp *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[metadataFields](resp)
	if err != nil {
		return nil, err
	}

	metadata := &common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	// Ranging on the fields Slice, to construct the metadata fields.
	for _, f := range response.Fields {
		apiField, ok := f[apiKeyField].(string)
		if ok {
			metadata.FieldsMap[apiField] = apiField
		}
	}

	return metadata, nil
}
