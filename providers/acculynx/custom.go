package acculynx

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/spyzhov/ajson"
)

// AccuLynx returns entityType as "Job"/"Contact" (capitalized) but the docs
// show it lowercased — lowercase at the boundary so case drift doesn't break
// the registry lookup.

// AccuLynx custom-field model:
//   - Definitions are global, fetched from /company-settings/custom-fields and
//     bucketed by entityType ("contact" | "job"). The endpoint already filters
//     to active definitions only.
//   - Values are per-record, fetched from /{contacts|jobs}/{id}/custom-fields.
//     Unlike the inline pattern in other providers (Copper, Sellsy, etc.),
//     AccuLynx requires one extra call per parent record on every read.
//
// References:
//   - Definitions: https://apidocs.acculynx.com/reference/getcompanysettingscustomfields
//   - Contact values: https://apidocs.acculynx.com/reference/getcontactcustomfields
//   - Job values: https://apidocs.acculynx.com/reference/getjobcustomfields

const (
	customFieldDefinitionsPath = "company-settings/custom-fields"

	entityTypeContact = "contact"
	entityTypeJob     = "job"
)

//nolint:gochecknoglobals
var customFieldEntityByObject = map[string]string{
	objectContacts: entityTypeContact,
	objectJobs:     entityTypeJob,
}

// usesCustomFields reports whether AccuLynx exposes custom fields directly on
// the records of the given object. Only top-level contacts and jobs do — the
// "contacts/custom-fields" / "jobs/custom-fields" nested objects are handled
// via the separate nested-object pipeline.
func usesCustomFields(objectName string) bool {
	_, ok := customFieldEntityByObject[objectName]

	return ok
}

// customFieldDefinition mirrors the customFieldDefinitionFull schema.
type customFieldDefinition struct {
	ID         string                 `json:"id"`
	Label      string                 `json:"label"`
	EntityType string                 `json:"entityType"`
	FieldType  string                 `json:"fieldType"`
	Options    []customFieldOptionDef `json:"options"`
}

type customFieldOptionDef struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// fieldName derives the slug used to expose the custom field at the top level
// of a record. Lower-cased label with spaces replaced by underscores; no
// prefix. Built-in AccuLynx fields are camelCase, so collisions are avoided
// by case alone. (Same shape as Copper's slug, sans the "custom_field_" prefix.)
func (d customFieldDefinition) fieldName() string {
	return slugFromLabel(d.Label)
}

// valueType maps AccuLynx field types to the connector value-type taxonomy.
// AccuLynx's customFieldType enum is Text|Number|Date|Boolean. Definitions
// with a non-empty options list are treated as singleSelect — the framework
// caller can read possible values from FieldMetadata.Values.
//
// Note: AccuLynx exposes no API signal to distinguish single-select from
// multi-select definitions (unlike Copper's "MultiSelect" type or Sellsy's
// "checkbox"). All options-bearing fields surface as singleSelect; downstream
// code that cares about cardinality should look at the returned values array
// at read time.
func (d customFieldDefinition) valueType() common.ValueType {
	if len(d.Options) > 0 {
		return common.ValueTypeSingleSelect
	}

	switch d.FieldType {
	case "Text":
		return common.ValueTypeString
	case "Number":
		return common.ValueTypeFloat
	case "Date":
		return common.ValueTypeDate
	case "Boolean":
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}

func (d customFieldDefinition) getValues() []common.FieldValue {
	if len(d.Options) == 0 {
		return nil
	}

	values := make([]common.FieldValue, 0, len(d.Options))
	for _, opt := range d.Options {
		values = append(values, common.FieldValue{
			Value:        opt.ID,
			DisplayValue: opt.Value,
		})
	}

	return values
}

func slugFromLabel(label string) string {
	slug := strings.ToLower(strings.TrimSpace(label))

	return strings.ReplaceAll(slug, " ", "_")
}

// customFieldValue mirrors the per-record customField schema. The label
// travels with the value so the read transformer can build slugs without a
// definitions cross-lookup. AccuLynx returns "values" as a plain string array
// (e.g. ["42"] or ["Phone","Email"]), accompanied by "formattedValues" with
// the same shape — we ignore the latter.
type customFieldValue struct {
	ID        string   `json:"id"`
	Label     string   `json:"label"`
	FieldType string   `json:"fieldType"`
	Values    []string `json:"values"`
}

// customFieldDefinitionsResponse mirrors the customFieldDefinitionsCollection
// schema (an `items` array on top of baseCollection).
type customFieldDefinitionsResponse struct {
	Items []customFieldDefinition `json:"items"`
}

// customFieldsResponse mirrors the customFieldsCollection schema.
type customFieldsResponse struct {
	Items []customFieldValue `json:"items"`
}

// fetchCustomFieldDefinitions retrieves all active custom-field definitions
// across both contact and job entity types in one paginated sweep, and
// returns them bucketed by entityType — matching the salesloft pattern.
func (c *Connector) fetchCustomFieldDefinitions(
	ctx context.Context,
) (datautils.NamedLists[customFieldDefinition], error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, c.modulePath(), customFieldDefinitionsPath)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	url.WithQueryParam(pageSizeParam, defaultPageSize)
	url.WithQueryParam(recordStartParam, "0")

	registry := make(datautils.NamedLists[customFieldDefinition])

	for {
		resp, err := c.JSONHTTPClient().Get(ctx, url.String())
		if err != nil {
			return nil, errors.Join(common.ErrResolvingCustomFields, err)
		}

		page, err := common.UnmarshalJSON[customFieldDefinitionsResponse](resp)
		if err != nil {
			return nil, errors.Join(common.ErrResolvingCustomFields, err)
		}

		if page == nil {
			break
		}

		for _, def := range page.Items {
			registry.Add(strings.ToLower(def.EntityType), def)
		}

		if len(page.Items) < maxPageSize {
			break
		}

		current, _ := url.GetFirstQueryParam(recordStartParam)
		currentInt, _ := strconv.Atoi(current)
		url.WithQueryParam(recordStartParam, strconv.Itoa(currentInt+len(page.Items)))
	}

	return registry, nil
}

// fetchCustomFieldValuesForRecords fans out one GET per parent ID concurrently,
// honouring the per-key rate limit via maxConcurrentChildFetch. Returns a
// map keyed by parent ID with the slice of values that record carries.
func (c *Connector) fetchCustomFieldValuesForRecords(
	ctx context.Context,
	parentObject string,
	parentIDs []string,
) (map[string][]customFieldValue, error) {
	if len(parentIDs) == 0 {
		return map[string][]customFieldValue{}, nil
	}

	results := make([][]customFieldValue, len(parentIDs))
	jobs := make([]simultaneously.Job, len(parentIDs))

	for i, parentID := range parentIDs {
		idx, id := i, parentID

		jobs[idx] = func(ctx context.Context) error {
			values, err := c.fetchCustomFieldValuesForRecord(ctx, parentObject, id)
			if err != nil {
				return fmt.Errorf("fetch custom field values for %s %s: %w", parentObject, id, err)
			}

			results[idx] = values

			return nil
		}
	}

	if err := simultaneously.DoCtx(ctx, maxConcurrentChildFetch, jobs...); err != nil {
		return nil, err
	}

	out := make(map[string][]customFieldValue, len(parentIDs))
	for i, id := range parentIDs {
		out[id] = results[i]
	}

	return out, nil
}

func (c *Connector) fetchCustomFieldValuesForRecord(
	ctx context.Context,
	parentObject string,
	parentID string,
) ([]customFieldValue, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, c.modulePath(), parentObject, parentID, "custom-fields")
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(pageSizeParam, defaultPageSize)
	url.WithQueryParam(recordStartParam, "0")

	var collected []customFieldValue

	for {
		resp, err := c.JSONHTTPClient().Get(ctx, url.String())
		if err != nil {
			return nil, err
		}

		page, err := common.UnmarshalJSON[customFieldsResponse](resp)
		if err != nil {
			return nil, err
		}

		if page == nil {
			break
		}

		collected = append(collected, page.Items...)

		if len(page.Items) < maxPageSize {
			break
		}

		current, _ := url.GetFirstQueryParam(recordStartParam)
		currentInt, _ := strconv.Atoi(current)
		url.WithQueryParam(recordStartParam, strconv.Itoa(currentInt+len(page.Items)))
	}

	return collected, nil
}

// attachReadCustomFields returns a RecordTransformer that flattens the
// pre-fetched custom-field values into the record map. The slug is derived
// from the value's own label, which the AccuLynx response carries on each
// record so the transformer doesn't need a definitions cross-lookup. Single-
// element value arrays are unwrapped to scalars; multi-element arrays
// preserved. Raw is never modified — this only mutates the marshalled Fields
// map.
func attachReadCustomFields(valuesByParentID map[string][]customFieldValue) common.RecordTransformer {
	return func(node *ajson.Node) (map[string]any, error) {
		object, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, err
		}

		if len(valuesByParentID) == 0 {
			return object, nil
		}

		parentID, err := jsonquery.New(node).StrWithDefault("id", "")
		if err != nil {
			return nil, err
		}

		for _, value := range valuesByParentID[parentID] {
			object[slugFromLabel(value.Label)] = unwrapValueItems(value.Values)
		}

		return object, nil
	}
}

// unwrapValueItems converts the AccuLynx values array into the most natural
// shape: a scalar string for single-element arrays, a slice of strings for
// multi-element arrays, and nil for empty arrays.
func unwrapValueItems(items []string) any {
	switch len(items) {
	case 0:
		return nil
	case 1:
		return items[0]
	default:
		return items
	}
}

// extractParentIDsFromBody pulls the records array out of the response and
// returns just the "id" field of each record. Used to plan the value fan-out
// before invoking the framework's record-iteration pipeline.
func (c *Connector) extractParentIDsFromBody(objectName string, body *ajson.Node) ([]string, error) {
	records, err := c.recordsFunc(objectName)(body)
	if err != nil {
		return nil, err
	}

	return extractIDs(records), nil
}
