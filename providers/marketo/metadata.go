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

// ListObjectMetadata returns metadata for an object by sampling an object from Marketo's API.
// If that fails, it generates object metadata by parsing Marketo's OpenAPI files.
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
		url, err := c.constructMetadataURL(obj)
		if err != nil {
			return nil, err
		}

		resp, err := c.JSONHTTPClient().Get(ctx, url.String())
		if err != nil {
			runFallback(c.Module(), obj, &metadataResult)

			continue
		}

		if _, ok := resp.Body(); !ok {
			runFallback(c.Module(), obj, &metadataResult)

			continue
		}

		data, err := parseMetadataFromResponse(resp, obj)
		if err != nil {
			if errors.Is(err, common.ErrMissingExpectedValues) {
				runFallback(c.Module(), obj, &metadataResult)

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

func parseMetadataFromResponse(resp *common.JSONHTTPResponse, objectName string) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[responseObject](resp)
	if err != nil {
		return nil, err
	}

	if _, ok := hasMetadataResource(objectName); ok {
		return parseDescribeResponse(response.Result[0])
	}

	if len(response.Result) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	data := common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	// Using the first result data to generate the metadata.
	for k := range response.Result[0] {
		data.FieldsMap[k] = k
	}

	return &data, nil
}

func metadataFallback(moduleID common.ModuleID, objectName string) (*common.ObjectMetadata, error) {
	metadatResult, err := metadata.Schemas.Select(moduleID, []string{objectName})
	if err != nil {
		return nil, err
	}

	data := metadatResult.Result[objectName]

	return &data, nil
}

func runFallback(
	moduleID common.ModuleID, obj string, res *common.ListObjectMetadataResult,
) *common.ListObjectMetadataResult { //nolint:unparam
	// Try fallback function
	data, err := metadataFallback(moduleID, obj)
	if err != nil {
		res.Errors[obj] = err

		return res
	}

	data.DisplayName = obj
	res.Result[obj] = *data

	return res
}

func parseDescribeResponse(results any) (*common.ObjectMetadata, error) {
	data := common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	fieldsResult, ok := results.(map[string]any)
	if !ok {
		return nil, ErrFailedConvertFields
	}

	fields := fieldsResult["fields"]

	result, ok := fields.([]any)
	if !ok {
		return nil, ErrFailedConvertFields
	}

	for _, flds := range result {
		flds, ok := flds.(map[string]any)
		if !ok {
			return nil, ErrFailedConvertFields
		}

		fld, ok := flds["name"].(string)
		if !ok {
			return nil, ErrFailedConvertFields
		}

		data.FieldsMap[fld] = fld
	}

	return &data, nil
}
