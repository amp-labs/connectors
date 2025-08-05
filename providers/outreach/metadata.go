package outreach

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

type Data struct {
	Data []dataItem `json:"data"`
}

type includedObjects struct {
	Included []dataItem `json:"included,omitempty"`
}

type dataItem struct {
	Type          string         `json:"type"`
	ID            int            `json:"id"`
	Relationships map[string]any `json:"relationships"`
	Attributes    map[string]any `json:"attributes"`
	Links         map[string]any `json:"links"`
}

func (item dataItem) ToMapStringAny() (map[string]any, error) {
	jsonBytes, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DataItem: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	return result, nil
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
		// Constructing the  request url.
		url, err := c.getApiURL(obj)
		if err != nil {
			return nil, err
		}

		res, err := c.Client.Get(ctx, url.String())
		if err != nil {
			objMetadata.Errors[obj] = err

			continue
		}

		metadata, err := metadataMapper(res)
		if err != nil {
			objMetadata.Errors[obj] = err

			continue
		}

		metadata.DisplayName = obj
		objMetadata.Result[obj] = *metadata
	}

	return &objMetadata, nil
}

func metadataMapper(resp *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[Data](resp)
	if err != nil {
		return nil, err
	}

	metadata := &common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	attributes := response.Data[0].Attributes
	for k := range attributes {
		metadata.FieldsMap[k] = k
	}

	// Append id in the metadata response. Only adds it, if available.
	// 0 is not a valid id in outreach types. Id are read-only and starts at 1.
	if response.Data[0].ID != 0 {
		metadata.FieldsMap[idKey] = idKey
	}

	return metadata, nil
}
