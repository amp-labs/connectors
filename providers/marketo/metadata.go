package marketo

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
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
	ctx = logging.With(ctx, "connector", "marketo")

	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResult := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, obj := range objectNames {
		url, err := c.constructMetadataURL(obj)
		if err != nil {
			return nil, err
		}

		httpResp, body, err := c.Client.HTTPClient.Get(ctx, url.String())
		if err != nil {
			logging.Logger(ctx).Error("failed to get metadata", "object", obj, "body", body, "err", err.Error())
			runFallback(obj, &metadataResult)

			continue
		}

		defer httpResp.Body.Close()

		resp, err := common.ParseJSONResponse(httpResp, body)
		if err != nil {
			logging.Logger(ctx).Error("failed to parse metadata response", "object", obj, "body", body, "err", err.Error())
			runFallback(obj, &metadataResult)

			continue
		}

		if _, ok := resp.Body(); !ok {
			runFallback(obj, &metadataResult)

			continue
		}

		data, err := parseMetadataFromResponse(resp, obj)
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

func metadataFallback(objectName string) (*common.ObjectMetadata, error) {
	metadata, err := metadata.Schemas.SelectOne(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func runFallback(obj string, res *common.ListObjectMetadataResult,
) *common.ListObjectMetadataResult { //nolint:unparam
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
