package lightfield

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// makeNextPageFunc advances the offset by the number of records returned.
// "A client can assume there are no more records to fetch when the number of returned records
// is less than the limit provided by the client."
// https://docs.lightfield.app/using-the-api/list-endpoints#pagination
func makeNextPageFunc(requestURL *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		records, err := jsonquery.New(node).ArrayRequired("data")
		if err != nil {
			return "", err
		}

		query := requestURL.Query()

		limit, err := strconv.Atoi(query.Get("limit"))
		if err != nil {
			return "", fmt.Errorf("invalid limit in read request: %w", err)
		}

		offset, err := strconv.Atoi(query.Get("offset"))
		if err != nil {
			return "", fmt.Errorf("invalid offset in read request: %w", err)
		}

		if len(records) < limit {
			return "", nil
		}

		nextURL, err := urlbuilder.FromRawURL(requestURL)
		if err != nil {
			return "", err
		}

		nextURL.WithQueryParam("offset", strconv.Itoa(offset+len(records)))

		return nextURL.String(), nil
	}
}

// makeUpdatedAtFilter keeps records whose last update falls within (Since, Until],
// matching ReadParams: Since is exclusive and Until inclusive.
// It serves /v1/members, which rejects filter query parameters (400 unrecognized_keys).
// Lists are sorted by createdAt, newest first, so every page must be read and paging can never
// stop early: an old record may have been updated recently.
// updatedAt is nullable; a record that was never updated is placed at its createdAt.
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
	query := jsonquery.New(record)

	value, err := query.StringOptional("updatedAt")
	if err != nil {
		return time.Time{}, err
	}

	if value == nil {
		value, err = query.StringOptional("createdAt")
		if err != nil {
			return time.Time{}, err
		}
	}

	if value == nil {
		return time.Time{}, fmt.Errorf("%w: updatedAt and createdAt", common.ErrMissingExpectedValues)
	}

	return time.Parse(time.RFC3339, *value)
}

// flattenRecord turns a Lightfield record into flat fields:
//
//	{"id": "acc_1", "fields": {"$name": {"value": "Acme", "valueType": "TEXT"}},
//	 "relationships": {"$owner": {"values": ["usr_1"], ...}}}
//
// becomes {"id": "acc_1", "$name": "Acme", "$owner": ["usr_1"]}.
// Field and relationship keys are kept verbatim, as the definitions endpoint names them.
func flattenRecord(node *ajson.Node) (map[string]any, error) {
	record, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	fields, _ := record["fields"].(map[string]any)
	relationships, _ := record["relationships"].(map[string]any)

	delete(record, "fields")
	delete(record, "relationships")

	for key, entry := range fields {
		if typed, ok := entry.(map[string]any); ok {
			setIfAbsent(record, key, typed["value"])
		}
	}

	for key, entry := range relationships {
		if typed, ok := entry.(map[string]any); ok {
			setIfAbsent(record, key, typed["values"])
		}
	}

	return record, nil
}

// setIfAbsent decides key clashes. System fields and relationships are `$`-prefixed, but custom
// ones are not (https://docs.lightfield.app/using-the-api/fields-and-relationships#field-keys),
// so a custom field may share a name with a record property (id, createdAt, updatedAt, httpLink,
// externalId) or with a custom relationship. The earlier value wins: record properties, then
// fields, then relationships. Record properties are kept because id and updatedAt drive record
// identity and incremental reads. The shadowed value stays available in Raw.
func setIfAbsent(record map[string]any, key string, value any) {
	if _, exists := record[key]; !exists {
		record[key] = value
	}
}
