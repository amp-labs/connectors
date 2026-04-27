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
	vt := d.mapValueType()
	isCustom := true

	return common.FieldMetadata{
		DisplayName:  d.Name,
		ValueType:    vt,
		ProviderType: d.ValueType,
		Values:       d.selectValues(),
		IsCustom:     &isCustom,
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

		items, err := jsonquery.New(body).ArrayOptional("")
		if err != nil {
			return nil, err
		}

		if len(items) == 0 {
			break
		}

		for _, n := range items {
			if n == nil {
				continue
			}

			p, err := jsonquery.ParseNode[getResponseCustomField](n)
			if err != nil {
				return nil, err
			}

			if p.CustomFieldId != "" {
				out = append(out, *p)
			}
		}

		if len(items) < maxPageSizeInt {
			break
		}

		page++
	}

	return out, nil
}

// contactReadRecordTransformer flattens customFieldValues onto the record map
// (keys cf_<id>) so they participate in field selection like other connectors.
func contactReadRecordTransformer(node *ajson.Node) (map[string]any, error) {
	obj, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	flattenContactCustomFieldValues(obj)

	return obj, nil
}

func flattenContactCustomFieldValues(object map[string]any) {
	raw, ok := object["customFieldValues"]
	if !ok {
		return
	}

	entries, ok := raw.([]any)
	if !ok || len(entries) == 0 {
		return
	}

	for _, e := range entries {
		m, ok := e.(map[string]any)
		if !ok {
			continue
		}

		id, _ := m["customFieldId"].(string)
		if id == "" {
			continue
		}

		normalized := normalizeCustomFieldAPIValue(m["value"])
		object[CustomFieldKey(id)] = normalized
	}
}

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

	for _, f := range fieldNames {
		if strings.EqualFold(f, "customFieldValues") {
			return fieldNames
		}
	}

	for _, f := range fieldNames {
		if strings.HasPrefix(f, customFieldKeyPrefix) {
			out := make([]string, 0, len(fieldNames)+1)
			out = append(out, fieldNames...)
			out = append(out, "customFieldValues")

			return out
		}
	}

	return fieldNames
}

// mergeContactCustomFieldValuesIntoBody moves keys prefixed with cf_ into
// the customFieldValues array expected by the GetResponse API. Keys with an
// empty id after the prefix are left on the record unchanged.
func mergeContactCustomFieldValuesIntoBody(record map[string]any) map[string]any {
	if len(record) == 0 {
		return record
	}

	var fromPrefix []map[string]any

	for k, v := range record {
		if !strings.HasPrefix(k, customFieldKeyPrefix) {
			continue
		}

		id := strings.TrimPrefix(k, customFieldKeyPrefix)
		if id == "" {
			continue
		}

		fromPrefix = append(fromPrefix, map[string]any{
			"customFieldId": id,
			"value":         toAPIValueArray(v),
		})
	}

	if len(fromPrefix) == 0 {
		return record
	}

	merged := make(map[string]any, len(record))

	for k, v := range record {
		if strings.HasPrefix(k, customFieldKeyPrefix) {
			continue
		}

		if k == "customFieldValues" {
			continue
		}

		merged[k] = v
	}

	existing, hadExisting := record["customFieldValues"]
	combined := make([]any, 0)

	if hadExisting {
		if list, ok := existing.([]any); ok {
			for _, e := range list {
				if m, ok2 := e.(map[string]any); ok2 {
					combined = append(combined, m)
				}
			}
		}
	}

	for _, e := range fromPrefix {
		combined = append(combined, e)
	}

	merged["customFieldValues"] = combined

	return merged
}

func toAPIValueArray(v any) []any {
	if arr, ok := v.([]any); ok {
		return arr
	}

	if s, ok := v.(string); ok {
		return []any{s}
	}

	if v == nil {
		return []any{}
	}

	return []any{v}
}
