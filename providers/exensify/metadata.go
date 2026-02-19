package exensify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

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
		metadata, err := c.fetchObjectMetadata(ctx, obj)
		if err != nil {
			metadataResult.Errors[obj] = err
		} else {
			metadataResult.Result[obj] = *metadata
		}
	}

	return &metadataResult, nil
}

func (c *Connector) fetchObjectMetadata(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {

	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	reqURL, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, err
	}

	body, err := buildReadBody(objectName)
	if err != nil {
		return nil, err
	}

	newForm := url.Values{}

	newForm.Set("requestJobDescription", body)

	ecoded := newForm.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL.String(), bytes.NewBufferString(ecoded))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient().Client.Do(req)
	if err != nil {
		logging.Logger(ctx).Error("failed to get metadata", "object", objectName, "err", err.Error())

		return nil, fmt.Errorf("failed to get metadata for object %s: %w", objectName, err)
	}

	defer resp.Body.Close()

	var jsonBody map[string]any

	err = json.Unmarshal(common.GetResponseBodyOnce(resp), &jsonBody)
	if err != nil {
		logging.Logger(ctx).Error("failed to unmarshal metadata response", "object", objectName, "err", err.Error())

		return nil, fmt.Errorf("failed to unmarshal metadata response for object %s: %w", objectName, err)
	}

	records, ok := jsonBody["policyList"].([]any)
	if !ok {
		return nil, fmt.Errorf("failed to parse metadata response for object %s: %w", objectName, common.ErrMissingExpectedValues)
	}

	if len(records) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
	}

	for field, value := range firstRecord {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    inferValueTypeFromData(value),
			ProviderType: "", // not available
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}
