package attio

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// This struct is used for when the response data having slice of data.
type responseObject struct {
	Result []map[string]any `json:"data"`
	// Other fields
}

var errCannotLoadMetadata = errors.New("cannot load metadata")

// ListObjectMetadata creates metadata of object via reading objects using Attio API.
// If it fails to fretrieve the metadata, It tried using static schema file in metadata dir.
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
			metadata, err = parseMetadataForSelfObject(resp)
		} else {
			metadata, err = parseMetadataFromResponse(resp)
		}
		if err != nil {
			if errors.Is(err, errCannotLoadMetadata) {
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

func parseMetadataFromResponse(resp *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[responseObject](resp)
	if err != nil {
		return nil, err
	}

	if len(response.Result) == 0 {
		return nil, errCannotLoadMetadata
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

func parseMetadataForSelfObject(resp *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	body, ok := resp.Body()
	if !ok {
		return nil, errCannotLoadMetadata
	}
	bodyBytes, err := ajson.Marshal(body)
	if err != nil || len(bodyBytes) == 0 {
		return nil, errCannotLoadMetadata
	}
	// Parse the JSON bytes into a map[string]interface{}
	var responseMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &responseMap); err != nil {
		return nil, err
	}
	fieldsMap := map[string]string{}
	for k := range responseMap {
		fieldsMap[k] = k
	}

	return &common.ObjectMetadata{
		FieldsMap: fieldsMap,
	}, nil

}
