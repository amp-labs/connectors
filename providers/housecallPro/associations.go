package housecallpro

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
)

const customerObject = "customers"

var embeddedAssociationFields = map[string]map[string]string{ //nolint:gochecknoglobals
	"jobs": {
		customerObject: "customer",
	},
}

func readMarshaller(params common.ReadParams) common.MarshalFromNodeFunc {
	base := readhelper.MakeMarshaledDataFuncWithId(nil, readIDFieldByObject.Get(params.ObjectName))
	if len(params.AssociatedObjects) == 0 {
		return base
	}

	return readhelper.ChainedMarshaller(base, func(rows []common.ReadResultRow) error {
		extractAssociations(params.ObjectName, params.AssociatedObjects, rows)

		return nil
	})
}

// extractAssociations attaches related objects to rows for each requested association.
func extractAssociations(objectName string, associatedObjects []string, rows []common.ReadResultRow) {
	fieldByAssociation, ok := embeddedAssociationFields[objectName]
	if !ok {
		return
	}

	for _, assocObj := range associatedObjects {
		field, ok := fieldByAssociation[assocObj]
		if !ok {
			continue
		}

		attachEmbeddedAssociation(rows, assocObj, field)
	}
}

// attachEmbeddedAssociation attaches the object embedded under field on each row
// to its Associations, keyed by the requested association name.
func attachEmbeddedAssociation(rows []common.ReadResultRow, associationName, field string) {
	for idx := range rows {
		embedded, ok := rows[idx].Raw[field].(map[string]any)
		if !ok || len(embedded) == 0 {
			continue
		}

		id, _ := embedded["id"].(string)
		if id == "" {
			continue
		}

		if rows[idx].Associations == nil {
			rows[idx].Associations = make(map[string][]common.Association)
		}

		rows[idx].Associations[associationName] = []common.Association{{
			ObjectId: id,
			Raw:      embedded,
		}}
	}
}
