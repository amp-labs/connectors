package outreach

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/amp-labs/connectors/common"
)

type ObjectOKResponse struct {
	Data []map[string]any `json:"data"`
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
		objURL, err := url.JoinPath(c.BaseURL, obj)
		if err != nil {
			return nil, err
		}

		res, err := c.get(ctx, objURL)
		if err != nil {
			objMetadata.Errors[obj] = err

			continue
		}

		// Check nil response body, to avoid panic.
		if res == nil || res.Body == nil {
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

	var response ObjectOKResponse

	err := json.Unmarshal(body, &response)
	if err != nil {
		return metadata, err
	}

	for _, dataMap := range response.Data {
		for attr := range dataMap {
			metadata.FieldsMap[attr] = attr
		}
	}

	return metadata, nil
}
