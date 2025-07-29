package suiteql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// SuiteQL has no metadata endpoint, so we sample 1 record from the read endpoint.
func (a *Adapter) buildObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, "suiteql")
	if err != nil {
		return nil, err
	}

	body := suiteQLQueryBody{
		Query: fmt.Sprintf("SELECT * FROM %s", objectName),
	}

	url.WithQueryParam("limit", "1")

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Prefer", "transient")

	fmt.Println(url.String())

	return req, nil
}

// parseObjectMetadataResponse parses the metadata response from SuiteQL.
func (a *Adapter) parseObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[suiteQLResponse](resp)
	if err != nil {
		return nil, err
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("no records found for metadata request: %w", common.ErrNotFound)
	}

	record := response.Items[0]

	fields := make(map[string]common.FieldMetadata)
	fieldsMap := make(map[string]string)

	// TODO: It is possible that SuiteQL doesn't return all fields (custom fields, etc.)
	// We should verify this in the future.
	for field := range record {
		fields[field] = common.FieldMetadata{DisplayName: field}
		fieldsMap[field] = field
	}

	metadata := &common.ObjectMetadata{
		DisplayName: objectName,
		Fields:      fields,
		FieldsMap:   fieldsMap,
	}

	return metadata, nil
}
