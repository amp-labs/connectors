// nolint
package attio

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
)

// This struct is used for when the response data having slice of data.
type responseObject struct {
	Result []map[string]any `json:"data"`
	// Other fields
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
		if obj == "self" {
			// Getting the metadata for self object in separate function
			metadata, err = parseMetadataForSingleObject(resp)
		} else {
			metadata, err = parseMetadataFromResponse(resp)
		}

		if err != nil {
			if errors.Is(err, common.ErrMissingExpectedValues) {
				continue
			} else {
				return nil, err
			}
		}

		metadata.DisplayName = obj
		metadataResult.Result[obj] = *metadata
	}

	return &metadataResult, nil
}

// This function is used for slice of objects
func parseMetadataFromResponse(resp *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[responseObject](resp)
	if err != nil {
		return nil, err
	}

	if len(response.Result) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	metadata := &common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	// Using the first result data to generate the metadata.
	for k := range response.Result[0] {
		metadata.FieldsMap[k] = k
	}

	return metadata, nil
}

// This function used for parse single object response
func parseMetadataForSingleObject(resp *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, common.ErrMissingExpectedValues
	}

	fieldsMap := map[string]string{}
	for k := range *(response) {
		fieldsMap[k] = k
	}

	return &common.ObjectMetadata{
		FieldsMap: fieldsMap,
	}, nil
}
