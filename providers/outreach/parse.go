package outreach

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func getRecords(node *ajson.Node) ([]*ajson.Node, error) {
	return jsonquery.New(node).ArrayOptional(dataKey)
}

func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node, "links").StrWithDefault("next", "")
}

func (c *Connector) getOutreachDataMarshaller(ctx context.Context, config common.ReadParams,
	transformer common.RecordTransformer,
) common.MarshalFromNodeFunc {
	return func(records []*ajson.Node, fields []string) ([]common.ReadResultRow, error) {
		data := make([]common.ReadResultRow, len(records))

		// Go through each record, attach associations (if any) to the record
		// and convert the record to a common.ReadResultRow.
		for idx, nodeRecord := range records {
			raw, err := jsonquery.Convertor.ObjectToMap(nodeRecord)
			if err != nil {
				return nil, err
			}

			record, err := transformer(nodeRecord)
			if err != nil {
				return nil, err
			}

			idF, ok := raw["id"].(float64)
			if !ok {
				return nil, errors.New("invalid or missing id field") // nolint: err113
			}

			idStr := strconv.Itoa(int(idF))

			data[idx] = common.ReadResultRow{
				Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
				Raw:    raw,
				Id:     idStr,
			}

			if len(config.AssociatedObjects) > 0 {
				assoc, err := c.fetchAssociations(ctx, idStr, config.ObjectName, config.AssociatedObjects)
				if err != nil {
					return nil, err
				}

				data[idx].Associations = assoc
			}
		}

		return data, nil
	}
}

type AssocData struct {
	Included []map[string]any `json:"included"`
}

// fetchAssociations fetches the list of associated objects from the outreach API.
func (c *Connector) fetchAssociations(ctx context.Context, id string, objectName string,
	assoc []string,
) (map[string][]common.Association, error) {
	associations := make(map[string][]common.Association)

	url, err := urlbuilder.New(c.BaseURL, apiVersion, objectName, id)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("include", strings.Join(assoc, ","))

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	data, err := common.UnmarshalJSON[AssocData](resp)
	if err != nil {
		return nil, err
	}

	for _, record := range data.Included {
		recordId, ok := record["id"].(float64)
		if !ok {
			return nil, errors.New("objectID expected to be a number") //nolint: err113
		}

		assocType, ok := record["type"].(string)
		if !ok {
			return nil, errors.New("object type expected to be a string") //nolint: err113
		}

		associations[assocType] = append(associations[assocType], common.Association{
			ObjectId:        strconv.Itoa(int(recordId)),
			AssociationType: assocType,
			Raw:             record,
		})
	}

	return associations, nil
}
