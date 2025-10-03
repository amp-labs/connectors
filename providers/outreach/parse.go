package outreach

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func getRecords(node *ajson.Node) ([]*ajson.Node, error) {
	return jsonquery.New(node).ArrayOptional(dataKey)
}

func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node, "links").StrWithDefault("next", "")
}

func getOutreachDataMarshaller(config common.ReadParams, included []dataItem,
	transformer common.RecordTransformer,
) common.MarshalFromNodeFunc {
	if len(config.AssociatedObjects) == 0 {
		return common.MakeMarshaledDataFunc(common.FlattenNestedFields(attributesKey))
	}

	return func(records []*ajson.Node, fields []string) ([]common.ReadResultRow, error) {
		result := make([]common.ReadResultRow, len(records))

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

			// Populate the result row with fields, raw data, and ID.
			result[idx] = common.ReadResultRow{
				Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
				Raw:    raw,
			}

			relationship, err := assertMapStringAny(record[relationshipsKey])
			if err != nil {
				return nil, err
			}

			// Fetch associations if requested in the config.
			if len(config.AssociatedObjects) > 0 {
				details, err := collectAssociationsTypesAndIDs(config.AssociatedObjects, relationship)
				if err != nil {
					return nil, err
				}

				associations, err := collectAssociations(included, details)
				if err != nil {
					return nil, err
				}

				result[idx].Associations = associations
			}
		}

		return result, nil
	}
}

// fetchAssociations fetches associated objects for a given record from the Outreach API.
// The API is queried with the `include` parameter to fetch related resources (e.g. owner, account).
func collectAssociationsTypesAndIDs(assoc []string, relationships map[string]any, // nolint: cyclop
) (map[string]map[string][]int, error) {
	// the first string is a requested object i.e `creator`
	// the second string a mapped object by outreach `user`
	// the last are the associated ids.
	detailsMap := make(map[string]map[string][]int)

	for _, assocObject := range assoc {
		relInfoAny, exists := relationships[assocObject]
		if !exists {
			continue
		}

		reldata, err := assertMapStringAny(relInfoAny)
		if err != nil {
			return nil, err
		}

		records, exists := reldata[dataKey]
		if !exists || records == nil {
			continue
		}

		switch rcds := records.(type) {
		// The relationships data response can be either:
		// - nil (no related objects)
		// - a single object (if only one related record exists)
		// - an array of objects (for multiple related records)
		case map[string]any:
			if err := processRecord(rcds, assocObject, detailsMap); err != nil {
				return nil, fmt.Errorf("single record: %w", err)
			}
		case []any:
			if len(rcds) == 0 {
				continue
			}

			allRecordsInfo, err := assertSliceMapStringAny(rcds)
			if err != nil {
				return nil, fmt.Errorf("records list: %w", err)
			}

			for _, rcd := range allRecordsInfo {
				if err := processRecord(rcd, assocObject, detailsMap); err != nil {
					return nil, fmt.Errorf("record in list: %w", err)
				}
			}
		}
	}

	return detailsMap, nil
}

func collectAssociations(included []dataItem, details map[string]map[string][]int,
) (map[string][]common.Association, error) {
	associations := make(map[string][]common.Association)

	for _, data := range included {
		for assoc, record := range details {
			if ids, exists := record[data.Type]; exists {
				if slices.Contains(ids, data.ID) {
					raw, err := data.ToMapStringAny()
					if err != nil {
						return nil, err
					}

					associations[assoc] = append(associations[assoc], common.Association{
						ObjectId:        strconv.Itoa(data.ID),
						AssociationType: data.Type,
						Raw:             raw,
					})
				}
			}
		}
	}

	return associations, nil
}

func processRecord(rcd map[string]any, associatedObject string, detailsMap map[string]map[string][]int) error {
	typeStr, ok := rcd["type"].(string)
	if !ok {
		return fmt.Errorf("data field missing or invalid 'type' for associated object %s", associatedObject) //nolint: err113
	}

	idF, ok := rcd["id"].(float64)
	if !ok {
		return fmt.Errorf("data field missing or invalid 'id' for associated object: %s", associatedObject) //nolint: err113
	}

	if detailsMap[associatedObject] == nil {
		detailsMap[associatedObject] = make(map[string][]int)
	}

	detailsMap[associatedObject][typeStr] = append(detailsMap[associatedObject][typeStr], int(idF))

	return nil
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
