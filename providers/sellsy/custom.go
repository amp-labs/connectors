package sellsy

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Object names used in custom field queries differ from regular API usage.
// Keys are the relatedObjects as they appear in custom fields response.
// Values are object names.
// https://docs.sellsy.com/api/v2/#operation/search-custom-fields
var customFieldObjects = datautils.Map[string, string]{ // nolint:gochecknoglobals
	"client":      "clients",
	"contact":     "contacts",
	"document":    "documents",
	"item":        "items",
	"opportunity": "opportunities",
	"staff":       "staffs",
	"task":        "tasks",
}

// https://docs.sellsy.com/api/v2/#operation/search-custom-fields
func (c *Connector) fetchCustomFieldDefinitions( // nolint:cyclop
	ctx context.Context, objectNames []string,
) (map[string]customFieldDefinitions, error) {
	url, err := c.getCustomFieldsURL()
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", defaultPageSize)

	// Filter requested objects to only those that support custom fields.
	qualifiedObjectNames := datautils.NewSet(customFieldObjects.Values()...)
	requestedObjects := datautils.NewStringSet()

	for _, name := range objectNames {
		if qualifiedObjectNames.Has(name) {
			requestedObjects.AddOne(name)
		}
	}

	if len(requestedObjects) == 0 {
		// No need to make an API call. The provided objects do not support custom fields.
		return map[string]customFieldDefinitions{}, nil
	}

	// Init map of object names to custom fields.
	result := make(map[string]customFieldDefinitions)
	for name := range requestedObjects {
		result[name] = make(customFieldDefinitions, 0)
	}

	// Read custom fields page by page.
	for {
		res, err := c.JSONHTTPClient().Get(ctx, url.String())
		if err != nil {
			return nil, err
		}

		response, err := common.UnmarshalJSON[customFieldDefinitionResponse](res)
		if err != nil {
			return nil, err
		}

		if response == nil || len(response.Data) == 0 {
			// There are no items, stop reading pages.
			break
		}

		for _, item := range response.Data {
			for _, relatedObject := range item.RelatedObjects {
				objectName, ok := customFieldObjects[relatedObject]
				if ok && requestedObjects.Has(objectName) {
					result[objectName] = append(result[objectName], item)
				}
			}
		}

		if response.Pagination.Count < response.Pagination.Limit {
			// This page has only several records. Next page should be empty.
			break
		}

		// Advance offset tag to the next page.
		url.WithQueryParam("offset", response.Pagination.Offset)
	}

	return result, nil
}

type customFieldDefinitions []customFieldDefinition

func (d customFieldDefinitions) getIDs() []string {
	result := make([]string, len(d))
	for index, definition := range d {
		result[index] = strconv.Itoa(definition.Id)
	}

	return result
}

type customFieldDefinitionResponse struct {
	Data       []customFieldDefinition `json:"data"`
	Pagination struct {
		Limit  int    `json:"limit"`
		Count  int    `json:"count"`
		Offset string `json:"offset"`
	} `json:"pagination"`
}

type customFieldDefinition struct {
	Type       string `json:"type"`
	Parameters struct {
		Items []struct {
			Id    int    `json:"id"`
			Label string `json:"label"`
		} `json:"items,omitempty"`
	} `json:"parameters"`
	Id             int      `json:"id"`
	Name           string   `json:"name"`
	Code           string   `json:"code"`
	RelatedObjects []string `json:"related_objects"`
}

func (d customFieldDefinition) getValueType() common.ValueType {
	switch d.Type {
	case "numeric":
		return common.ValueTypeInt
	case "simple-text":
		return common.ValueTypeString
	case "boolean":
		return common.ValueTypeBoolean
	case "date":
		return common.ValueTypeDate
	case "radio":
		return common.ValueTypeSingleSelect
	case "checkbox":
		return common.ValueTypeMultiSelect
	default:
		return common.ValueTypeOther
	}
}

func (d customFieldDefinition) getValues() []common.FieldValue {
	result := make([]common.FieldValue, 0)

	for _, item := range d.Parameters.Items {
		result = append(result, common.FieldValue{
			Value:        strconv.Itoa(item.Id),
			DisplayValue: item.Label,
		})
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func (f *customFieldDefinitionResponse) GetIDs() []string {
	if f == nil {
		return nil
	}

	identifiers := make([]string, len(f.Data))
	for index, item := range f.Data {
		identifiers[index] = strconv.Itoa(item.Id)
	}

	return identifiers
}

type customFieldReadResponse struct {
	// ...
	// ... object specific fields
	// ...
	Embed struct {
		CustomFields []customFieldRead `json:"custom_fields"`
	} `json:"_embed"`
}

type customFieldRead struct {
	Value      any    `json:"value"`
	Type       string `json:"type"`
	Parameters struct {
		Items []struct {
			Id      int    `json:"id"`
			Label   string `json:"label"`
			Checked bool   `json:"checked"`
			Rank    int    `json:"rank"`
		} `json:"items,omitempty"`
		Min          any `json:"min"`
		Max          any `json:"max"`
		DefaultValue any `json:"default_value"`
		MinValue     any `json:"min_value"`
		MaxValue     any `json:"max_value"`
	} `json:"parameters"`
	Id               int      `json:"id"`
	Name             string   `json:"name"`
	Code             string   `json:"code"`
	Description      string   `json:"description"`
	Mandatory        bool     `json:"mandatory"`
	Rank             int      `json:"rank"`
	RelatedObjects   []string `json:"related_objects"`
	ShowOnPdf        bool     `json:"show_on_pdf"`
	CustomfieldGroup struct {
		Id            int    `json:"id"`
		Name          string `json:"name"`
		Code          string `json:"code"`
		OpenByDefault bool   `json:"open_by_default"`
	} `json:"customfield_group"`
}
