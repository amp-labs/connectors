package attio

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// This struct is used for when the response data having slice of data.
type responseObject struct {
	Data []map[string]any `json:"data"`
}

// ListObjectMetadata creates metadata of object via reading objects using Attio API.
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResult := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, obj := range objectNames {
		// Constructing the request url.
		url, err := c.getApiURL(obj)
		if err != nil {
			return nil, err
		}

		resp, err := c.Client.Get(ctx, url.String())
		if err != nil {
			metadataResult.Errors[obj] = err

			continue
		}

		var metadata *common.ObjectMetadata

		metadata, err = parseMetadataFromResponse(resp)
		if err != nil {
			metadataResult.Errors[obj] = err

			continue
		}

		metadata.DisplayName = obj
		metadataResult.Result[obj] = *metadata
	}

	return &metadataResult, nil
}

// This function is used for slice of objects.
func parseMetadataFromResponse(resp *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[responseObject](resp)
	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	metadata := &common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	// Using the first result data to generate the metadata.
	for k := range response.Data[0] {
		metadata.FieldsMap[k] = k
	}

	return metadata, nil
}
