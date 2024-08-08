package outreach

import (
	"context"
	"encoding/json"

	"github.com/amp-labs/connectors/common"
)

type Data struct {
	Data []DataItem `json:"data"`
}

type DataItem struct {
	Type          string         `json:"type"`
	ID            int            `json:"id"`
	Relationships map[string]any `json:"relationships"`
	Attributes    map[string]any `json:"attributes"`
}

func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames is not empty
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

		// Check nil response body, to avoid panic.
		if res == nil || res.Body == nil {
			objMetadata.Errors[obj] = common.ErrEmptyResponse

			continue
		}

		metadata, err := metadataMapper(res.Body.Source())
		if err != nil {
			return nil, err
		}

		metadata.DisplayName = obj
		objMetadata.Result[obj] = metadata
	}

	return &objMetadata, nil
}

func metadataMapper(body []byte) (common.ObjectMetadata, error) {
	metadata := common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	var response Data

	err := json.Unmarshal(body, &response)
	if err != nil {
		return metadata, err
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
