package salesloft

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// Salesloft supports custom fields only for 3 objects.
// https://developers.salesloft.com/docs/api/custom-fields-create/
var objectsWithCustomFields = datautils.Map[string, string]{ // nolint:gochecknoglobals
	"person":      "people",
	"company":     "companies",
	"opportunity": "opportunities",
}

func (c *Connector) attachCustomMetadata(
	ctx context.Context, objectNames []string,
	metadataResult *common.ListObjectMetadataResult,
) (*common.ListObjectMetadataResult, error) {
	inputObjects := datautils.NewSetFromList(objectNames)
	supportedObjects := datautils.NewSetFromList(objectsWithCustomFields.Values())

	if len(supportedObjects.Intersection(inputObjects)) == 0 {
		// Requested objects do not support custom fields.
		// No-op.
		return metadataResult, nil
	}

	customFields, err := c.requestCustomFields(ctx)
	if err != nil {
		return nil, err
	}

	for objectName, fields := range customFields {
		// Attach fields to the object metadata.
		objectMetadata := metadataResult.GetObjectMetadata(objectName)
		if objectMetadata == nil {
			// object not found
			continue
		}

		for _, field := range fields {
			objectMetadata.AddFieldMetadata(field.Name, common.FieldMetadata{
				DisplayName:  field.Name,
				ValueType:    field.ValueType(),
				ProviderType: field.ProviderType,
				ReadOnly:     nil,
				IsCustom:     goutils.Pointer(true),
				IsRequired:   nil,
				Values:       nil,
			})
		}

		metadataResult.Result[objectName] = *objectMetadata
	}

	return metadataResult, nil
}

// requestCustomFields retrieves all custom fields defined in the system and
// groups them by object name.
//
// The returned map keys are object names ("person", "company", "opportunity),
// and each value is the list of custom fields associated with that object.
//
// This method:
//   - Calls the Salesloft Custom Fields API
//   - Uses the maximum supported page size (100)
//   - Transparently paginates until all fields are collected
func (c *Connector) requestCustomFields(
	ctx context.Context,
) (datautils.NamedLists[modelCustomField], error) {
	// Resolve the Custom Fields API endpoint.
	url, err := c.getURL("custom_fields")
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	// Salesloft custom fields endpoint supports a maximum of 100 items per page.
	// https://developers.salesloft.com/docs/api/custom-fields-index/
	url.WithQueryParam("per_page", "100")

	// Registry maps object name -> list of custom fields.
	registry := make(datautils.NamedLists[modelCustomField])

	// Paginate through all available pages and collect fields.
	for {
		res, err := c.Client.Get(ctx, url.String())
		if err != nil {
			return nil, errors.Join(common.ErrResolvingCustomFields, err)
		}

		fieldsResponse, err := common.UnmarshalJSON[modelCustomFieldsResponse](res)
		if err != nil {
			return nil, errors.Join(common.ErrResolvingCustomFields, err)
		}

		// Group fields by their associated object.
		for _, datum := range fieldsResponse.Data {
			registry.Add(datum.ObjectName(), datum)
		}

		// Advance pagination if a next page exists.
		nextPage := fieldsResponse.Metadata.Paging.NextPage
		if nextPage != nil {
			url.WithQueryParam("page", strconv.Itoa(*nextPage))
		} else {
			break
		}
	}

	return registry, nil
}

type modelCustomFieldsResponse struct {
	Metadata struct {
		Paging struct {
			PerPage     int  `json:"per_page"`
			CurrentPage int  `json:"current_page"`
			NextPage    *int `json:"next_page"`
			PrevPage    *int `json:"prev_page"`
		} `json:"paging"`
	} `json:"metadata"`
	Data []modelCustomField `json:"data"`
}

type modelCustomField struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	FieldType    string    `json:"field_type"`
	ProviderType string    `json:"value_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (f modelCustomField) ObjectName() string {
	return objectsWithCustomFields[f.FieldType]
}

func (f modelCustomField) ValueType() common.ValueType {
	switch f.ProviderType {
	case "text":
		return common.ValueTypeString
	case "date":
		return common.ValueTypeDateTime
	default:
		return common.ValueTypeOther
	}
}

func flattenCustomEmbed(node *ajson.Node) (map[string]any, error) {
	object, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	customFieldsResponse, err := jsonquery.ParseNode[customFieldReadResponse](node)
	if err != nil {
		return nil, err
	}

	// Attach custom fields on the top read object.
	for name, value := range customFieldsResponse.CustomFields {
		object[name] = value
	}

	return object, nil
}

type customFieldReadResponse struct {
	// Field name to Field Value mapping.
	CustomFields map[string]string `json:"custom_fields"`
}
