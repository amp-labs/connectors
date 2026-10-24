package clarify

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

// makeNextPageFunc advances page[offset] past the records returned while links.next is set; links.next
// is null on the last page (https://developer.clarify.ai/docs/api-basics/pagination#response).
// The docs ask to fetch links.next verbatim, but the live API returns it with an http:// scheme
// (http://api.clarify.ai/v1/...), which does not answer, so the next page is built from the request
// URL instead. That also keeps the request's filters and page size.
func makeNextPageFunc(requestURL *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		next, err := jsonquery.New(node, "links").StrWithDefault("next", "")
		if err != nil || next == "" {
			return "", err
		}

		records, err := jsonquery.New(node).ArrayRequired("data")
		if err != nil {
			return "", err
		}

		offset, err := strconv.Atoi(requestURL.Query().Get(pageOffsetParam))
		if err != nil {
			return "", fmt.Errorf("invalid %s in read request: %w", pageOffsetParam, err)
		}

		nextURL, err := urlbuilder.FromRawURL(requestURL)
		if err != nil {
			return "", err
		}

		nextURL.WithQueryParam(pageOffsetParam, strconv.Itoa(offset+len(records)))
		nextURL.AddEncodingExceptions(spaceEncoding)

		return nextURL.String(), nil
	}
}

// makeUpdatedAtFilter keeps records whose `_updated_at` falls within [Since, Until]. Since is inclusive so
// a record stamped exactly at the previous read's Until is not skipped; re-reading it is harmless because
// records match by id.
// For records it trims the whole-second window the API filter returns (see updatedSinceParam).
// For users it is the only filter: GET /users documents only page and sortOrder parameters, so every
// page is read.
func makeUpdatedAtFilter(nextPage common.NextPageFunc) common.RecordsFilterFunc {
	boundary := readhelper.NewTimeBoundary()

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
			timestamp, ok := lastModified(record)

			// A record whose update time can't be read is kept, rather than failing the whole read or
			// silently dropping it.
			if !ok || boundary.Contains(params, timestamp) {
				filtered = append(filtered, record)
			}
		}

		return filtered, next, nil
	}
}

// lastModified reads attributes._updated_at, an RFC 3339 string with milliseconds in every live response.
func lastModified(record *ajson.Node) (time.Time, bool) {
	value, err := jsonquery.New(record, "attributes").StringOptional("_updated_at")
	if err != nil || value == nil {
		return time.Time{}, false
	}

	timestamp, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return time.Time{}, false
	}

	return timestamp, true
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
