package marketo

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/amp-labs/connectors/common"
)

type responseObject struct {
	Result  []map[string]any `json:"result"`
	Success bool             `json:"success"`
	// Other fields
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

		resp, err := c.Client.Get(ctx, url.String())
		if err != nil {
			objMetadata.Errors[obj] = err

			continue
		}

		// Check nil response body, to avoid panic.
		if resp == nil || resp.Body == nil {
			objMetadata.Errors[obj] = common.ErrEmptyResponse

			continue
		}

		metadata, err := metadataMapper(resp.Body.Source())
		if err != nil {
			if errors.Is(err, common.ErrMetadataLoadFailure) {
				objMetadata.Errors[obj] = common.ErrEmptyResponse
			} else {
				return nil, err
			}
		}

		metadata.DisplayName = obj
		objMetadata.Result[obj] = metadata
	}

	return &objMetadata, nil
}

func metadataMapper(body []byte) (common.ObjectMetadata, error) {
	var response responseObject

	metadata := common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	err := json.Unmarshal(body, &response)
	if err != nil {
		return metadata, err
	}

	if len(response.Result) == 0 {
		return metadata, common.ErrMetadataLoadFailure
	}

	// Using the first result data to generate the metadata.
	for k := range response.Result[0] {
		metadata.FieldsMap[k] = k
	}

	return metadata, nil
}
