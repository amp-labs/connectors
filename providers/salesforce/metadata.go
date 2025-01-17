package salesforce

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// ListObjectMetadata returns object metadata for each object name provided.
func (c *Connector) ListObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	requests := make([]compositeRequestItem, len(objectNames))

	// Construct describe requests for each object name
	for idx, objectName := range objectNames {
		describeObjectURL, err := c.getURIPartSobjectsDescribe(objectName)
		if err != nil {
			return nil, err
		}

		requests[idx] = compositeRequestItem{
			Method:      "GET",
			URL:         describeObjectURL.String(),
			ReferenceId: objectName,
		}
	}

	// Construct endpoint for the request
	compositeRequestEndpoint, err := c.getRestApiURL("composite")
	if err != nil {
		return nil, err
	}

	// Make the request
	result, err := c.Client.Post(
		ctx,
		compositeRequestEndpoint.String(),
		compositeRequest{
			CompositeRequest: requests,
			// If we fail to fetch metadata for one object, we don't want to fail the entire request.
			AllOrNone: false,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error fetching Salesforce fields: %w", err)
	}

	// Construct map of object names to object metadata
	return constructResponseMap(result)
}

// constructResponseMap constructs a map of object names to object metadata from the composite response.
func constructResponseMap(response *common.JSONHTTPResponse) (*common.ListObjectMetadataResult, error) {
	resp, err := common.UnmarshalJSON[compositeResponse](response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response from JSON: %w", err)
	}

	// Construct map of object names to object metadata
	objectsMap := common.NewListObjectMetadataResult()

	for _, subRes := range resp.CompositeResponse {
		result := &describeSObjectResult{}

		err = json.Unmarshal(subRes.Body, result)
		if err != nil {
			// If one of the sub-requests of the composite request fails, then subRes.Body will look like:
			// "[{\"errorCode\":\"NOT_FOUND\",\"message\":\"The requested resource does not exist\"}]"
			// which will fail the json.Unmarshall
			objectsMap.Errors[strings.ToLower(subRes.ReferenceId)] = fmt.Errorf(
				"%w: %s", ErrCannotReadMetadata, string(subRes.Body),
			)
		} else {
			objectsMap.Result[strings.ToLower(result.Name)] = *common.NewObjectMetadata(
				result.Label, result.transformToFields(),
			)
		}
	}

	return objectsMap, nil
}

type compositeRequest struct {
	AllOrNone        bool                   `json:"allOrNone"`
	CompositeRequest []compositeRequestItem `json:"compositeRequest"`
}

type compositeResponse struct {
	CompositeResponse []compositeResponseItem `json:"compositeResponse"`
}

type compositeRequestItem struct {
	// ReferenceId allows us to map the result to the original request
	ReferenceId string `json:"referenceId"`
	Method      string `json:"method"`
	URL         string `json:"url"`
	Body        any    `json:"body,omitempty"`
}

type compositeResponseItem struct {
	// ReferenceId comes from the original request
	ReferenceId    string            `json:"referenceId"`
	Body           json.RawMessage   `json:"body"`
	HttpHeaders    map[string]string `json:"httpHeaders"`    //nolint:revive
	HttpStatusCode int               `json:"httpStatusCode"` //nolint:revive
}

// See https://developer.salesforce.com/docs/atlas.en-us.244.0.api.meta/api/sforce_api_calls_describesobjects_describesobjectresult.htm.
// NOTE: doc page is for SOAP API, but REST API returns the same result.
//
//nolint:lll
type describeSObjectResult struct {
	Name   string        `json:"name"`
	Label  string        `json:"label"`
	Fields []fieldResult `json:"fields" validate:"required"`
}

// See https://developer.salesforce.com/docs/atlas.en-us.244.0.api.meta/api/sforce_api_calls_describesobjects_describesobjectresult.htm#field.
//
//nolint:lll
type fieldResult struct {
	// Field name used in API calls, such as create(), delete(), and query().
	Name        string `json:"name"`
	DisplayName string `json:"label"`

	// https://developer.salesforce.com/docs/atlas.en-us.244.0.api.meta/api/sforce_api_calls_describesobjects_describesobjectresult.htm#FieldType
	Type string `json:"type"`

	PicklistValues []picklistValue `json:"picklistValues"`
}

type picklistValue struct {
	DisplayName string `json:"label"`
	Value       string `json:"value"`
}

func (r describeSObjectResult) transformToFields() map[string]common.FieldMetadata {
	fieldsMap := make(map[string]common.FieldMetadata)

	for _, field := range r.Fields {
		fieldName := strings.ToLower(field.Name)
		fieldsMap[fieldName] = field.transformToFieldMetadata()
	}

	return fieldsMap
}

func (o fieldResult) transformToFieldMetadata() common.FieldMetadata {
	var (
		valueType common.ValueType
		values    []common.FieldValue
	)

	// Based on type property map value to Ampersand value type.
	switch o.Type {
	case "string", "textarea":
		valueType = common.ValueTypeString
	case "boolean":
		valueType = common.ValueTypeBoolean
	case "int":
		valueType = common.ValueTypeInt
	case "double":
		valueType = common.ValueTypeFloat
	case "date":
		valueType = common.ValueTypeDate
	case "datetime":
		valueType = common.ValueTypeDateTime
	case "picklist", "combobox":
		valueType = common.ValueTypeSingleSelect
		values = o.getFieldValues()
	case "multipicklist":
		valueType = common.ValueTypeMultiSelect
		values = o.getFieldValues()
	default:
		// Examples: base64, ID, reference, currency, percent, phone, url, email, anyType, location
		valueType = common.ValueTypeOther
	}

	return common.FieldMetadata{
		DisplayName:  o.DisplayName,
		ValueType:    valueType,
		ProviderType: o.Type,
		ReadOnly:     true,
		Values:       values,
	}
}

func (o fieldResult) getFieldValues() []common.FieldValue {
	result := make([]common.FieldValue, len(o.PicklistValues))
	for index, option := range o.PicklistValues {
		result[index] = common.FieldValue{
			Value:        option.Value,
			DisplayValue: option.DisplayName,
		}
	}

	return result
}
