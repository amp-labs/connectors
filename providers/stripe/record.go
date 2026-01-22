package stripe

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

var _ connectors.BatchRecordReaderConnector = &Connector{}

// GetRecordsByIds fetches full records from Stripe for a specific set of IDs.
// since Stripe doesn't support batch fetching, this method makes individual
// GET requests for each ID: /v1/{objectName}/{id}
//
//nolint:revive
func (c *Connector) GetRecordsByIds(
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	// Sanitize method arguments.
	config := common.ReadParams{
		ObjectName:        objectName,
		Fields:            datautils.NewSetFromList(fields),
		AssociatedObjects: associations,
	}

	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []common.ReadResultRow{}, nil
	}

	results := make([]common.ReadResultRow, 0, len(ids))

	for _, recordID := range ids {
		row, err := c.fetchSingleRecord(ctx, objectName, recordID, fields, associations)
		if err != nil {
			if errors.Is(err, common.ErrNotFound) {
				continue
			}

			return nil, err
		}

		results = append(results, *row)
	}

	return results, nil
}

// fetchSingleRecord fetches and processes a single record by ID.
func (c *Connector) fetchSingleRecord(
	ctx context.Context,
	objectName string,
	recordID string,
	fields []string,
	associations []string,
) (*common.ReadResultRow, error) {
	url, err := c.buildGetRecordURL(objectName, recordID, associations)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	body, hasBody := res.Body()
	if !hasBody {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	record, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse record %s: %w", recordID, err)
	}

	// extract ID from record (Stripe  returns id field)
	idStr, _ := record["id"].(string)
	if idStr == "" {
		// fallback to requested ID if record doesn't have id field
		idStr = recordID
	}

	filteredFields := common.ExtractLowercaseFieldsFromRaw(fields, record)

	row := &common.ReadResultRow{
		Id:     idStr,
		Fields: filteredFields,
		Raw:    record,
	}

	if len(associations) > 0 {
		extractedAssociations := extractAssociations(record, associations)
		if len(extractedAssociations) > 0 {
			row.Associations = extractedAssociations
		}
	}

	return row, nil
}

// buildGetRecordURL constructs a URL for fetching a single record by ID.
// Format: /v1/{objectName}/{id}
// Supports expand[] query parameters for associated objects.
func (c *Connector) buildGetRecordURL(objectName string, id string, associations []string) (*urlbuilder.URL, error) {
	url, err := c.getURL(objectName)
	if err != nil {
		return nil, err
	}

	url.AddPath(id)

	if len(associations) > 0 {
		url.WithQueryParamList("expand[]", associations)
	}

	return url, nil
}

// extractAssociations extracts expanded associations from a Stripe record.
// when Stripe expands objects, they are nested directly in the record as a single object.
func extractAssociations(record map[string]any, associationNames []string) map[string][]common.Association {
	associations := make(map[string][]common.Association)

	for _, assocName := range associationNames {
		assocValue, exists := record[assocName]
		if !exists || assocValue == nil {
			continue
		}

		// Handle expanded object
		assocMap, isMap := assocValue.(map[string]any)
		if !isMap {
			continue
		}

		idValue, exists := assocMap["id"]
		if !exists {
			continue
		}

		idStr, isString := idValue.(string)
		if !isString || idStr == "" {
			continue
		}

		rawCopy := make(map[string]any, len(assocMap))
		maps.Copy(rawCopy, assocMap)

		associations[assocName] = []common.Association{
			{
				ObjectId: idStr,
				Raw:      rawCopy,
			},
		}
	}

	return associations
}
