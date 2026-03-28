package associations

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/codec"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
)

var ErrUnresolvedAssociationTypeID = errors.New("missing association types")

// ParseInput converts the generic input (of type any) into a BatchInput
// expected by the connector. The input usually comes from the caller as
// untyped data (e.g., a map[string]any or JSON payload), and this function
// decodes it into a concrete BatchInput structure.
func ParseInput(input any) (*BatchInput, error) {
	if input == nil {
		return goutils.Pointer(make(BatchInput, 0)), nil
	}

	return codec.Parse[*BatchInput](input)
}

// BatchInput is the data structure expected by the connector when performing
// batch create or batch update for an object.
// It represents a list of individual association definitions.
//
// See example of association payload used for Deals object:
// https://developers.hubspot.com/docs/api-reference/crm-deals-v3/batch/post-crm-v3-objects-0-3-batch-create
type BatchInput []InputDefinition

type InputDefinition struct {
	To    Identifier `json:"to"`
	Types []Type     `json:"types,omitempty"`
}

type Identifier struct {
	ID string `json:"id"`
}

type Type struct {
	Category string `json:"associationCategory"`
	ID       int    `json:"associationTypeId"`
}

// BatchCreatePayload is the request payload for the HubSpot batch create associations.
//
//	Endpoint:
//		POST /crm/v4/associations/{fromObjectType}/{toObjectType}/batch/create
//
// Each payload contains a list of associations to create in a single call.
//
// Reference:
// nolint:lll
// https://developers.hubspot.com/docs/api-reference/crm-associations-v4/batch/post-crm-v4-associations-fromObjectType-toObjectType-batch-create
type BatchCreatePayload struct {
	Inputs []CreateDefinition `json:"inputs"`
}

func (p *BatchCreatePayload) AddEntry(def *CreateDefinition) {
	p.Inputs = append(p.Inputs, *def)
}

// newBatchCreatePayload returns a new BatchCreatePayload.
func newBatchCreatePayload() *BatchCreatePayload {
	return &BatchCreatePayload{Inputs: make([]CreateDefinition, 0)}
}

// CreateDefinition represents a single association between two objects.
// It specifies the source object (From), the target object (To), and association types (Types).
type CreateDefinition struct {
	From  Identifier `json:"from"`
	To    Identifier `json:"to"`
	Types []Type     `json:"types"`
}

// BatchCreateParams collects and organizes batch association inputs for Strategy.BatchCreate.
// It groups associations by relationship (fromObjectType -> toObjectType) and builds
// one BatchCreatePayload per relationship. Use BatchCreateParams.WithAssociation to build params.
type BatchCreateParams struct {
	// payloads holds BatchCreatePayloads keyed by associationTypeId.
	//
	// Each payload corresponds to a specific relationship (fromObjectType -> toObjectType);
	// the key is one of the association type IDs from the payload’s Types.
	payloads map[int]*BatchCreatePayload
}

func NewBatchCreateParams() *BatchCreateParams {
	return &BatchCreateParams{
		payloads: make(map[int]*BatchCreatePayload),
	}
}

func (t *BatchCreateParams) WithAssociation(fromObjectID string, associationsList *BatchInput) {
	if len(*associationsList) == 0 {
		return
	}

	// Add every association of this object to a combined list of associations.
	for _, association := range *associationsList {
		if len(association.Types) == 0 {
			continue
		}

		// Use the first type's ID to identify the relationship (ObjectType1 -> ObjectType2).
		// All identifiers in one association share the same relationship, the list of types
		// exists to hold extra metadata like relationship labels.
		identifier := association.Types[0].ID

		// Ensure payload for this relationship is defined.
		if _, ok := t.payloads[identifier]; !ok {
			t.payloads[identifier] = newBatchCreatePayload()
		}

		// Add association to this payload.
		t.payloads[identifier].AddEntry(&CreateDefinition{
			From: Identifier{
				ID: fromObjectID,
			},
			To:    association.To,
			Types: association.Types,
		})
	}
}

// makePayloadsForRelationships converts the internal payloads map into a registry
// of objectRelationship -> BatchCreatePayload, using the provided (from) objectName and it's schema.
//
// It returns:
//   - registry: a mapping from each resolved relationship to its payload.
//   - warning: an error if any associationTypeId could not be found in the schema.
//
// Note: Unresolved type IDs are collected into a warning message rather than hard‑stopping.
func (t *BatchCreateParams) makePayloadsForRelationships(
	objectName common.ObjectName, schema *ObjectSchemaResponse,
) (registry map[objectRelationship]*BatchCreatePayload, warning error) {
	registry = make(map[objectRelationship]*BatchCreatePayload, len(t.payloads))

	unresolvedIdentifiers := make([]int, 0)

	for identifier, payload := range t.payloads {
		if relationship := schema.lookupRelationship(identifier); relationship == nil {
			unresolvedIdentifiers = append(unresolvedIdentifiers, identifier)
		} else {
			registry[*relationship] = payload
		}
	}

	if len(unresolvedIdentifiers) != 0 {
		identifiersMessage := strings.Join(
			datautils.ForEach(unresolvedIdentifiers, strconv.Itoa), // list of strings
			", ", // separator
		)

		warning = fmt.Errorf("%w: objectName(%v) indetifiers(%v)",
			ErrUnresolvedAssociationTypeID, objectName, identifiersMessage)
	}

	return registry, warning
}
