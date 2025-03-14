package attio

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// This struct is used for when the response data having slice of data.
type responseObject struct {
	Data []map[string]any `json:"data"`
}

// This struct is used for when the response data having single data.
type singleResponseObject struct {
	Data map[string]any `json:"data"`
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

	var (
		objName                    string
		isAttioStandardOrCustomObj bool
	)

	for _, obj := range objectNames {
		isAttioStandardOrCustomObj = false

		// Constructing the request url.
		if supportAttioApiObj.Has(obj) {
			objName = obj
		} else {
			isAttioStandardOrCustomObj = true
			objName = c.getObjectsURL(obj)
		}

		resp, err := getResponse(c, ctx, objName)

		if err != nil {
			metadataResult.Errors[obj] = err

			continue
		}

		var metadata *common.ObjectMetadata

		metadata, err = parseMetadataFromResponse(resp, isAttioStandardOrCustomObj)
		if err != nil {
			metadataResult.Errors[obj] = err

			continue
		}

		if isAttioStandardOrCustomObj {
			objName = c.getObjects(obj)
			res, err := getResponse(c, ctx, objName)
			if err != nil {
				metadataResult.Errors[obj] = err

				continue
			}

			displayName, err := getDisplayName(res)
			if err != nil {
				metadataResult.Errors[obj] = err

				continue
			}

			metadata.DisplayName = displayName

		} else {
			metadata.DisplayName = obj
		}

		metadataResult.Result[obj] = *metadata
	}

	return &metadataResult, nil
}

// Getting the response
func getResponse(c *Connector, ctx context.Context, objName string) (*common.JSONHTTPResponse, error) {
	url, err := c.getApiURL(objName)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func parseMetadataFromResponse(resp *common.JSONHTTPResponse, isAttioStandardOrCustomObj bool) (*common.ObjectMetadata, error) {
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

	// Retrieving metadata for standard and custom objects in Attio using the api_slug field.
	if isAttioStandardOrCustomObj {
		for _, value := range response.Data {
			for k, objData := range value {
				if k == "api_slug" {
					if strValue, ok := objData.(string); ok {
						metadata.FieldsMap[strValue] = strValue
					}
				}
			}
		}
	} else {
		// Using the first result data to generate the metadata.
		for k := range response.Data[0] {
			metadata.FieldsMap[k] = k
		}
	}

	return metadata, nil
}

// Getting display name for both standard and custom objects.
func getDisplayName(resp *common.JSONHTTPResponse) (string, error) {
	response, err := common.UnmarshalJSON[singleResponseObject](resp)
	if err != nil {
		return "", err
	}

	if len(response.Data) == 0 {
		return "", common.ErrMissingExpectedValues
	}

	for key, value := range response.Data {
		if key == "plural_noun" {
			return value.(string), nil
		}
	}

	return "", common.ErrNotFound
}
