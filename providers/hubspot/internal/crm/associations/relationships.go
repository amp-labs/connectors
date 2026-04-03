package associations

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
)

// fetchObjectSchema retrieves the schema for the given object type from the Hubspot API.
// It returns the schema, an optional warning for failed request, and an error for bad implementation.
func (s Strategy) fetchObjectSchema(
	ctx context.Context, objectName common.ObjectName,
) (schema *ObjectSchemaResponse, warning error, err error) {
	url, err := s.getReadObjectSchema(objectName.String())
	if err != nil {
		return nil, nil, err // nolint:nilnil
	}

	resp, warning := s.clientCRM.Get(ctx, url.String())
	if warning != nil {
		return nil, warning, nil
	}

	response, err := common.UnmarshalJSON[ObjectSchemaResponse](resp)

	return response, nil, err
}

// objectRelationship represents a directed relationship between two object types.
//
// Note: Hubspot uses non-human readable names called object type id.
//
//	Example: "0-1" is a "contact".
type objectRelationship struct {
	// from is an object name or object type id.
	from string
	// to is an object name or object type id.
	to string
}

type ObjectSchemaResponse struct {
	Name         string                            `json:"name"`
	ObjectTypeID string                            `json:"objectTypeId"`
	Associations []ObjectSchemaAssociationResponse `json:"associations"`
}

// ObjectSchemaAssociationResponse represents association between 2 object types.
// Hubspot defined association can be found here:
// https://developers.hubspot.com/docs/api-reference/crm-associations-v4/guide#association-type-id-values
type ObjectSchemaAssociationResponse struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	FromObjectTypeID string `json:"fromObjectTypeId"`
	ToObjectTypeID   string `json:"toObjectTypeId"`
}

// lookupRelationship finds the objectRelationship corresponding to the given association type ID.
// Returns nil if relationship is not found.
func (r ObjectSchemaResponse) lookupRelationship(identifier int) *objectRelationship {
	identifierStr := strconv.Itoa(identifier)

	for _, association := range r.Associations {
		if association.ID == identifierStr {
			return &objectRelationship{
				from: association.FromObjectTypeID,
				to:   association.ToObjectTypeID,
			}
		}
	}

	// Relationship found.
	return nil
}
