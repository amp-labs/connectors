package microsoftdynamicscrm

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/subchen/go-xmldom"
)

var (
	CRMMetadataSchemaName = "Microsoft.Dynamics.CRM" // nolint:gochecknoglobals

	ErrMissingSchema      = fmt.Errorf("missing schema %v in response", CRMMetadataSchemaName)
	ErrMetadataProcessing = errors.New("metadata couldn't be processed")
	ErrObjectNotFound     = errors.New("object not found")
)

// Please note: MSDynamics API does not return proper display names for objects and fields,
// so the ListObjectMetadataResult will have display names that look like "accountleads".
func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames is not empty
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	rsp, err := c.getXML(ctx, c.getURL("$metadata"))
	if err != nil {
		return nil, err
	}

	root, err := rsp.GetRoot()
	if err != nil {
		return nil, err
	}

	entities, err := extractEntities(root)
	if err != nil {
		return nil, err
	}

	result, err := convertEntitySetToMetadataSet(objectNames, entities)
	if err != nil {
		return nil, err
	}

	return &common.ListObjectMetadataResult{
		Result: result,
		Errors: nil,
	}, nil
}

// collects field properties and groups them in entities, other data in XML is ignored.
func extractEntities(root *xmldom.Node) (EntitySet, error) {
	querySchema := fmt.Sprintf("/DataServices/Schema[@Namespace='%v']", CRMMetadataSchemaName)

	schema := root.QueryOne(querySchema)
	if schema == nil {
		return nil, ErrMissingSchema
	}

	entities := NewEntitySet()
	// List all field properties that exist for current schema
	queryListAllSchemaProperties := fmt.Sprintf(
		"/DataServices/Schema[@Namespace='%v']/EntityType[*]/Property/@Name", CRMMetadataSchemaName)
	root.QueryEach(queryListAllSchemaProperties, func(index int, property *xmldom.Node) {
		// parent of a property is an Entity.
		// Entity may inherit properties from a parent
		// We save entity name and the name of its parent, so later we can infer all properties by denormalisation
		entityName := property.Parent.GetAttributeValue("Name")
		parentName := property.Parent.GetAttributeValue("BaseType")
		entity := entities.GetOrCreate(entityName, parentName)
		propertyName := property.GetAttributeValue("Name")
		entity.AddProperty(propertyName)
	})

	queryListAbstractEntities := fmt.Sprintf(
		"/DataServices/Schema[@Namespace='%v']/EntityType[@Abstract='true']", CRMMetadataSchemaName)
	root.QueryEach(queryListAbstractEntities, func(index int, abstractEntity *xmldom.Node) {
		if len(abstractEntity.Children) == 0 {
			// these entities were not included by previous query as they have no properties
			// we programmatically find these special types, which are "primary values" but for structs
			// Ex: crmbaseentity, crmmodelbaseentity,
			entityName := abstractEntity.GetAttributeValue("Name")
			// effectively only create
			_ = entities.GetOrCreate(entityName, "")
		}
	})

	// link every child with parent completing hierarchy
	schemaAlias := schema.GetAttributeValue("Alias")
	if err := entities.MatchParentsWithChildren(schemaAlias); err != nil {
		return nil, errors.Join(ErrMetadataProcessing, err)
	}

	return entities, nil
}

// Select entities that match entity names of interest.
// Every property has display identical to itself.
func convertEntitySetToMetadataSet(names []string, entities EntitySet) (map[string]common.ObjectMetadata, error) {
	result := map[string]common.ObjectMetadata{}

	for _, name := range names {
		entity, ok := entities[name]
		if !ok {
			return nil, fmt.Errorf("unknown entity %v %w", name, ErrObjectNotFound)
		}

		properties := entity.GetAllProperties()
		fieldsMap := make(map[string]string)

		for _, p := range properties {
			fieldsMap[p] = p
		}

		result[name] = common.ObjectMetadata{
			DisplayName: name,
			FieldsMap:   fieldsMap,
		}
	}

	return result, nil
}
