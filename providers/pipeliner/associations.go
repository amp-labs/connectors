package pipeliner

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
)

func fillAssociations(params common.ReadParams) readhelper.RowMarshallProcessor {
	return func(rows []common.ReadResultRow) error {
		for index := range rows {
			rows[index].Associations = make(map[string][]common.Association)

			for _, associatedObjectName := range params.AssociatedObjects {
				associationData := rows[index].Raw[associatedObjectName]

				associationRaw, ok := associationData.(map[string]any) // nolint:varnamelen
				if !ok {
					return errors.New("associationsData is not a map") // nolint:err113
				}

				identifierData, found := associationRaw["id"]
				if !found {
					return errors.New("associations data does not have an id") // nolint:err113
				}

				identifier, ok := identifierData.(string)
				if !ok {
					return fmt.Errorf("failed to convert [%v] to string", identifierData) // nolint:err113
				}

				rows[index].Associations[associatedObjectName] = []common.Association{
					{
						ObjectId:        identifier,
						AssociationType: associatedObjectName,
						Raw:             associationRaw,
					},
				}
			}
		}

		return nil
	}
}
