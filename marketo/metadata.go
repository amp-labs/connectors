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

// ListObjectMetadata creates metadata of object via reading objects using Marketo API.
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames is not empty
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResult := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, obj := range objectNames {
		// Constructing the request url.
		url, err := c.getAPIURL(obj)
		if err != nil {
			return nil, err
		}

		resp, err := c.Client.Get(ctx, url.String())
		if err != nil {
			metadataResult.Errors[obj] = err

			continue
		}

		// Check nil response body, to avoid panic.
		if resp == nil || resp.Body == nil {
			metadataResult.Errors[obj] = common.ErrEmptyResponse

			continue
		}

		metadata, err := parseMetadataFromResponse(resp)
		if err != nil {
			if errors.Is(err, common.ErrMetadataLoadFailure) {
				metadataResult.Errors[obj] = common.ErrEmptyResponse
			} else {
				return nil, err
			}
		}

		metadata.DisplayName = obj
		metadataResult.Result[obj] = metadata
	}

	return &metadataResult, nil
}

func parseMetadataFromResponse(resp *common.JSONHTTPResponse) (common.ObjectMetadata, error) {
	var response responseObject

	bbytes := resp.Body.Source()
	if bbytes == nil {
		return common.ObjectMetadata{}, common.ErrEmptyResponse
	}

	metadata := common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	err := json.Unmarshal(bbytes, &response)
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
