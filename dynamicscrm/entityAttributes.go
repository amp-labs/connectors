package dynamicscrm

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/spyzhov/ajson"
)

// UnderscoreFieldFormat is used to format field names that will be present in Read response.
// These fields are references used for search.
const UnderscoreFieldFormat = "_%v_value"

var (
	ErrObjectNotFound          = errors.New("object not found")
	ErrObjectMissingAttributes = errors.New("object missing metadata attributes")
)

// Returns pairs of field names to display names.
// Internally will make an API call to Attributes endpoint.
func (c *Connector) getFieldsForObject(
	ctx context.Context, objectName naming.SingularString,
) (map[string]string, error) {
	url, err := c.getEntityAttributesURL(objectName)
	if err != nil {
		return nil, err
	}

	// Filter attributes to ensure they are:
	// 1. Present in Read responses (IsValidODataAttribute == true)
	// 2. Can be queried in GET requests (IsValidForRead == true)
	// This ensures we only work with fields that are both
	// returned in the payload and can be used in query parameters.
	url.WithQueryParam("$filter", "(IsValidODataAttribute eq true and IsValidForRead eq true)")
	// We cannot use $select clause to scope response, unfortunately, `Targets` field breaks $select.
	// Falling back to requesting the whole payload.

	body, err := c.performGetRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	fields, err := extractFieldsFromJSON(body)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, objectName)
	}

	return fields, nil
}

// Attributes from endpoint response will be converted from JSON objects
// to field names mapped to display names.
func extractFieldsFromJSON(attributes *ajson.Node) (map[string]string, error) {
	array, err := jsonquery.New(attributes).Array("value", false)
	if err != nil {
		return nil, errors.Join(ErrObjectNotFound, err)
	}

	if len(array) == 0 {
		// nothing to read, we expected some attributes
		return nil, ErrObjectMissingAttributes
	}

	fieldsMap := make(map[string]string)

	for _, item := range array {
		name, displayName, err := getAttributeNames(item)
		if err != nil {
			return nil, err
		}

		fieldsMap[name] = displayName
	}

	return fieldsMap, nil
}

// Single attribute payload holds logical name of a field and its display name.
func getAttributeNames(attribute *ajson.Node) (name string, displayName string, err error) {
	logicalName, err := jsonquery.New(attribute).Str("LogicalName", false)
	if err != nil {
		return "", "", errors.Join(ErrObjectNotFound, err)
	}

	name = *logicalName

	displayName, err = getAttributeDisplayName(attribute, name)
	if err != nil {
		return "", "", errors.Join(ErrObjectNotFound, err)
	}

	// check if attribute has targets
	targets, err := jsonquery.New(attribute).Array("Targets", true)
	if err != nil {
		return "", "", errors.Join(ErrObjectNotFound, err)
	}

	if len(targets) > 0 {
		// This field is a reference to other entities.
		// Apply underscore formating, because this is how such fields appear in the Read response.
		name = fmt.Sprintf(UnderscoreFieldFormat, name)
	}

	return name, displayName, nil
}

// Display Name is picked based on priority list. Last element is the least preferred fallback.
//
// Below is the priority list using EntityDefinitions.Attribute object.
//
// 1. DisplayName.LocalizedLabels[0].Label -> ex: Entity Image Id
// 2. SchemaName -> ex: EntityImageId
// 3. LogicalName -> ex: entityimageid.
func getAttributeDisplayName(item *ajson.Node, logicalName string) (string, error) {
	localizedLabels, err := jsonquery.New(item, "DisplayName").Array("LocalizedLabels", true)
	if err != nil {
		return "", err
	}

	if len(localizedLabels) != 0 {
		// First occurring label should be sufficient to get to know display name.
		firstLabel := localizedLabels[0]

		displayLabel, err := jsonquery.New(firstLabel).Str("Label", true)
		if err != nil {
			return "", err
		}

		if displayLabel != nil {
			return *displayLabel, nil
		}
	}

	// try to use SchemaName which has better format than logical name
	return jsonquery.New(item).StrWithDefault("SchemaName", logicalName)
}
