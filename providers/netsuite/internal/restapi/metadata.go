package restapi

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (a *Adapter) buildObjectMetadataRequest(ctx context.Context, object string) (*http.Request, error) {
	url, err := urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, "metadata-catalog", object)
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

func (a *Adapter) parseObjectMetadataResponse(
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

	for name, prop := range metadata.Properties {
		var (
			displayName  = oneOf(prop.Title, name)
			providerType string
			values       []common.FieldValue
		)

		switch prop.Type {
		// This is a case where the field allows single/multiselect.
		// NetSuite puts the enum on the child "id" field.
		case "object":
			if idProp, ok := prop.Properties["id"]; ok {
				providerType = oneOf(idProp.Format, idProp.Type)
				values = make([]common.FieldValue, len(idProp.Enum))

				for i, v := range idProp.Enum {
					values[i] = common.FieldValue{
						Value:        v,
						DisplayValue: v,
					}
				}
			}
		// The usual fields (strings, booleans, etc.)
		default:
			providerType = oneOf(prop.Format, prop.Type)
		}

		result.Fields[name] = common.FieldMetadata{
			DisplayName:  displayName,
			ProviderType: providerType,
			Values:       values,
		}

		// backward compatibility
		result.FieldsMap[name] = displayName
	}

	return result, nil
}

func oneOf(candidates ...string) string {
	for _, s := range candidates {
		if s != "" {
			return s
		}
	}

	return ""
}

// metadataResponse is the response from the /metadata-catalog/{object} endpoint.
type metadataResponse struct {
	// The type of the object.
	Type string `json:"type"`

	// Contains key-value pairs of field names and their metadata.
	// Metadata can have title, type,description and nullable, even properties again.
	Properties map[string]fieldMetadata `json:"properties"`
}

type fieldMetadata struct {
	Title      string                     `json:"title"`
	Type       string                     `json:"type"`   // string, number, boolean, object, array, null
	Format     string                     `json:"format"` // date-time, date, etc.
	Properties map[string]fieldProperties `json:"properties"`
}

type fieldProperties struct {
	Type   string   `json:"type"`   // Type of the enum values
	Enum   []string `json:"enum"`   // Enum values
	Format string   `json:"format"` // date-time, date, etc.
}
