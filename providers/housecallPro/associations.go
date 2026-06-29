package housecallpro

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
)

const customerAssociation = "customer"

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
	for _, assocObj := range associatedObjects {
		if objectName == "jobs" && assocObj == customerAssociation {
			attachJobCustomer(rows)
		}
	}
}

// attachJobCustomer attaches the customer object embedded on each job to its
// Associations. Housecall Pro returns job.customer inline on GET /jobs.
func attachJobCustomer(rows []common.ReadResultRow) {
	for idx := range rows {
		customer, ok := rows[idx].Raw[customerAssociation].(map[string]any)
		if !ok || len(customer) == 0 {
			continue
		}

		id, _ := customer["id"].(string)
		if id == "" {
			continue
		}

		if rows[idx].Associations == nil {
			rows[idx].Associations = make(map[string][]common.Association)
		}

		rows[idx].Associations[customerAssociation] = []common.Association{{
			ObjectId: id,
			Raw:      customer,
		}}
	}
}
