package asana

import (
	"context"
	"log"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	data := common.NewListObjectMetadataResult()

	for _, objectName := range objectNames {
		url, err := c.getAPIURL(objectName)
		if err != nil {
			data.AppendError(objectName, err)

			continue
		}

		if supportLimitAndOffset.Has(objectName) {
			url.WithQueryParam("limit", "1")
		}

		recordId, err := c.fetchSingleRecord(ctx, url.String())
		if err != nil {
			log.Println("Error fetching record for object:", objectName, "error:", err)
			data.AppendError(objectName, err)

			continue
		}

		metadata, err := c.fetchObjectMetadata(ctx, objectName, recordId)
		if err != nil {
			data.AppendError(objectName, err)

			continue
		}

		data.Result[objectName] = *metadata
	}

	return data, nil
}

func (c *Connector) fetchSingleRecord(ctx context.Context, url string) (string, error) {
	result, err := c.JSONHTTPClient().Get(ctx, url)
	if err != nil {
		return "", err
	}

	res, err := common.UnmarshalJSON[map[string]any](result)
	if err != nil {
		return "", err
	}

	if res == nil || len(*res) == 0 {
		return "", common.ErrMissingExpectedValues
	}

	records, ok := (*res)["data"].([]any) //nolint:varnamelen
	if !ok {
		return "", common.ErrMissingExpectedValues
	}

	if len(records) == 0 {
		return "", common.ErrMissingExpectedValues
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return "", common.ErrMissingExpectedValues
	}

	// We just need record id to fetch the object metadata.
	recordId, ok := firstRecord["gid"].(string)
	if !ok {
		return "", common.ErrMissingExpectedValues
	}

	return recordId, nil
}

func (c *Connector) fetchObjectMetadata(
	ctx context.Context, objectName, recordId string,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	url, err := c.getAPIURL(objectName)
	if err != nil {
		return nil, err
	}

	url.AddPath(recordId)

	req, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	metadata, err := common.UnmarshalJSON[map[string]any](req)
	if err != nil {
		return nil, err
	}

	if metadata == nil || len(*metadata) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	record, ok := (*metadata)["data"].(map[string]any)
	if !ok {
		return nil, common.ErrMissingExpectedValues
	}

	for field, value := range record {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    inferValueTypeFromData(value),
			ProviderType: "", // not available
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func inferValueTypeFromData(value any) common.ValueType {
	if value == nil {
		return common.ValueTypeOther
	}

	switch value.(type) {
	case string:
		return common.ValueTypeString
	case float64, int, int64:
		return common.ValueTypeFloat
	case bool:
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}
