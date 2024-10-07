package marketo

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/marketo/metadata"
)

type responseObject struct {
	Result  []map[string]any `json:"result"`
	Success bool             `json:"success"`
}

// ListObjectMetadata creates metadata of object via reading object using Marketo API.
// If it fails to retrieve the metadata, It retries using marketo's OpenAPI schema files.
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
		url, err := c.getAPIURL(obj)
		if err != nil {
			return nil, err
		}

		resp, err := c.Client.Get(ctx, url.String())
		if err != nil {
			runFallback(obj, &metadataResult)

			continue
		}

		if _, ok := resp.Body(); !ok {
			runFallback(obj, &metadataResult)

			continue
		}

		data, err := parseMetadataFromResponse(resp)
		if err != nil {
			if errors.Is(err, common.ErrMissingExpectedValues) {
				runFallback(obj, &metadataResult)

				continue
			} else {
				return nil, err
			}
		}

		data.DisplayName = obj
		metadataResult.Result[obj] = *data
	}

	return &metadataResult, nil
}

func parseMetadataFromResponse(resp *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[responseObject](resp)
	if err != nil {
		return nil, err
	}

	if len(response.Result) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	data := &common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	// Using the first result data to generate the metadata.
	for k := range response.Result[0] {
		data.FieldsMap[k] = k
	}

	return data, nil
}

func metadataFallback(objectName string) (*common.ObjectMetadata, error) {
	metadatResult, err := metadata.Schemas.Select([]string{objectName})
	if err != nil {
		return nil, err
	}

	data := metadatResult.Result[objectName]

	return &data, nil
}

func runFallback(obj string, res *common.ListObjectMetadataResult) *common.ListObjectMetadataResult { //nolint:unparam
	// Try fallback function
	data, err := metadataFallback(obj)
	if err != nil {
		res.Errors[obj] = err

		return res
	}

	data.DisplayName = obj
	res.Result[obj] = *data

	return res
}
