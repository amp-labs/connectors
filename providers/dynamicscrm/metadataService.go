package dynamicscrm

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
)

// UnderscoreFieldFormat is used to format field names that will be present in Read response.
// These fields are references used for search.
const UnderscoreFieldFormat = "_%v_value"

var (
	ErrObjectNotFound          = errors.New("object not found")
	ErrObjectMissingAttributes = errors.New("object missing metadata attributes")
)

// Make a call to EntityDefinition endpoint.
// We are looking for one field: DisplayCollectionName.
// This is the formal name of a collection that this ObjectName represents.
func (c *Connector) fetchObjectDisplayName(
	ctx context.Context, objectName naming.SingularString,
) (string, error) {
	entityDefinition, err := c.metadataDiscoveryRepository.fetchEntityDefinition(ctx, objectName)
	if err != nil {
		return "", err
	}

	displayName, ok := entityDefinition.DisplayCollectionName.getName()
	if !ok {
		// There is no localized display name.
		// Default to object name pluralised, that's the best we can do.
		return objectName.Plural().String(), nil
	}

	return displayName, nil
}

// Returns fields metadata.
// Internally will make an API call to Attributes endpoint.
func (c *Connector) fetchFieldsForObject(
	ctx context.Context, objectName naming.SingularString,
) (map[string]common.FieldMetadata, error) {
	attributes, err := c.metadataDiscoveryRepository.fetchAttributes(ctx, objectName)
	if err != nil {
		return nil, err
	}

	if len(attributes.Values) == 0 {
		// nothing to read, we expected some attributes
		return nil, fmt.Errorf("%w: %s", ErrObjectMissingAttributes, objectName)
	}

	var optionsErr error

	attributesPicklists, err := c.metadataDiscoveryRepository.fetchAttributesPicklists(ctx, objectName)
	if err != nil {
		attributesPicklists = &attributesPicklistsResponse{}
		optionsErr = errors.Join(optionsErr, err)
	}

	attributesStatuses, err := c.metadataDiscoveryRepository.fetchAttributesStatuses(ctx, objectName)
	if err != nil {
		attributesStatuses = &attributesStatusesResponse{}
		optionsErr = errors.Join(optionsErr, err)
	}

	attributesStates, err := c.metadataDiscoveryRepository.fetchAttributesStates(ctx, objectName)
	if err != nil {
		attributesStates = &attributesStatesResponse{}
		optionsErr = errors.Join(optionsErr, err)
	}

	return combineAttributesMetadata(attributes, attributesPicklists, attributesStatuses, attributesStates), optionsErr
}

// Attributes response will be converted to FieldMetadata.
// However, this data is not enough. We need to list enumeration options for some fields.
// Additional responses each contains a set of options.
func combineAttributesMetadata(
	attributes *attributesResponse, attributePicklists *attributesPicklistsResponse,
	attributeStatuses *attributesStatusesResponse, attributeStates *attributesStatesResponse,
) map[string]common.FieldMetadata {
	// Regardless of the attribute type merge them into single registry.
	// It is a `list of field values` each named after an `attribute`.
	attributeOptions := datautils.MergeNamedLists(
		attributePicklists.getOptionsPerAttribute(),
		attributeStatuses.getOptionsPerAttribute(),
		attributeStates.getOptionsPerAttribute(),
	)

	fieldsMap := make(map[string]common.FieldMetadata)

	for _, item := range attributes.Values {
		name := item.getName()
		modifiable := item.IsValidForCreate || item.IsValidForUpdate
		valueType := item.getValueType()
		values := attributeOptions[item.LogicalName]

		fieldsMap[name] = common.FieldMetadata{
			DisplayName:  item.getDisplayName(),
			ValueType:    valueType,
			ProviderType: item.AttributeTypeName.Value,
			ReadOnly:     goutils.Pointer(!modifiable),
			Values:       values,
		}
	}

	return fieldsMap
}

func (item attributeItem) getName() string {
	// check if attribute has targets
	if len(item.Targets) > 0 {
		// This field is a reference to other entities.
		// Apply underscore formating, because this is how such fields appear in the Read response.
		return fmt.Sprintf(UnderscoreFieldFormat, item.LogicalName)
	}

	return item.LogicalName
}

// Display Name is picked based on priority list. Last element is the least preferred fallback.
//
// Below is the priority list using EntityDefinitions.Attribute object.
//
// 1. DisplayName.LocalizedLabels[0].Label -> ex: Entity Image Id
// 2. SchemaName -> ex: EntityImageId
// 3. LogicalName -> ex: entityimageid.
func (item attributeItem) getDisplayName() string {
	labels := item.DisplayName.LocalizedLabels

	if len(labels) != 0 {
		// First occurring label should be sufficient to get to know display name.
		firstLabel := labels[0]

		displayLabel := firstLabel.Label

		if len(displayLabel) != 0 {
			return displayLabel
		}
	}

	// try to use SchemaName which has better format than logical name
	name := item.SchemaName
	if len(name) != 0 {
		return name
	}

	return item.LogicalName
}

// nolint:lll
// Based on the attribute type infer value type.
// https://learn.microsoft.com/en-us/dynamics365/customerengagement/on-premises/developer/introduction-to-entity-attributes?view=op-9-1#types-of-attributes
func (item attributeItem) getValueType() common.ValueType { // nolint:cyclop
	switch item.AttributeTypeName.Value {
	case "StringType", "MemoType":
		return common.ValueTypeString
	case "BooleanType":
		return common.ValueTypeBoolean
	case "BigIntType", "IntegerType":
		return common.ValueTypeInt
	case "DecimalType", "MoneyType":
		return common.ValueTypeFloat
	case "DateTimeType":
		// https://learn.microsoft.com/en-us/dynamics365/customerengagement/on-premises/developer/introduction-to-entity-attributes?view=op-9-1#date-and-time-data-attribute
		switch item.Format {
		case "DateAndTime", "UserLocal":
			return common.ValueTypeDateTime
		case "DateOnly":
			return common.ValueTypeDate
		default:
			return common.ValueTypeOther
		}
	case "PicklistType", "StatusType", "StateType":
		return common.ValueTypeSingleSelect
	default:
		// Examples: EntityNameType or
		//
		// ImageType.
		// https://learn.microsoft.com/en-us/dynamics365/customerengagement/on-premises/developer/introduction-to-entity-attributes?view=op-9-1#image-data-attributes
		//
		// CustomerType, LookupType, OwnerType.
		// https://learn.microsoft.com/en-us/dynamics365/customerengagement/on-premises/developer/introduction-to-entity-attributes?view=op-9-1#reference-data-attributes
		//
		// UniqueidentifierType.
		// https://learn.microsoft.com/en-us/dynamics365/customerengagement/on-premises/developer/introduction-to-entity-attributes?view=op-9-1#unique-identifier-data-attributes
		//
		// Virtual attributes
		// https://learn.microsoft.com/en-us/dynamics365/customerengagement/on-premises/developer/introduction-to-entity-attributes?view=op-9-1#virtual-attributes
		//
		// Logical attributes.
		// https://learn.microsoft.com/en-us/dynamics365/customerengagement/on-premises/developer/introduction-to-entity-attributes?view=op-9-1#logical-attributes
		return common.ValueTypeOther
	}
}
