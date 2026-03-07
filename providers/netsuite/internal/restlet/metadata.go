package restlet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

func (a *Adapter) buildObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	payload := schemaRequest{
		Action: "getschema",
		Type:   objectName,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.restletURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (a *Adapter) parseObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	fullResp, err := common.UnmarshalJSON[restletResponse](resp)
	if err != nil {
		return nil, err
	}

	if fullResp.Header.Status != "SUCCESS" {
		return nil, parseRestletError(fullResp)
	}

	var schemaBody schemaResponseBody
	if err := json.Unmarshal(fullResp.Body, &schemaBody); err != nil {
		return nil, fmt.Errorf("failed to parse schema response: %w", err)
	}

	result := common.NewObjectMetadata(objectName, common.FieldsMetadata{})

	// Add body fields. Sublists are intentionally skipped — they represent
	// sub-objects (e.g. line items) rather than fields on the record itself.
	for fieldName, fieldInfo := range schemaBody.Fields {
		isRequired := fieldInfo.IsMandatory
		readOnly := fieldInfo.IsReadOnly

		displayName := fieldInfo.Label
		if displayName == "" {
			displayName = fieldName
		}

		result.AddFieldMetadata(fieldName, common.FieldMetadata{
			DisplayName:  displayName,
			ProviderType: fieldInfo.Type,
			IsRequired:   &isRequired,
			ReadOnly:     &readOnly,
		})
	}

	return result, nil
}
