package clarify

import (
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// getNextRecordsURL returns links.next, which is null on the last page.
// https://developer.clarify.ai/docs/api-basics/pagination#response
func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node, "links").StrWithDefault("next", "")
}

// makeUpdatedAtFilter keeps records whose `_updated_at` falls within (Since, Until],
// matching ReadParams: Since is exclusive and Until inclusive.
// It serves GET /users, which documents only page and sortOrder parameters, so every page is read.
func makeUpdatedAtFilter(nextPage common.NextPageFunc) common.RecordsFilterFunc {
	boundary := readhelper.NewTimeBoundary(readhelper.ExcludeSince())

	return func(params common.ReadParams, body *ajson.Node, records []*ajson.Node) ([]*ajson.Node, string, error) {
		next, err := nextPage(body)
		if err != nil {
			return nil, "", err
		}

		if params.Since.IsZero() && params.Until.IsZero() {
			return records, next, nil
		}

		filtered := make([]*ajson.Node, 0, len(records))

		for _, record := range records {
			timestamp, err := lastModified(record)
			if err != nil {
				return nil, "", err
			}

			if boundary.Contains(params, timestamp) {
				filtered = append(filtered, record)
			}
		}

		return filtered, next, nil
	}
}

func lastModified(record *ajson.Node) (time.Time, error) {
	value, err := jsonquery.New(record, "attributes").StringOptional("_updated_at")
	if err != nil {
		return time.Time{}, err
	}

	if value == nil {
		return time.Time{}, fmt.Errorf("%w: attributes._updated_at", common.ErrMissingExpectedValues)
	}

	return time.Parse(time.RFC3339, *value)
}

// flattenRecord promotes a JSON:API resource's attributes to top-level fields:
//
//	{"id": "5f8b...", "type": "person", "attributes": {"job_title": "VP", "_updated_at": "..."}}
//
// becomes {"id": "5f8b...", "job_title": "VP", "_updated_at": "..."}.
// Attribute names are kept verbatim, as the schemas endpoint names them. The resource `id` wins over
// an attribute of the same name, though attributes carry the id as `_id`.
// https://developer.clarify.ai/docs/getting-started/key-concepts#the-jsonapi-envelope
func flattenRecord(node *ajson.Node) (map[string]any, error) {
	attributes, err := jsonquery.New(node).ObjectOptional("attributes")
	if err != nil {
		return nil, err
	}

	record := make(map[string]any)

	if attributes != nil {
		record, err = jsonquery.Convertor.ObjectToMap(attributes)
		if err != nil {
			return nil, err
		}
	}

	id, err := jsonquery.New(node).StringRequired("id")
	if err != nil {
		return nil, err
	}

	record["id"] = id

	return record, nil
}
