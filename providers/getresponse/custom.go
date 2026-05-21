package getresponse

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// Field keys for custom field values on contacts use this prefix and the
// GetResponse customFieldId (see also ListObjectMetadata merge).
const customFieldKeyPrefix = "cf_"

const objectContacts = "contacts"

// contactCustomFieldValue is one element of the customFieldValues array on GET /v3/contacts.
// https://apireference.getresponse.com
type contactCustomFieldValue struct {
	CustomFieldId string `json:"customFieldId"`
	Value         any    `json:"value"`
}

func (v contactCustomFieldValue) normalizedValue() any {
	return normalizeCustomFieldAPIValue(v.Value)
}

// getResponseCustomField is a subset of GET /v3/custom-fields items.
// https://apireference.getresponse.com
type getResponseCustomField struct {
	CustomFieldId string   `json:"customFieldId"`
	Name          string   `json:"name"`
	FieldType     string   `json:"fieldType"`
	ValueType     string   `json:"valueType"`
	Values        []string `json:"values"`
}

// CustomFieldKey returns the connector field name for a GetResponse customFieldId.
func CustomFieldKey(customFieldID string) string {
	return customFieldKeyPrefix + customFieldID
}

func (d getResponseCustomField) fieldMetadata() common.FieldMetadata {
	return common.FieldMetadata{
		DisplayName:  d.Name,
		ValueType:    d.mapValueType(),
		ProviderType: d.ValueType,
		Values:       d.selectValues(),
		IsCustom:     new(true),
	}
}

func (d getResponseCustomField) mapValueType() common.ValueType {
	switch strings.ToLower(d.ValueType) {
	case "string", "phone", "url", "ip", "country", "currency", "multi_line_text":
		return common.ValueTypeString
	case "number", "integer":
		return common.ValueTypeInt
	case "date", "datetime":
		return common.ValueTypeDate
	case "single_select", "singleselect":
		return common.ValueTypeSingleSelect
	case "multi_select", "multiselect":
		return common.ValueTypeMultiSelect
	case "checkbox":
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}

func (d getResponseCustomField) selectValues() []common.FieldValue {
	if len(d.Values) == 0 {
		return nil
	}

	out := make([]common.FieldValue, 0, len(d.Values))
	for _, v := range d.Values {
		out = append(out, common.FieldValue{Value: v, DisplayValue: v})
	}

	return out
}

// fetchCustomFieldDefinitions lists all custom field definitions (paginated).
func (c *Connector) fetchCustomFieldDefinitions(ctx context.Context) ([]getResponseCustomField, error) {
	var out []getResponseCustomField

	page := 1

	for {
		items, err := c.fetchCustomFieldsPageNodes(ctx, page)
		if err != nil {
			return nil, err
		}

		if len(items) == 0 {
			break
		}

		parsed, err := parseCustomFieldDefinitionNodes(items)
		if err != nil {
			return nil, err
		}

		out = append(out, parsed...)

		if len(items) < maxPageSizeInt {
			break
		}

		page++
	}

	return out, nil
}

func (c *Connector) fetchCustomFieldsPageNodes(ctx context.Context, page int) ([]*ajson.Node, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "/custom-fields")
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(pageKey, strconv.Itoa(page))
	url.WithQueryParam(pageSizeKey, pageSize)

	res, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	body, ok := res.Body()
	if !ok {
		return nil, fmt.Errorf("%w: empty custom-fields body", common.ErrEmptyJSONHTTPResponse)
	}

	return jsonquery.New(body).ArrayOptional("")
}

func parseCustomFieldDefinitionNodes(items []*ajson.Node) ([]getResponseCustomField, error) {
	out := make([]getResponseCustomField, 0, len(items))

	for _, node := range items {
		if node == nil {
			continue
		}

		parsed, err := jsonquery.ParseNode[getResponseCustomField](node)
		if err != nil {
			return nil, err
		}

		if parsed.CustomFieldId != "" {
			out = append(out, *parsed)
		}
	}

	return out, nil
}

// contactReadRecordCustomFieldsTransformer flattens customFieldValues onto the record map
// (keys cf_<id>) so they participate in field selection like other connectors.
//
// FlattenNestedFields is not used here: GetResponse returns an array of {customFieldId, value}
// entries, not a single nested object whose keys should be promoted (see Capsule custom.go).
func contactReadRecordCustomFieldsTransformer(node *ajson.Node) (map[string]any, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	entries, err := jsonquery.New(node).ArrayOptional("customFieldValues")
	if err != nil {
		return nil, err
	}

	for _, entryNode := range entries {
		field, err := jsonquery.ParseNode[contactCustomFieldValue](entryNode)
		if err != nil {
			return nil, err
		}

		if field.CustomFieldId == "" {
			continue
		}

		root[CustomFieldKey(field.CustomFieldId)] = field.normalizedValue()
	}

	return root, nil
}

func flattenContactCustomFieldValues(object map[string]any) {
	entries := parseContactCustomFieldValuesFromMap(object)
	applyFlattenedCustomFieldValues(object, entries)
}

func parseContactCustomFieldValuesFromMap(object map[string]any) []contactCustomFieldValue {
	raw, ok := object["customFieldValues"]
	if !ok {
		return nil
	}

	slice, ok := raw.([]any)
	if !ok || len(slice) == 0 {
		return nil
	}

	entries := make([]contactCustomFieldValue, 0, len(slice))

	for _, item := range slice {
		record, ok := item.(map[string]any)
		if !ok {
			continue
		}

		id, _ := record["customFieldId"].(string)
		if id == "" {
			continue
		}

		entries = append(entries, contactCustomFieldValue{
			CustomFieldId: id,
			Value:         record["value"],
		})
	}

	return entries
}

func applyFlattenedCustomFieldValues(object map[string]any, entries []contactCustomFieldValue) {
	for _, field := range entries {
		if field.CustomFieldId == "" {
			continue
		}

		object[CustomFieldKey(field.CustomFieldId)] = field.normalizedValue()
	}
}

// normalizeCustomFieldAPIValue unwraps GetResponse custom-field "value" shapes so flattened
// cf_<id> fields are easy to read and compare.
//
// The API often wraps scalars in a one-element array (e.g. single_select). We collapse that
// to a scalar; multi-value fields stay as slices; empty arrays become nil.
//
// Sample customFieldValues entry from GET /v3/contacts:
//
//	{"customFieldId": "abc", "value": ["gold"]}           // single_select → cf_abc: "gold"
//	{"customFieldId": "def", "value": ["a", "b"]}         // multi_select  → cf_def: ["a", "b"]
//	{"customFieldId": "ghi", "value": "plain"}            // text          → cf_ghi: "plain"
//	{"customFieldId": "jkl", "value": []}                  // unset         → cf_jkl: nil
func normalizeCustomFieldAPIValue(v any) any {
	arr, ok := v.([]any)
	if !ok {
		return v
	}

	if len(arr) == 0 {
		return nil
	}

	if len(arr) == 1 {
		return arr[0]
	}

	return arr
}

// contactReadFieldsQueryForAPI builds the `fields` query parameter for GET contacts.
// GetResponse only returns customFieldValues when that property is requested; callers
// asking for cf_<customFieldId> need it present or flattening has no source data.
func contactReadFieldsQueryForAPI(fieldNames []string) []string {
	if len(fieldNames) == 0 {
		return fieldNames
	}

	needsCustomFieldValues := false

	for _, fieldName := range fieldNames {
		if strings.EqualFold(fieldName, "customFieldValues") {
			return fieldNames
		}

		if strings.HasPrefix(fieldName, customFieldKeyPrefix) {
			needsCustomFieldValues = true
		}
	}

	if !needsCustomFieldValues {
		return fieldNames
	}

	// Copy the original fields so appending customFieldValues does not mutate the caller's slice.
	queryFields := make([]string, len(fieldNames), len(fieldNames)+1)
	copy(queryFields, fieldNames)
	queryFields = append(queryFields, "customFieldValues")

	return queryFields
}

// mergeContactCustomFieldValuesIntoBody moves keys prefixed with cf_ into
// the customFieldValues array expected by the GetResponse API. Keys with an
// empty id after the prefix are left on the record unchanged.
func mergeContactCustomFieldValuesIntoBody(record map[string]any) map[string]any {
	if len(record) == 0 {
		return record
	}

	fromPrefix := collectCustomFieldEntriesFromCfPrefixedKeys(record)
	if len(fromPrefix) == 0 {
		return record
	}

	merged := copyRecordExcludingCfKeysAndCustomFieldValues(record)
	merged["customFieldValues"] = combineCustomFieldValueArrays(record["customFieldValues"], fromPrefix)

	return merged
}

func collectCustomFieldEntriesFromCfPrefixedKeys(record map[string]any) []contactCustomFieldValue {
	fromPrefix := make([]contactCustomFieldValue, 0, len(record))

	for key, value := range record {
		if !strings.HasPrefix(key, customFieldKeyPrefix) {
			continue
		}

		id := strings.TrimPrefix(key, customFieldKeyPrefix)
		if id == "" {
			continue
		}

		fromPrefix = append(fromPrefix, contactCustomFieldValue{
			CustomFieldId: id,
			Value:         toAPIValueArray(value),
		})
	}

	return fromPrefix
}

func copyRecordExcludingCfKeysAndCustomFieldValues(record map[string]any) map[string]any {
	merged := make(map[string]any, len(record))

	for key, value := range record {
		if strings.HasPrefix(key, customFieldKeyPrefix) {
			continue
		}

		if key == "customFieldValues" {
			continue
		}

		merged[key] = value
	}

	return merged
}

func combineCustomFieldValueArrays(existing any, fromPrefix []contactCustomFieldValue) []any {
	existingList, _ := existing.([]any)

	capacity := len(fromPrefix) + len(existingList)
	combined := make([]any, 0, capacity)

	for _, entry := range existingList {
		if m, ok := entry.(map[string]any); ok {
			combined = append(combined, m)
		}
	}

	for i := range fromPrefix {
		combined = append(combined, fromPrefix[i].toAPIEntry())
	}

	return combined
}

func (v contactCustomFieldValue) toAPIEntry() map[string]any {
	return map[string]any{
		"customFieldId": v.CustomFieldId,
		"value":         toAPIValueArray(v.Value),
	}
}

func toAPIValueArray(raw any) []any {
	if arr, ok := raw.([]any); ok {
		return arr
	}

	if s, ok := raw.(string); ok {
		return []any{s}
	}

	if raw == nil {
		return []any{}
	}

	return []any{raw}
}
