package clarify

import (
	"context"
	"encoding/json"
	"net/url"
	"path"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Schema resources are identified by a URL whose last segments name the object:
// https://getclarify.ai/schemas/entities/person, .../entities/c_sales_order.
// `.../core/*` entries are shared base definitions, not objects.
// https://developer.clarify.ai/docs/guides/schemas-and-relationships#schema-fields
const entitySchemaMarker = "/entities/"

// listObjectMetadata describes objects from GET /v1/workspaces/{slug}/schemas, which returns every
// object's JSON Schema (fields, types, enums, relationships), built-in and custom alike, so custom
// fields are described with the workspace's own names. The list is fetched once per call.
// https://developer.clarify.ai/docs/getting-started/key-concepts#schemas-and-fields
func (c *Connector) listObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	schemas, err := c.fetchSchemas(ctx)
	if err != nil {
		return nil, err
	}

	result := common.NewListObjectMetadataResult()

	for _, objectName := range datautils.NewSetFromList(objectNames).List() {
		entity, ok := schemas[entityForObject(objectName)]
		if !ok || !isReadableObject(objectName) {
			result.AppendError(objectName, common.ErrObjectNotSupported)

			continue
		}

		metadata := entity.objectMetadata(objectName)
		if objectName == objectNameUsers {
			restrictToUsersEndpoint(metadata)
		}

		result.Result[objectName] = *metadata
	}

	return result, nil
}

// The `user` entity schema also describes stored columns that GET /users does not return (auth0_id,
// workos_id, _deleted_at, _created_by, ...), and lacks the derived is_service_account it does return.
// These are the attributes of ListOfUsersResponse in the OpenAPI spec, which the live response matches.
// https://developer.clarify.ai/docs/api-reference/users/getUsers
var usersEndpointFields = datautils.NewSet( // nolint:gochecknoglobals
	"id", "_id", "name", "email", "roles", "_created_at", "_updated_at", "last_active_at", "profile_picture",
)

func restrictToUsersEndpoint(metadata *common.ObjectMetadata) {
	for name := range metadata.Fields {
		if !usersEndpointFields.Has(name) {
			metadata.RemoveFieldMetadata(name)
		}
	}

	metadata.AddFieldMetadata("is_service_account", common.FieldMetadata{
		DisplayName:  "Is service account",
		ValueType:    common.ValueTypeBoolean,
		ProviderType: "boolean",
		ReadOnly:     new(true),
	})
}

// fetchSchemas reads every page of the schemas list, following links.next, keyed by object name.
func (c *Connector) fetchSchemas(ctx context.Context) (map[string]entitySchema, error) {
	schemasURL, err := c.workspaceURL("schemas")
	if err != nil {
		return nil, err
	}

	schemas := make(map[string]entitySchema)
	next := schemasURL.String()

	for next != "" {
		page, err := c.fetchSchemasPage(ctx, next)
		if err != nil {
			return nil, err
		}

		for _, resource := range page.Data {
			_, objectName, found := strings.Cut(resource.ID, entitySchemaMarker)
			if found && objectName != "" {
				schemas[objectName] = resource.Attributes
			}
		}

		next = ""
		if page.Links.Next != nil {
			next, err = c.onBaseURL(*page.Links.Next)
			if err != nil {
				return nil, err
			}
		}
	}

	return schemas, nil
}

func (c *Connector) fetchSchemasPage(ctx context.Context, pageURL string) (*schemasResponse, error) {
	response, err := c.JSONHTTPClient().Get(ctx, pageURL)
	if err != nil {
		return nil, err
	}

	page, err := common.UnmarshalJSON[schemasResponse](response)
	if err != nil {
		return nil, err
	}

	if page == nil {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	return page, nil
}

// onBaseURL keeps a links.next URL's path and cursor but sends it to the connector's base URL: list
// endpoints return links.next with an http:// scheme that does not answer (see makeNextPageFunc).
func (c *Connector) onBaseURL(link string) (string, error) {
	next, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	base, err := url.Parse(c.ProviderInfo().BaseURL)
	if err != nil {
		return "", err
	}

	next.Scheme = base.Scheme
	next.Host = base.Host

	return next.String(), nil
}

// https://developer.clarify.ai/docs/api-reference/schemas/getSchemas
type schemasResponse struct {
	Data []struct {
		ID         string       `json:"id"`
		Attributes entitySchema `json:"attributes"`
	} `json:"data"`
	Links struct {
		Next *string `json:"next"`
	} `json:"links"`
}

type entitySchema struct {
	Title      string                    `json:"title"`
	Label      *entityLabel              `json:"xClarifyLabel"`
	Properties map[string]schemaProperty `json:"properties"`
}

type entityLabel struct {
	Plural   string `json:"plural"`
	Singular string `json:"singular"`
}

type schemaProperty struct {
	Type         schemaType            `json:"type"`
	Format       string                `json:"format"`
	Title        string                `json:"title"`
	Enum         []any                 `json:"enum"`
	ReadOnly     *bool                 `json:"readOnly"`
	Items        *schemaProperty       `json:"items"`
	OneOf        []schemaVariant       `json:"oneOf"`
	Relationship *relationshipProperty `json:"xClarifyRelationship"`
}

// schemaVariant is one branch of a nullable shared type: {"$ref": ".../core/collectionOfEmails"}.
type schemaVariant struct {
	Ref string `json:"$ref"`
}

type relationshipProperty struct {
	Kind   string `json:"kind"`
	Field  string `json:"field"`
	Entity string `json:"entity"`
}

// schemaType is a JSON Schema `type`: one name, or a list of names such as ["string", "null"].
type schemaType []string

func (t *schemaType) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*t = schemaType{single}

		return nil
	}

	var list []string
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	*t = list

	return nil
}

// name is the type without "null", which only marks the field as nullable.
func (t schemaType) name() string {
	for _, name := range t {
		if name != "null" {
			return name
		}
	}

	return ""
}

func (e entitySchema) objectMetadata(objectName string) *common.ObjectMetadata {
	fields := make(common.FieldsMetadata, len(e.Properties)+1)

	for name, property := range e.Properties {
		if !property.isReadable() {
			continue
		}

		fields[name] = property.fieldMetadata(name)
	}

	// Read returns the resource id as `id`, beside the `_id` attribute.
	if _, ok := fields["id"]; !ok {
		fields["id"] = common.FieldMetadata{
			DisplayName:  "ID",
			ValueType:    common.ValueTypeString,
			ProviderType: "string",
			ReadOnly:     new(true),
		}
	}

	return common.NewObjectMetadata(e.displayName(objectName), fields)
}

func (e entitySchema) displayName(objectName string) string {
	if e.Label != nil && e.Label.Plural != "" {
		return e.Label.Plural
	}

	if e.Title != "" {
		return e.Title
	}

	return naming.CapitalizeFirstLetterEveryWord(objectName)
}

// isReadable reports whether Read returns the property among a record's attributes. To-one
// relationships are stored as a foreign-key attribute (company_id); to-many relationships are kept
// between records and are only returned through `include`, which this connector does not request.
// https://developer.clarify.ai/docs/guides/import-crm-data#4-link-people-to-deals
func (p schemaProperty) isReadable() bool {
	if p.Relationship == nil {
		return true
	}

	return p.Relationship.Kind == "many-to-one" || p.Relationship.Kind == "one-to-one"
}

func (p schemaProperty) fieldMetadata(name string) common.FieldMetadata {
	valueType, values := p.valueType()

	displayName := p.Title
	if displayName == "" {
		displayName = name
	}

	field := common.FieldMetadata{
		DisplayName:  displayName,
		ValueType:    valueType,
		ProviderType: p.providerType(),
		ReadOnly:     p.ReadOnly,
		Values:       values,
	}

	// A link is a reference only when its target is an object this connector reads. Links to entities it
	// doesn't read (meeting.recurrence_rule_id, meeting.public_link_id) keep the type of the id they hold.
	if p.Relationship != nil {
		if target := objectNameForEntity(p.Relationship.Entity); isReadableObject(target) {
			field.ValueType = common.ValueTypeReference
			field.ReferenceTo = []string{target}
		}
	}

	return field
}

// providerType is the JSON Schema type with its format, or the shared type a oneOf points to
// (collectionOfEmails, personName, summary).
func (p schemaProperty) providerType() string {
	for _, variant := range p.OneOf {
		if variant.Ref != "" {
			return path.Base(variant.Ref)
		}
	}

	if p.Format != "" {
		return p.Type.name() + ":" + p.Format
	}

	return p.Type.name()
}

// valueType maps a JSON Schema type and format to an Ampersand value type.
// Collection fields (email_addresses, domains) are objects holding an `items` array, so they are other.
// https://developer.clarify.ai/docs/getting-started/key-concepts#collection-fields
func (p schemaProperty) valueType() (common.ValueType, []common.FieldValue) {
	// Enums list null among their values; an enum of only null offers no options.
	if values := enumValues(p.Enum); len(values) != 0 {
		return common.ValueTypeSingleSelect, values
	}

	// Timestamps are typed "object" or ["object", "string"] with format date-time (_created_at,
	// _updated_at, start), yet records carry them as ISO strings, so the format decides.
	switch p.Format {
	case "date-time":
		return common.ValueTypeDateTime, nil
	case "date":
		return common.ValueTypeDate, nil
	}

	if p.Type.name() == "array" && p.Items != nil {
		if values := enumValues(p.Items.Enum); len(values) != 0 {
			return common.ValueTypeMultiSelect, values
		}
	}

	if valueType, ok := valueTypeByJSONType[p.Type.name()]; ok {
		return valueType, nil
	}

	return common.ValueTypeOther, nil
}

// JSON Schema types with a direct Ampersand value type; objects, arrays and shared types are other.
var valueTypeByJSONType = map[string]common.ValueType{ // nolint:gochecknoglobals
	"string":  common.ValueTypeString,
	"number":  common.ValueTypeFloat,
	"integer": common.ValueTypeInt,
	"boolean": common.ValueTypeBoolean,
}

func enumValues(enum []any) []common.FieldValue {
	values := make([]common.FieldValue, 0, len(enum))

	for _, value := range enum {
		text, ok := value.(string)
		if !ok {
			continue
		}

		values = append(values, common.FieldValue{Value: text, DisplayValue: text})
	}

	return values
}

// entityForObject names the schema that describes an object: users are the `user` entity.
func entityForObject(objectName string) string {
	if objectName == objectNameUsers {
		return "user"
	}

	return objectName
}

// objectNameForEntity names a relationship target by this connector's object name.
func objectNameForEntity(entity string) string {
	if entity == "user" {
		return objectNameUsers
	}

	return entity
}
