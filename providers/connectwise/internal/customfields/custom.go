package customfields

import (
	"context"
	"fmt"
	"maps"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/connectwise/internal/batch"
	"github.com/amp-labs/connectors/providers/connectwise/internal/core"
	"github.com/spyzhov/ajson"
)

func EnhanceObjectMetadata(ctx context.Context,
	base *core.Base,
	objectName string,
	objectMetadata *common.ObjectMetadata,
) error {
	fields, err := requestCustomFields(ctx, base, objectName)
	if err != nil {
		return err
	}

	for _, field := range fields {
		fieldMetadata := common.FieldMetadata{
			DisplayName:  field.Caption,
			ValueType:    field.getValueType(),
			ProviderType: field.getProviderType(),
			ReadOnly:     new(field.ReadOnlyFlag),
			IsCustom:     new(true),
			IsRequired:   new(field.RequiredFlag),
			Values:       field.getValues(),
		}

		objectMetadata.AddFieldMetadata(field.makeFieldName(), fieldMetadata)
	}

	return nil
}

// requestCustomFields fetches custom field definitions for objectName.
// It samples one record to collect custom field IDs, then loads the matching
// definitions and returns them indexed by ID.
func requestCustomFields(ctx context.Context,
	base *core.Base,
	objectName string,
) (map[string]modelCustomField, error) {
	if !supportedObjects.Has(objectName) {
		// This object cannot have custom fields, we are done.
		return map[string]modelCustomField{}, nil
	}

	customFieldIds, err := getCustomFieldIdsFromSampleRecord(ctx, base, objectName)
	if err != nil {
		return nil, err
	}

	if len(customFieldIds) == 0 {
		// This object does not have custom fields, we are done.
		return map[string]modelCustomField{}, nil
	}

	customFieldDefinitions, err := batch.Read[modelCustomField](
		ctx, base, "userDefinedFields", customFieldIds)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]modelCustomField)
	for _, def := range customFieldDefinitions {
		fields[strconv.Itoa(def.Id)] = def
	}

	return fields, nil
}

func EnhanceReadRecord(node *ajson.Node, root map[string]any) error {
	customFields, err := jsonquery.New(node).ArrayOptional("customFields")
	if err != nil {
		return err
	}

	if len(customFields) == 0 {
		// This object doesn't have custom fields defined.
		// Nothing to attach.
		return nil
	}

	fields := make(map[string]any)

	for _, customField := range customFields {
		field, err := jsonquery.ParseNode[readCustomField](customField)
		if err != nil {
			return err
		}

		fields[field.makeFieldName()] = field.Value
	}

	// Move custom fields to the top root level.
	maps.Copy(root, fields)

	return nil
}

// TransformUpdatePayload normalizes JSON Patch payloads for objects that
// support custom fields.
//
// ConnectWise requires that PATCH requests for such objects include a
// `/customFields` replace operation; otherwise the API may return 400 Bad
// Request, even if custom fields are not being modified.
//
// This function:
//   - Leaves payloads unchanged for objects that do not support custom fields.
//   - For supported objects:
//   - Extracts individual patch operations targeting custom field paths
//     (e.g. `/customField1`, `/customField2`) and converts them into entries
//     in a `/customFields` array of the form [{id, value}, ...].
//   - Preserves all non-custom-field operations as-is.
//   - Appends a single `/customFields` replace operation with the constructed
//     array.
//
// Notes:
//   - Only explicitly mentioned custom fields are included in the array.
//   - Omitting a custom field from the array does not clear its value.
//   - Sending an empty `customFields` array does not clear existing values.
func TransformUpdatePayload(objectName string,
	payloads []core.PatchOperationPayload,
) []core.PatchOperationPayload {
	if !supportedObjects.Has(objectName) {
		// Object does not support custom fields; return payload unchanged.
		return payloads
	}

	customFields := make([]map[string]any, 0)
	result := make([]core.PatchOperationPayload, 0)

	for _, payload := range payloads {
		// Normalize path by stripping leading slash.
		path, _ := strings.CutPrefix(payload.Path, "/")

		fieldIdStr, ok := strings.CutPrefix(path, "customField")
		if !ok {
			// Not a custom field operation; preserve as-is.
			result = append(result, payload)

			continue
		}

		fieldId, err := strconv.Atoi(fieldIdStr)
		if err != nil {
			// Malformed custom field ID; preserve operation as-is.
			result = append(result, payload)

			continue
		}

		customFields = append(customFields, map[string]any{
			"id":    fieldId,
			"value": payload.Value,
		})
	}

	// Always emit a /customFields replace operation for supported objects.
	result = append(result, core.PatchOperationPayload{
		Op:    "replace",
		Path:  "/customFields",
		Value: customFields,
	})

	return result
}

func TransformCreatePayload(record common.Record) {
	if record == nil {
		// No-op. Nothing to modify.
		return
	}

	customFields := make([]map[string]any, 0)

	for key, value := range record {
		if key == "customFields" {
			continue
		}

		if fieldIdStr, ok := strings.CutPrefix(key, "customField"); ok {
			fieldId, err := strconv.Atoi(fieldIdStr)
			if err != nil {
				continue
			}

			// This is in-house field, shouldn't be part of actual payload.
			delete(record, key)

			customFields = append(customFields, map[string]any{
				"id":    fieldId,
				"value": value,
			})
		}
	}

	if len(customFields) != 0 {
		record["customFields"] = customFields
	}
}

// supportedObjects lists ConnectWise object names that can include customFields in read responses.
// From the ConnectWise documentation:
// > Wondering if an endpoint supports custom fields?
// > Supported endpoints will have customFields(CustomFieldValue[])
// > listed at the end of the documentation.
var supportedObjects = datautils.NewSet( // nolint:gochecknoglobals
	"activities",
	"adjustments",
	"agreements",
	"calculateSla",
	"catalog",
	"companies",
	"configurations",
	"contacts",
	"expense/entries",
	"finance/companyFinance",
	"invoices",
	"orders",
	"products",
	"project/tickets",
	"projects",
	"purchaseorders",
	"rmaTags",
	"sales/opportunities",
	"service/tickets",
	"system/members",
	"system/membertemplates",
	"time/entries",
	"withSso",
)

// getCustomFieldIdsFromSampleRecord reads one record for objectName and returns
// the IDs of the custom fields present in that response.
// It is guaranteed that all custom fields will be returned even if they are empty and set to null.
func getCustomFieldIdsFromSampleRecord(ctx context.Context, base *core.Base, objectName string) ([]string, error) {
	url, err := base.GetUrl(objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("pageSize", "1")

	res, err := base.JSONHTTPClient().Get(ctx, url.String(), base.ClientIdHeader())
	if err != nil {
		return nil, err
	}

	records, err := common.UnmarshalJSON[readResponses](res)
	if err != nil {
		return nil, err
	}

	if len(*records) == 0 {
		// There are no records for sampling.
		return []string{}, nil
	}

	record := (*records)[0]

	ids := make([]string, len(record.CustomFields))
	for i, field := range record.CustomFields {
		ids[i] = strconv.Itoa(field.Id)
	}

	return ids, nil
}

// readResponses represents the response shape returned by object read endpoints.
// Only the customFields field is modeled here.
type readResponses []readResponse

// readResponse represents a single object read result.
type readResponse struct {
	CustomFields []readCustomField `json:"customFields"`
}

// readCustomField models the customFields item embedded in a read response.
type readCustomField struct {
	Id                    int    `json:"id"`
	Caption               string `json:"caption"`
	Type                  string `json:"type"`
	EntryMethod           string `json:"entryMethod"`
	NumberOfDecimals      int    `json:"numberOfDecimals"`
	Value                 any    `json:"value"`
	ConnectWiseId         string `json:"connectWiseId"`
	RowNum                int    `json:"rowNum"`
	UserDefinedFieldRecId int    `json:"userDefinedFieldRecId"`
	PodId                 string `json:"podId"`
}

// modelCustomField represents a custom field definition returned by ConnectWise.
// Endpoint: "/system/userDefinedFields".
type modelCustomField struct {
	Id                  int    `json:"id"`
	PodId               int    `json:"podId"`
	PodDescription      string `json:"podDescription"`
	Caption             string `json:"caption"`
	SequenceNumber      int    `json:"sequenceNumber"`
	ScreenId            string `json:"screenId"`
	ScreenDescription   string `json:"screenDescription"`
	FieldTypeIdentifier string `json:"fieldTypeIdentifier"`
	NumberDecimals      int    `json:"numberDecimals"`
	EntryTypeIdentifier string `json:"entryTypeIdentifier"`
	RequiredFlag        bool   `json:"requiredFlag"`
	DisplayOnScreenFlag bool   `json:"displayOnScreenFlag"`
	ReadOnlyFlag        bool   `json:"readOnlyFlag"`
	ListViewFlag        bool   `json:"listViewFlag"`
	Options             []struct {
		Id           int    `json:"id"`
		OptionValue  string `json:"optionValue"`
		DefaultFlag  bool   `json:"defaultFlag"`
		InactiveFlag bool   `json:"inactiveFlag"`
		SortOrder    int    `json:"sortOrder"`
	} `json:"options"`
	BusinessUnitIds     []any     `json:"businessUnitIds"`
	LocationIds         []any     `json:"locationIds"`
	ConnectWiseID       string    `json:"connectWiseID"`
	DisplayScreenInASIO bool      `json:"displayScreenInASIO"`
	DateCreated         time.Time `json:"dateCreated"`
	Info                any       `json:"_info"`
}

// getValueType returns the common.ValueType corresponding to the custom field.
func (f modelCustomField) getValueType() common.ValueType {
	switch f.FieldTypeIdentifier {
	case "Button":
		return common.ValueTypeOther
	case "Checkbox":
		return common.ValueTypeBoolean
	case "Date":
		return common.ValueTypeDate
	case "Hyperlink":
		return common.ValueTypeString
	case "TextArea":
		return common.ValueTypeString
	case "Number":
		return selectValueOrDefault(f, common.ValueTypeFloat)
	case "Percent":
		return selectValueOrDefault(f, common.ValueTypeFloat)
	case "Text":
		return selectValueOrDefault(f, common.ValueTypeString)
	default:
		return common.ValueTypeOther
	}
}

// selectValueOrDefault returns SingleSelect for list-style fields, and defaultValue otherwise.
// Number, Percent, Text can all be either primitive data type or Single Select.
// Both entry methods hold exactly one value: "List" is a dropdown pick-list, "Option" is radio buttons.
// ConnectWise user-defined fields have no multi-select entry method; values are always scalar.
func selectValueOrDefault(field modelCustomField, defaultValue common.ValueType) common.ValueType {
	switch field.EntryTypeIdentifier {
	case "List", "Option":
		return common.ValueTypeSingleSelect
	default:
		return defaultValue
	}
}

func (f modelCustomField) getProviderType() string {
	return f.EntryTypeIdentifier + "_" + f.FieldTypeIdentifier
}

func (f modelCustomField) getValues() []common.FieldValue {
	values := make(common.FieldValues, len(f.Options))
	for i, option := range f.Options {
		values[i] = common.FieldValue{
			Value:        option.OptionValue,
			DisplayValue: option.OptionValue,
		}
	}

	if len(values) == 0 {
		return nil
	}

	return values
}

func (f modelCustomField) makeFieldName() string {
	return fmt.Sprintf("customField%v", strconv.Itoa(f.Id))
}

func (f readCustomField) makeFieldName() string {
	return fmt.Sprintf("customField%v", strconv.Itoa(f.Id))
}
