package hubspot

import (
	"context"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
)

// Type definitions for HubSpot associations API.

type assocInputs struct {
	Inputs []assocId `json:"inputs"`
}

type assocId struct {
	Id string `json:"id"`
}

type assocType struct {
	Category string  `json:"category"`
	TypeId   int     `json:"typeId"`
	Label    *string `json:"label"`
}

type assocObject struct {
	ToObjectId       int64       `json:"toObjectId"`
	AssociationTypes []assocType `json:"associationTypes"`
}

type assocResult struct {
	From assocId       `json:"from"`
	To   []assocObject `json:"to"`
}

type assocOutput struct {
	Status  string        `json:"status"`
	Results []assocResult `json:"results"`
}

// String returns a string representation of the association type.
func (t *assocType) String() string {
	if t.Label != nil && len(*t.Label) > 0 {
		return fmt.Sprintf("category=%s id=%d label=%s", t.Category, t.TypeId, *t.Label)
	}

	return fmt.Sprintf("category=%s id=%d", t.Category, t.TypeId)
}

// getUniqueIds returns a slice of unsorted unique IDs from the given data.
func getUniqueIds(data *[]common.ReadResultRow) []string {
	uniqueIds := make(map[string]struct{})

	for _, row := range *data {
		uniqueIds[row.Id] = struct{}{}
	}

	var ids []string

	for id := range uniqueIds {
		ids = append(ids, id)
	}

	return ids
}

// fillAssociations fills the associations for the given object names and data.
func (c *Connector) fillAssociations(ctx context.Context, fromObjName string, data *[]common.ReadResultRow, toAssociatedObjects []string) error {
	ids := getUniqueIds(data)

	for _, associatedObject := range toAssociatedObjects {
		associations, err := c.getObjectAssociations(ctx, fromObjName, ids, associatedObject)
		if err != nil {
			return err
		}

		if len(associations) == 0 {
			continue
		}

		for i, row := range *data {
			if assocs, ok := associations[row.Id]; ok {
				if (*data)[i].Associations == nil {
					(*data)[i].Associations = make(map[string][]common.Association)
				}

				(*data)[i].Associations[associatedObject] = assocs
			}
		}
	}

	return nil
}

// getObjectAssociations returns the associations for the given object names and IDs. It returns
// a mapping of object IDs to their associations.
func (c *Connector) getObjectAssociations(
	ctx context.Context,
	fromObject string,
	fromIds []string,
	toObject string,
) (map[string][]common.Association, error) {
	if len(fromIds) == 0 {
		return nil, nil
	}

	u, err := c.getURL(fmt.Sprintf("/crm/v4/associations/%s/%s/batch/read", fromObject, toObject))
	if err != nil {
		return nil, err
	}

	var inputs assocInputs

	for _, id := range fromIds {
		inputs.Inputs = append(inputs.Inputs, assocId{Id: id})
	}

	// Do one big batch request to get all associations.
	// See https://developers.hubspot.com/docs/guides/api/crm/associations/associations-v4#retrieve-associated-records
	rsp, err := c.Client.Post(ctx, u, &inputs)
	if err != nil {
		return nil, fmt.Errorf("error fetching HubSpot associations: %w", err)
	}

	output, err := common.UnmarshalJSON[assocOutput](rsp)
	if err != nil {
		return nil, err
	}

	out := make(map[string][]common.Association)

	for _, result := range output.Results {
		var assocs []common.Association

		for _, assoc := range result.To {
			for _, t := range assoc.AssociationTypes {
				assocs = append(assocs, common.Association{
					ObjectID:        strconv.FormatInt(assoc.ToObjectId, 10),
					AssociationType: t.String(),
				})
			}
		}

		if len(assocs) > 0 {
			out[result.From.Id] = assocs
		}
	}

	return out, nil
}
