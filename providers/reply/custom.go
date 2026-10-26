package reply

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/reply/metadata"
	"github.com/spyzhov/ajson"
)

// Reply custom field model:
//   - Definitions are fetched from GET /v3/custom-fields, one entity type per
//     call: contact (the default) or account. The endpoint is not paginated.
//   - Values travel inline on each record as customFields: [{key, value}], where
//     key is the numeric definition id as a string. Only fields holding a value
//     are listed.
//   - Writes name a field differently per operation (see referencesByKey), and
//     every value is a string, even for number fields.
//
// Custom fields are exposed natively by their title. Reply refuses titles that
// match its standard fields (Industry, Company, ...) but not those that match an
// API field (notes, name), so a title equal to a built-in field is skipped.
//
// References:
//   - Definitions: https://docs.reply.io/api-reference/custom-fields/list-all-custom-fields
//   - Contact values: https://docs.reply.io/api-reference/contacts/update-a-contact
//   - Account values: https://docs.reply.io/api-reference/contact-accounts/update-an-account
const (
	entityTypeContact = "contact"
	entityTypeAccount = "account"

	customFieldsKey = "customFields"
)

// customFieldEntityByObject lists the objects whose records carry custom fields
// and the Reply entity type their definitions are filed under.
//
//nolint:gochecknoglobals
var customFieldEntityByObject = map[string]string{
	"contacts":         entityTypeContact,
	"contact-accounts": entityTypeAccount,
}

// usesCustomFields reports whether the records of the object carry custom fields.
func usesCustomFields(objectName string) bool {
	_, ok := customFieldEntityByObject[objectName]

	return ok
}

// writeOnlyFields are properties accepted by the write endpoints of an object
// that its list response never returns, so the read schema doesn't know them.
// They must not be mistaken for custom field names on write.
//
//nolint:gochecknoglobals
var writeOnlyFields = map[string]map[string]bool{
	"contacts": {"phone2": true},
}

// customFieldDefinition is one entry of the GET /v3/custom-fields response.
// https://docs.reply.io/api-reference/custom-fields/list-all-custom-fields
type customFieldDefinition struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	FieldType string `json:"fieldType"`
	Metadata  string `json:"metadata"`
	OrgWide   bool   `json:"orgWide"`
}

// valueType maps Reply field types to the connector value-type taxonomy.
// Reply has only text and number custom fields, no enum or select types. The
// definition's "metadata" is free-form text Reply doesn't validate, so no
// Values are offered.
func (d customFieldDefinition) valueType() common.ValueType {
	switch d.FieldType {
	case "number":
		return common.ValueTypeFloat
	case "text":
		return common.ValueTypeString
	default:
		return common.ValueTypeOther
	}
}

// fetchCustomFieldDefinitions returns every custom field definition of the
// entity type. The endpoint responds with a root-level array and is not paginated.
//
// Account custom fields are a feature Reply gates per team, so a 403 for the
// account entity means the team has none rather than a failure. Every other
// failure aborts the call, so callers don't silently miss custom fields.
// https://docs.reply.io/api-reference/contact-accounts/create-an-account
func (c *Connector) fetchCustomFieldDefinitions(
	ctx context.Context, entityType string,
) ([]customFieldDefinition, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "custom-fields")
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	url.WithQueryParam("entityType", entityType)

	response, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		if entityType == entityTypeAccount && errors.Is(err, common.ErrForbidden) {
			return []customFieldDefinition{}, nil
		}

		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	definitions, err := common.UnmarshalJSON[[]customFieldDefinition](response)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	if definitions == nil {
		return []customFieldDefinition{}, nil
	}

	return *definitions, nil
}

// builtInFieldNames returns the lowercased names of the fields the object
// already has, as ReadResult.Fields keys are lowercased.
func builtInFieldNames(objectName string) (map[string]bool, error) {
	selected, err := metadata.Schemas.Select(common.ModuleRoot, []string{objectName})
	if err != nil {
		return nil, err
	}

	names := make(map[string]bool)

	// GetObjectMetadata guards against an object whose maps are nil.
	objectMetadata := selected.GetObjectMetadata(objectName)
	if objectMetadata == nil {
		return names, nil
	}

	for name := range objectMetadata.Fields {
		names[strings.ToLower(name)] = true
	}

	return names, nil
}

// exposedCustomFields returns the definitions that are surfaced as top-level
// fields of the object, named by their title. A custom field whose title
// matches a built-in field of the object (ignoring case) is left out: the
// built-in field wins, and the custom value stays reachable through Raw.
func (c *Connector) exposedCustomFields(
	ctx context.Context, objectName string,
) ([]customFieldDefinition, error) {
	builtIn, err := builtInFieldNames(objectName)
	if err != nil {
		return nil, err
	}

	definitions, err := c.fetchCustomFieldDefinitions(ctx, customFieldEntityByObject[objectName])
	if err != nil {
		return nil, err
	}

	exposed := make([]customFieldDefinition, 0, len(definitions))

	for _, definition := range definitions {
		if !builtIn[strings.ToLower(definition.Title)] {
			exposed = append(exposed, definition)
		}
	}

	return exposed, nil
}

// attachReadCustomFields returns a RecordTransformer that resolves the record's
// customFields entries to human-readable names: each entry's key (the numeric
// definition id) is matched against the definition titles fetched once per
// page, and the value is placed at the top level of the record under the title.
// Raw is never modified, this only mutates the marshalled Fields map.
func attachReadCustomFields(definitions []customFieldDefinition) common.RecordTransformer {
	titleByID := make(map[string]string, len(definitions))
	for _, definition := range definitions {
		titleByID[strconv.Itoa(definition.ID)] = definition.Title
	}

	return func(node *ajson.Node) (map[string]any, error) {
		record, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, err
		}

		fields, ok := record[customFieldsKey].([]any)
		if !ok {
			return record, nil
		}

		for _, field := range fields {
			entry, ok := field.(map[string]any)
			if !ok {
				continue
			}

			key, _ := entry["key"].(string)

			if title, found := titleByID[key]; found {
				record[title] = entry["value"]
			}
		}

		return record, nil
	}
}

// fieldReference points at a custom field definition by id or by title.
type fieldReference struct {
	id   int
	name string
}

func (r fieldReference) byID() bool {
	return r.id > 0
}

// key is the value of the "key" property: the id serialized as a string, or the title.
func (r fieldReference) key() string {
	if r.byID() {
		return strconv.Itoa(r.id)
	}

	return r.name
}

// normalizedEntries rewrites each entry to the reference form the endpoint
// takes. A reference by title is kept as such: Reply resolves "key" and "name"
// alike, so only the property name changes.
func normalizedEntries(entries []any, byKey bool) []any {
	normalized := make([]any, len(entries))

	for index, field := range entries {
		normalized[index] = field

		entry, ok := field.(map[string]any)
		if !ok {
			continue
		}

		reference, found := entryReference(entry)
		if !found {
			continue
		}

		normalized[index] = entryWithReference(entry, reference, byKey)
	}

	return normalized
}

// entryWithReference rebuilds the entry as {<reference property>, "value"},
// dropping what a write doesn't take, like the enrichmentStatus returned on read.
// A missing value stays missing instead of becoming null, which would clear the field.
func entryWithReference(entry map[string]any, reference fieldReference, byKey bool) map[string]any {
	var rebuilt map[string]any

	switch {
	case byKey:
		rebuilt = map[string]any{"key": reference.key()}
	case reference.byID():
		rebuilt = map[string]any{"id": reference.id}
	default:
		rebuilt = map[string]any{"name": reference.name}
	}

	if value, hasValue := entry["value"]; hasValue {
		rebuilt["value"] = stringifyValue(value)
	}

	return rebuilt
}

// entryReference extracts what an entry points at, whichever of "id", "key" or
// "name" it uses. It reports false for an entry with none of them.
func entryReference(entry map[string]any) (fieldReference, bool) {
	for _, property := range []string{"id", "key"} {
		if reference, ok := referenceFrom(entry[property]); ok {
			return reference, true
		}
	}

	if name, ok := entry["name"].(string); ok && name != "" {
		return fieldReference{name: name}, true
	}

	return fieldReference{}, false
}

// referenceFrom reads an id-like value: a number, or text that is either a
// numeric id or a title.
func referenceFrom(value any) (fieldReference, bool) {
	switch typed := value.(type) {
	case int:
		return fieldReference{id: typed}, typed > 0
	case float64:
		return fieldReference{id: int(typed)}, typed > 0
	case string:
		if id, err := strconv.Atoi(typed); err == nil {
			return fieldReference{id: id}, id > 0
		}

		return fieldReference{name: typed}, typed != ""
	default:
		return fieldReference{}, false
	}
}

// withoutOverridden drops the entries pointing at a field id that is also set
// by a flattened title-named property. A ReadResult can carry both forms of
// the same value; the title-named one is what a caller edits, so it wins.
func withoutOverridden(entries []any, overridden map[int]bool) []any {
	kept := make([]any, 0, len(entries))

	for _, field := range entries {
		if entry, isMap := field.(map[string]any); isMap {
			if reference, found := entryReference(entry); found && overridden[reference.id] {
				continue
			}
		}

		kept = append(kept, field)
	}

	return kept
}

// flattenedField is a top-level record property that names a custom field.
type flattenedField struct {
	key   string // the property as the caller wrote it
	id    int    // the definition it resolves to
	value any
}

// flattenedCustomFields picks the top-level properties of the record that are
// not built-in fields and name a custom field by its title (ignoring case).
// The definitions are only fetched when the record has a property that could be one.
func (c *Connector) flattenedCustomFields(
	ctx context.Context, objectName string, record map[string]any,
) ([]flattenedField, error) {
	candidates, err := customFieldCandidates(objectName, record)
	if err != nil || len(candidates) == 0 {
		return nil, err
	}

	definitions, err := c.exposedCustomFields(ctx, objectName)
	if err != nil {
		return nil, err
	}

	idByTitle := make(map[string]int, len(definitions))
	for _, definition := range definitions {
		idByTitle[strings.ToLower(definition.Title)] = definition.ID
	}

	fields := make([]flattenedField, 0, len(candidates))

	for _, key := range candidates {
		if id, found := idByTitle[strings.ToLower(key)]; found {
			fields = append(fields, flattenedField{key: key, id: id, value: record[key]})
		}
	}

	return fields, nil
}

// customFieldCandidates returns the sorted keys of the record that could name a
// custom field: everything that is neither a built-in field, a write-only
// property nor the customFields container itself.
func customFieldCandidates(objectName string, record map[string]any) ([]string, error) {
	builtIn, err := builtInFieldNames(objectName)
	if err != nil {
		return nil, err
	}

	candidates := make([]string, 0)

	for key := range record {
		lowered := strings.ToLower(key)
		if lowered == strings.ToLower(customFieldsKey) || builtIn[lowered] || writeOnlyFields[objectName][lowered] {
			continue
		}

		candidates = append(candidates, key)
	}

	// Map iteration order is random; a stable order keeps the request body deterministic.
	slices.Sort(candidates)

	return candidates, nil
}

// stringifyValue renders a custom field value the way Reply expects it: every
// value is a string, including those of number fields. Null clears the value.
func stringifyValue(value any) any {
	switch typed := value.(type) {
	case nil, string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return fmt.Sprint(typed)
	}
}
