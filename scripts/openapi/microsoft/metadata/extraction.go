package main

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/xquery"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/scripts/openapi/microsoft/metadata/internal/models"
	"github.com/amp-labs/connectors/scripts/openapi/microsoft/metadata/internal/schemas"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	GraphNamespaceSchemaName = "microsoft.graph" // nolint:gochecknoglobals

	ErrMissingSchema      = fmt.Errorf("missing schema %v in response", GraphNamespaceSchemaName)
	ErrMetadataProcessing = errors.New("metadata couldn't be processed")
)

func extractSchemas() (metadatadef.Schemas[any], error) {
	xml, err := xquery.NewXML(schemas.Data)
	if err != nil {
		return nil, err
	}

	entities, err := extractEntities(xml)
	if err != nil {
		return nil, err
	}

	return convertEntitySetToMetadataSet(entities), nil
}

// collects field properties and groups them in entities, other data in XML is ignored.
func extractEntities(root *xquery.XML) (models.EntitySet, error) {
	querySchema := fmt.Sprintf("//edmx:DataServices/Schema[@Namespace='%v']", GraphNamespaceSchemaName)

	schema := root.FindOne(querySchema)
	if schema.IsEmpty() {
		return nil, ErrMissingSchema
	}

	entities := models.NewEntitySet()

	// List all entity types.
	queryListAllEntityTypes := fmt.Sprintf(
		"//edmx:DataServices/Schema[@Namespace='%v']/EntityType", GraphNamespaceSchemaName)
	for _, entityType := range root.FindMany(queryListAllEntityTypes) {
		entityName := entityType.Attr("Name")
		parentName := entityType.Attr("BaseType")
		isAbstract := entityType.Attr("Abstract") == "true"

		entity := entities.GetOrCreate(entityName, parentName, isAbstract)

		// Add properties.
		for _, property := range entityType.FindMany("Property") {
			if property.Attr("Name") != "" {
				entity.AddProperty(metadatadef.Field{
					Name:         property.Attr("Name"),
					Type:         property.Attr("Type"),
					ValueOptions: nil,
				})
			}
		}
	}

	// link every child with parent completing hierarchy
	schemaAlias := schema.Attr("Alias")
	if err := entities.MatchParentsWithChildren(schemaAlias); err != nil {
		return nil, errors.Join(ErrMetadataProcessing, err)
	}

	return entities.FilterAbstract(), nil
}

// Select entities that match entity names of interest.
// Every property has display identical to itself.
func convertEntitySetToMetadataSet(entities models.EntitySet) metadatadef.Schemas[any] {
	result := make(metadatadef.Schemas[any], len(entities))

	index := 0

	for name, entity := range entities {
		displayName := api3.CamelCaseToSpaceSeparated(name)
		displayName = api3.CapitalizeFirstLetterEveryWord(displayName)
		displayName = naming.NewPluralString(displayName).String()

		objectName := naming.NewPluralString(name).String()

		result[index] = metadatadef.Schema{
			ObjectName:  objectName,
			DisplayName: displayName,
			Fields:      entity.GetAllProperties(),
			URLPath:     naming.NewPluralString(name).String(),
			ResponseKey: "",
		}

		index++
	}

	return result
}
