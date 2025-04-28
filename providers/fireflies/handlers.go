package fireflies

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Use introspection query to get field information
	query := fmt.Sprintf(`{
		__type(name: "%s") {
			name
			fields {
				name
				type {
					name
					kind
				}
			}
		}
	}`, naming.NewSingularString(naming.CapitalizeFirstLetterEveryWord(objectName)).String())

	// Create the request body as a map
	requestBody := map[string]string{
		"query": query,
	}

	// Marshal the request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	metadataResp, err := common.UnmarshalJSON[MetadataResponse](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(metadataResp.Data.Type.Fields) == 0 {
		return nil, fmt.Errorf(
			"missing or empty fields for object: %s, error: %w",
			objectName,
			common.ErrMissingExpectedValues,
		)
	}

	// Process each field from the introspection result
	for _, field := range metadataResp.Data.Type.Fields {
		objectMetadata.FieldsMap[field.Name] = field.Name
	}

	return &objectMetadata, nil
}
