package netsuite

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
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

func (c *Connector) parseObjectMetadataResponse(
	ctx context.Context,
	object string,
	resp *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	body, ok := resp.Body()
	if !ok || body == nil {
		return nil, errors.New("response missing body")
	}

	// Get the properties object which has all the fields.
	properties, err := jsonquery.New(body).ObjectRequired("properties")
	if err != nil {
		return nil, fmt.Errorf("getting properties object: %w", err)
	}

	fields, err := properties.GetObject()
	if err != nil {
		return nil, fmt.Errorf("getting properties object: %w", err)
	}

	result := &common.ObjectMetadata{
		Fields: make(map[string]common.FieldMetadata),
	}

	// Iterate through fields
	for property, metadata := range fields {
		// Skip links
		if strings.EqualFold(property, "links") {
			continue
		}

		// If the field is actually a related object and not a first class field, skip it.
		// If it doesn't have a type, skip it.
		ftype, err := metadata.GetKey("type")
		if err != nil {
			continue
		}

		if ftype.IsNull() || ftype.MustString() != "object" {
			continue
		}

		// Title can be missing, so we default to the property name.
		title, err := metadata.GetKey("title")
		if err != nil {
			title = ajson.StringNode(property, property)
		}

		// Format is more specific than a type, so we use it if it exists.
		// Ex: A field could have a date-time format but a string type.
		format, err := metadata.GetKey("format")
		if err != nil {
			format = ftype
		}

		result.Fields[property] = common.FieldMetadata{
			DisplayName:  title.MustString(),
			ProviderType: format.MustString(),
		}
	}

	return result, nil
}
