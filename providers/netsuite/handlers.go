package netsuite

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildObjectMetadataRequest(ctx context.Context, object string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "services/rest/record", apiVersion, "metadata-catalog", object)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	// This tells the netsuite API to give us a shortened JSON schema for the properties.
	// Passing in schema+json, swagger+json, etc, will give us more information, but it's not
	// needed right now.
	req.Header.Add("Accept", "application/schema+json")

	return req, nil
}

// Example of a raw response is in test/netsuite/metadata/example.json.
func (c *Connector) parseObjectMetadataResponse(
	ctx context.Context,
	object string,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	metadata, err := common.UnmarshalJSON[metadataResponse](resp)
	if err != nil {
		return nil, err
	}

	if metadata == nil || metadata.Type != "object" || metadata.Properties == nil {
		return nil, fmt.Errorf("%w: invalid metadata response: %+v", common.ErrMissingExpectedValues, metadata)
	}

	result := &common.ObjectMetadata{
		DisplayName: object,
		Fields:      make(map[string]common.FieldMetadata),
		FieldsMap:   make(map[string]string),
	}

	for field, metadata := range metadata.Properties {
		if metadata.Type == "object" {
			continue // Skip nested associated objects
		}

		title := field
		if metadata.Title != "" {
			title = metadata.Title
		}

		format := metadata.Type
		if metadata.Format != "" {
			format = metadata.Format
		}

		result.Fields[field] = common.FieldMetadata{
			DisplayName:  title,
			ProviderType: format,
		}

		// Deprecated: this map includes only display names.
		result.FieldsMap[field] = title
	}

	return result, nil
}
