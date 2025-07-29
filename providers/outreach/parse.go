package outreach

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func getRecords(node *ajson.Node) ([]*ajson.Node, error) {
	return jsonquery.New(node).ArrayOptional(dataKey)
}

func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node, "links").StrWithDefault("next", "")
}

func (c *Connector) getOutreachDataMarshaller(ctx context.Context, assoc []string, //nolint: gocognit,cyclop,funlen
	transformer common.RecordTransformer,
) common.MarshalFromNodeFunc {
	return func(records []*ajson.Node, fields []string) ([]common.ReadResultRow, error) {
		data := make([]common.ReadResultRow, len(records))

		// Go through each record, attach associations (if any) to the record
		// and convert the record to a common.ReadResultRow.
		for idx, nodeRecord := range records {
			associations := make(map[string][]common.Association)

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

			for _, assoc := range assoc {
				relationship, ok := raw[relationshipsKey].(map[string]any)
				if !ok {
					return nil, errors.New("relationship key not found ") // nolint: err113
				}

				assocObjectName, typ, err := extractAssociationKey(relationship, assoc)
				if err != nil {
					return nil, err
				}

				assocList, err := c.fetchAssociations(ctx, assocObjectName, typ, relationship)
				if err != nil {
					return nil, err
				}

				if len(assocList) > 0 {
					associations[assoc] = assocList
				}
			}

			data[idx] = common.ReadResultRow{
				Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
				Raw:    raw,
				Id:     idStr,
			}

			if len(associations) > 0 {
				data[idx].Associations = associations
			}
		}

		return data, nil
	}
}

func extractAssociationKey(relationships map[string]any, reqAssoc string) (string, []string, error) { //nolint: cyclop
	var (
		singularObjectName string
		assocTypes         []string
	)

	for typ, relRecord := range relationships {
		data, err := assertMapStringAny(relRecord)
		if err != nil {
			return "", nil, fmt.Errorf("relationship %s: %w", typ, err)
		}

		records, exists := data[dataKey]
		if !exists {
			continue
		}

		switch rcds := records.(type) {
		case nil:
			continue
		case map[string]any:
			if typeStr, ok := rcds["type"].(string); ok {
				singularObjectName = typeStr
			} else {
				return "", nil, errors.New("missing or invalid 'type'") //nolint: err113
			}
		case []any:
			if len(rcds) == 0 {
				continue
			}

			firstRecord, err := assertMapStringAny(rcds[0])
			if err != nil {
				return "", nil, fmt.Errorf("first record: %w", err)
			}

			if typeStr, ok := firstRecord["type"].(string); ok {
				singularObjectName = typeStr
			} else {
				return "", nil, errors.New("missing or invalid 'type'") //nolint: err113
			}
		default:
			continue
		}

		associatedObjectName := naming.NewPluralString(singularObjectName).String()
		if associatedObjectName == reqAssoc {
			assocTypes = append(assocTypes, typ)
		}
	}

	return reqAssoc, assocTypes, nil
}

func (c *Connector) fetchAssociations(ctx context.Context, objectName string, keys []string,
	record map[string]any,
) ([]common.Association, error) {
	var assoc []common.Association

	for _, key := range keys {
		data, err := assertMapStringAny(record[key])
		if err != nil {
			return nil, fmt.Errorf("key %s: %w", key, err)
		}

		records, exists := data[dataKey]
		if !exists {
			return nil, fmt.Errorf("the requested associated object %s was not found", objectName) //nolint: err113
		}

		// Handle single association
		if rec, err := assertMapStringAny(records); err == nil {
			asc, err := c.fetchSingleAssociation(ctx, objectName, key, rec)
			if err != nil {
				return nil, err
			}

			assoc = append(assoc, asc...)

			continue
		}

		// Handle multiple associations
		if recs, err := assertSliceMapStringAny(records); err == nil {
			asc, err := c.fetchMultipleAssociations(ctx, objectName, key, recs)
			if err != nil {
				return nil, err
			}

			assoc = append(assoc, asc...)

			continue
		}

		return nil, fmt.Errorf("unexpected data type for associations of %s", objectName) // nolint: err113
	}

	return assoc, nil
}

func (c *Connector) fetchSingleAssociation(ctx context.Context, objectName string,
	key string, rel map[string]any,
) ([]common.Association, error) {
	recordId, ok := rel[idKey].(float64)
	if !ok {
		return nil, errors.New("unexpected association recordId data type") //nolint:err113
	}

	path := objectName + "/" + strconv.Itoa(int(recordId))

	records, err := c.getAssociation(ctx, path)
	if err != nil {
		return nil, err
	}

	assoc := common.Association{
		ObjectId:        strconv.Itoa(int(recordId)),
		AssociationType: key,
		Raw:             records,
	}

	return []common.Association{assoc}, nil
}

func (c *Connector) fetchMultipleAssociations(ctx context.Context, objectName string,
	key string, relationships []map[string]any,
) ([]common.Association, error) {
	var result []common.Association

	for _, rcd := range relationships {
		recod, err := c.fetchSingleAssociation(ctx, objectName, key, rcd)
		if err != nil {
			return nil, err
		}

		result = append(result, recod...)
	}

	return result, nil
}

type Records struct {
	Data map[string]any `json:"data"`
}

func (c *Connector) getAssociation(ctx context.Context, path string) (map[string]any, error) {
	u, err := c.getApiURL(path)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Get(ctx, u.String())
	if err != nil {
		return nil, err
	}

	d, err := common.UnmarshalJSON[Records](resp)
	if err != nil {
		return nil, err
	}

	return d.Data, nil
}

func assertMapStringAny(val any) (map[string]any, error) {
	if m, ok := val.(map[string]any); ok {
		return m, nil
	}

	return nil, fmt.Errorf("expected map[string]any, got %T", val) //nolint: err113
}

func assertSliceMapStringAny(val any) ([]map[string]any, error) {
	if s, ok := val.([]any); ok {
		result := make([]map[string]any, 0, len(s))

		for _, item := range s {
			if m, ok := item.(map[string]any); ok {
				result = append(result, m)
			} else {
				return nil, fmt.Errorf("expected []map[string]any, got element of type %T", item) //nolint: err113
			}
		}

		return result, nil
	}

	return nil, fmt.Errorf("expected []any, got %T", val) //nolint: err113
}
