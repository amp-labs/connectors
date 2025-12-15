package salesforce

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/amp-labs/amp-common/jsonpath"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
)

func (c *Connector) UpsertMetadata(
	ctx context.Context, params *common.UpsertMetadataParams,
) (*common.UpsertMetadataResult, error) {
	// Delegated.
	return c.customAdapter.UpsertMetadata(ctx, params)
}

// ListObjectMetadata returns object metadata for each object name provided.
func (c *Connector) ListObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	if c.isPardotModule() {
		return c.pardotAdapter.ListObjectMetadata(ctx, objectNames)
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

// See https://developer.salesforce.com/docs/atlas.en-us.api.meta/api/sforce_api_calls_describesobjects_describesobjectresult.htm#field.
//
//nolint:lll
type fieldResult struct {
	// Field name used in API calls, such as create(), delete(), and query().
	Name        string `json:"name"`
	DisplayName string `json:"label"`

	// https://developer.salesforce.com/docs/atlas.en-us.244.0.api.meta/api/sforce_api_calls_describesobjects_describesobjectresult.htm#FieldType
	Type string `json:"type"`

	PicklistValues []picklistValue `json:"picklistValues"`

	Autonumber        *bool `json:"autonumber,omitempty"`
	Calculated        *bool `json:"calculated,omitempty"`
	Createable        *bool `json:"createable,omitempty"`
	Updateable        *bool `json:"updateable,omitempty"`
	Custom            *bool `json:"custom,omitempty"`
	Nillable          *bool `json:"nillable,omitempty"`
	DefaultedOnCreate *bool `json:"defaultedOnCreate,omitempty"`

	// CompoundFieldName is the name of the parent compound field if this field is a component.
	// For example, "BillingStreet" has CompoundFieldName "BillingAddress".
	// Null/empty for non-component fields.
	// See: https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/compound_fields.htm
	CompoundFieldName *string `json:"compoundFieldName,omitempty"`
}

type picklistValue struct {
	DisplayName string `json:"label"`
	Value       string `json:"value"`
}

func (r describeSObjectResult) transformToFields() map[string]common.FieldMetadata {
	fieldsMap := make(map[string]common.FieldMetadata)

	// First pass: add all fields with their original names as flat fields.
	// Even if they are components of a compound field.
	for _, field := range r.Fields {
		fieldName := strings.ToLower(field.Name)
		fieldsMap[fieldName] = field.transformToFieldMetadata()
	}

	// Second pass: add nested fields using bracket notation.
	// Fields with a CompoundFieldName are components of a compound field (e.g., BillingAddress).
	// We add them as nested fields alongside the flat fields: $['compoundfield']['component']
	for _, field := range r.Fields {
		if field.CompoundFieldName == nil || *field.CompoundFieldName == "" {
			continue
		}

		parentName := strings.ToLower(*field.CompoundFieldName)
		childName := strings.ToLower(field.Name)
		path := jsonpath.ToNestedPath(parentName, childName)

		fieldsMap[path] = field.transformToFieldMetadata()
	}

	return fieldsMap
}

// See https://developer.salesforce.com/docs/atlas.en-us.api.meta/api/sforce_api_calls_describesobjects_describesobjectresult.htm#field
//
// Salesforce doesn't have a native concept of "read-only" fields, so we use some other
// fields to determine if a field is read-only.
//
//nolint:lll
func (f fieldResult) isReadOnly() bool {
	return (f.Autonumber != nil && *f.Autonumber) ||
		(f.Calculated != nil && *f.Calculated) ||
		(f.Createable != nil && !*f.Createable && f.Updateable != nil && !*f.Updateable)
}

func (f fieldResult) transformToFieldMetadata() common.FieldMetadata {
	var (
		valueType common.ValueType
		values    []common.FieldValue
	)

	// Based on type property map value to Ampersand value type.
	switch f.Type {
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
		values = f.getFieldValues()
	case "multipicklist":
		valueType = common.ValueTypeMultiSelect
		values = f.getFieldValues()
	default:
		// Examples: base64, ID, reference, currency, percent, phone, url, email, anyType, location
		valueType = common.ValueTypeOther
	}

	return common.FieldMetadata{
		DisplayName:  f.DisplayName,
		ValueType:    valueType,
		ProviderType: f.Type,
		ReadOnly:     goutils.Pointer(f.isReadOnly()),
		IsCustom:     f.Custom,
		IsRequired:   f.isRequired(),
		Values:       values,
	}
}

func (f fieldResult) getFieldValues() []common.FieldValue {
	result := make([]common.FieldValue, len(f.PicklistValues))
	for index, option := range f.PicklistValues {
		result[index] = common.FieldValue{
			Value:        option.Value,
			DisplayValue: option.DisplayName,
		}
	}

	return result
}

// isRequired returns whether the field must be supplied when creating a record.
// Salesforce only defines "required" in the context of CREATE (not update).
//
// A field is required on create when all of the following are true:
//   - it is createable
//   - it is not nillable (cannot be null)
//   - it is not defaulted on create (Salesforce will not autopopulate it)
//
// nolint:lll
// Reference: https://salesforce.stackexchange.com/questions/260294/in-order-to-check-if-a-field-is-required-or-not-is-the-result-of-isnillable-met
func (f fieldResult) isRequired() *bool {
	// Platform-populated fields are never required inputs.
	if f.Autonumber != nil && *f.Autonumber || f.Calculated != nil && *f.Calculated {
		return goutils.Pointer(false)
	}

	// Cannot determine without all three metadata flags.
	if f.Createable == nil || f.Nillable == nil || f.DefaultedOnCreate == nil {
		return nil
	}

	// Required when createable, non-nillable, and not defaulted by Salesforce.
	requiredOnCreate := *f.Createable && !*f.Nillable && !*f.DefaultedOnCreate

	return goutils.Pointer(requiredOnCreate)
}
