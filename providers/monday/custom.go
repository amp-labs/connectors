package monday

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// Field keys for board column values on items use cf_<columnId>.
// Column ids are board-scoped (e.g. status, date4).
const customFieldKeyPrefix = "cf_"

// objectsWithCustomFields lists objects whose column values are exposed as cf_<columnId>.
//
//nolint:gochecknoglobals
var objectsWithCustomFields = datautils.NewStringSet(mondayObjectItems)

// mondayColumnDefinition is a board column from the Monday.com GraphQL API.
// https://developer.monday.com/api-reference/reference/boards
type mondayColumnDefinition struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	SettingsStr string `json:"settings_str"`
}

// columnValue is one element of item column_values.
// https://developer.monday.com/api-reference/reference/items
type columnValue struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Value any    `json:"value"`
	Type  string `json:"type"`
}

// CustomFieldKey returns the connector field name for a Monday column id.
func CustomFieldKey(columnID string) string {
	return customFieldKeyPrefix + columnID
}

func (d mondayColumnDefinition) fieldMetadata() common.FieldMetadata {
	return common.FieldMetadata{
		DisplayName:  d.Title,
		ValueType:    d.mapValueType(),
		ProviderType: d.Type,
		Values:       d.selectValues(),
		IsCustom:     new(true),
	}
}

func (d mondayColumnDefinition) mapValueType() common.ValueType {
	switch strings.ToLower(d.Type) {
	case "text", "long_text", "email", "phone", "link", "country", "location":
		return common.ValueTypeString
	case "numbers", "rating", "week", "hour":
		return common.ValueTypeFloat
	case "checkbox":
		return common.ValueTypeBoolean
	case "date", "timeline", "time_tracking", "creation_log", "last_updated":
		return common.ValueTypeDate
	case "status", "color", "dropdown", "tags":
		return common.ValueTypeSingleSelect
	case "people", "team", "board_relation", "dependency", "connect_boards", "mirror", "subtasks":
		return common.ValueTypeOther
	default:
		return common.ValueTypeOther
	}
}

func (d mondayColumnDefinition) selectValues() []common.FieldValue {
	if d.SettingsStr == "" {
		return nil
	}

	var settings map[string]any
	if err := json.Unmarshal([]byte(d.SettingsStr), &settings); err != nil {
		return nil
	}

	labels, ok := settings["labels"].(map[string]any)
	if !ok || len(labels) == 0 {
		return nil
	}

	out := make([]common.FieldValue, 0, len(labels))
	labelValues := make([]string, 0, len(labels))
	for _, label := range labels {
		if s, ok := label.(string); ok {
			labelValues = append(labelValues, s)
		}
	}

	sort.Strings(labelValues)

	for _, s := range labelValues {
		out = append(out, common.FieldValue{Value: s, DisplayValue: s})
	}

	return out
}

func (c *Connector) fetchBoardColumnDefinitions(ctx context.Context, boardID string) ([]mondayColumnDefinition, error) {
	query := fmt.Sprintf(`query {
		boards(ids: [%s]) {
			columns {
				id
				title
				type
				settings_str
			}
		}
	}`, boardID)

	res, err := c.postGraphQL(ctx, query)
	if err != nil {
		return nil, err
	}

	body, ok := res.Body()
	if !ok {
		return nil, fmt.Errorf("%w: empty boards columns body", common.ErrEmptyJSONHTTPResponse)
	}

	dataNode, err := body.GetKey("data")
	if err != nil {
		return nil, err
	}

	boards, err := jsonquery.New(dataNode).ArrayOptional("boards")
	if err != nil {
		return nil, err
	}

	if len(boards) == 0 {
		return nil, nil
	}

	columns, err := jsonquery.New(boards[0]).ArrayOptional("columns")
	if err != nil {
		return nil, err
	}

	out := make([]mondayColumnDefinition, 0, len(columns))
	for _, colNode := range columns {
		col, err := jsonquery.ParseNode[mondayColumnDefinition](colNode)
		if err != nil || col.ID == "" {
			continue
		}

		out = append(out, *col)
	}

	return out, nil
}

func columnDefinitionsByID(columns []mondayColumnDefinition) map[string]mondayColumnDefinition {
	result := make(map[string]mondayColumnDefinition, len(columns))
	for _, col := range columns {
		result[col.ID] = col
	}

	return result
}

// normalizedValue picks the connector-facing value for a column_values entry.
//
// Monday often returns human-readable text alongside a JSON value string. Sample entries:
//
//	{"id": "status", "text": "Done", "type": "status", "value": "{\"index\":1,\"label\":\"Done\"}"}
//	→ cf_status: "Done"
//
//	{"id": "numbers", "text": "42", "type": "numbers", "value": "42"}
//	→ cf_numbers: "42"
//
//	{"id": "text", "text": "", "type": "text", "value": null}
//	→ cf_text: nil
func (v columnValue) normalizedValue() any {
	if strings.TrimSpace(v.Text) != "" {
		return v.Text
	}

	if v.Value == nil {
		return nil
	}

	if s, ok := v.Value.(string); ok {
		if strings.TrimSpace(s) == "" {
			return nil
		}

		return s
	}

	return v.Value
}

// itemReadRecordCustomFieldsTransformer flattens column_values onto the record map
// (keys cf_<columnId>) so they participate in field selection like other connectors.
//
// FlattenNestedFields is not used here: Monday returns column_values as an array of
// {id, text, value, type}, not a single nested object whose keys should be promoted.
func itemReadRecordCustomFieldsTransformer(node *ajson.Node) (map[string]any, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	entries, err := jsonquery.New(node).ArrayOptional("column_values")
	if err != nil {
		return nil, err
	}

	for _, entryNode := range entries {
		field, err := jsonquery.ParseNode[columnValue](entryNode)
		if err != nil {
			return nil, err
		}

		if field.ID == "" {
			continue
		}

		root[CustomFieldKey(field.ID)] = field.normalizedValue()
	}

	return root, nil
}

// itemReadCustomFieldsQueryNeedsColumnValues reports whether the items GraphQL query
// must request column_values (only needed when callers ask for cf_* or column_values).
func itemReadCustomFieldsQueryNeedsColumnValues(fieldNames []string) bool {
	for _, fieldName := range fieldNames {
		if strings.EqualFold(fieldName, "column_values") {
			return true
		}

		if strings.HasPrefix(fieldName, customFieldKeyPrefix) {
			return true
		}
	}

	return false
}

// prepareItemWriteCustomFieldsRecordData maps cf_<columnId> keys into column_values JSON for Monday mutations.
func prepareItemWriteCustomFieldsRecordData(
	record map[string]any,
	columns map[string]mondayColumnDefinition,
) (map[string]any, error) {
	if len(record) == 0 {
		return record, nil
	}

	fromCfKeys := collectColumnValuesFromCfKeys(record, columns)
	if len(fromCfKeys) == 0 {
		return record, nil
	}

	payload := copyItemWriteRecordExcludingCustomFields(record)

	existing, err := parseExistingColumnValuesJSON(payload["column_values"])
	if err != nil {
		return nil, err
	}

	for key, value := range fromCfKeys {
		existing[key] = value
	}

	columnValuesJSON, err := json.Marshal(existing)
	if err != nil {
		return nil, err
	}

	payload["column_values"] = string(columnValuesJSON)

	return payload, nil
}

func collectColumnValuesFromCfKeys(
	record map[string]any,
	columns map[string]mondayColumnDefinition,
) map[string]any {
	out := make(map[string]any)

	for key, value := range record {
		if !strings.HasPrefix(key, customFieldKeyPrefix) {
			continue
		}

		columnID := strings.TrimPrefix(key, customFieldKeyPrefix)
		if columnID == "" {
			continue
		}

		colType := ""
		if col, ok := columns[columnID]; ok {
			colType = col.Type
		}

		out[columnID] = formatColumnValueForAPI(colType, value)
	}

	return out
}

func formatColumnValueForAPI(columnType string, value any) any {
	switch strings.ToLower(columnType) {
	case "status", "color":
		if s, ok := value.(string); ok {
			return map[string]any{"label": s}
		}
	case "checkbox":
		if b, ok := value.(bool); ok {
			checked := "false"
			if b {
				checked = "true"
			}

			return map[string]any{"checked": checked}
		}
	}

	return value
}

func copyItemWriteRecordExcludingCustomFields(record map[string]any) map[string]any {
	merged := make(map[string]any, len(record))

	for key, value := range record {
		if strings.HasPrefix(key, customFieldKeyPrefix) {
			continue
		}

		merged[key] = value
	}

	return merged
}

func parseExistingColumnValuesJSON(existing any) (map[string]any, error) {
	if existing == nil {
		return map[string]any{}, nil
	}

	switch existingValue := existing.(type) {
	case string:
		if strings.TrimSpace(existingValue) == "" {
			return map[string]any{}, nil
		}

		var parsed map[string]any
		if err := json.Unmarshal([]byte(existingValue), &parsed); err != nil {
			return nil, err
		}

		return parsed, nil
	case map[string]any:
		return existingValue, nil
	default:
		return map[string]any{}, nil
	}
}
