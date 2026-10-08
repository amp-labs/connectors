package mailgun

import (
	neturl "net/url"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// optionalNodeRecords extracts the records array as ajson nodes, tolerating a
// missing or null array ("items": null on empty/last pages).
func optionalNodeRecords(recordsKey string) func(*ajson.Node) ([]*ajson.Node, error) {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayOptional(recordsKey)
	}
}

// timeFilter sieves a page's records by the object's timestamp field against
// ReadParams.Since/Until. The next-page token is computed from the raw page:
// an empty raw page ends pagination (Mailgun's cursor endpoints echo
// paging.next even on the last page), while a page that merely filters down to
// zero records keeps paginating.
//
// A dedicated filter (rather than readhelper.MakeTimeFilterFunc) is used because
// Mailgun mixes timestamp layouts across objects (RFC 1123 with UTC/GMT zones,
// RFC 1123Z, zone-less ISO) and some records carry an empty timestamp (e.g. the
// keys endpoint's public key). MakeTimeFilterFunc takes one layout and errors on
// an unparsable or empty value; this keeps such records instead, so a read never
// fails or silently drops data on a format surprise.
func (c *Connector) timeFilter(
	spec objectSpec, readParams common.ReadParams, reqURL *neturl.URL, recordsKey string,
) func(common.ReadParams, *ajson.Node, []*ajson.Node) ([]*ajson.Node, string, error) {
	return func(params common.ReadParams, body *ajson.Node, records []*ajson.Node) ([]*ajson.Node, string, error) {
		nextPage := ""

		if len(records) > 0 {
			var err error

			nextPage, err = c.nextPageFunc(spec, readParams, reqURL, recordsKey)(body)
			if err != nil {
				return nil, "", err
			}
		}

		filtered := make([]*ajson.Node, 0, len(records))

		for _, record := range records {
			if recordWithinWindow(record, spec.timeField, params.Since, params.Until) {
				filtered = append(filtered, record)
			}
		}

		return filtered, nextPage, nil
	}
}

// recordTimeLayouts are the timestamp formats Mailgun uses across record date
// fields, tried in order. Verified live: most date fields are RFC 1123 with a
// zone abbreviation ("Fri, 18 Sep 2026 12:48:51 UTC" on bounces/unsubscribes/
// templates, "Wed, 15 May 2024 02:50:52 GMT" on domains); analytics/tags
// last_seen carries a numeric zone ("... +0000", RFC 1123Z); the keys endpoint
// uses a zone-less ISO 8601 form ("2024-06-21T03:17:35", UTC). RFC 3339 is kept
// as a fallback for any zoned ISO variant.
//
//nolint:gochecknoglobals
var recordTimeLayouts = []string{time.RFC1123Z, time.RFC1123, time.RFC3339, isoUTCNoZoneLayout}

// isoUTCNoZoneLayout is the keys endpoint's zone-less ISO 8601 form; values are UTC.
const isoUTCNoZoneLayout = "2006-01-02T15:04:05"

// recordWithinWindow reports whether the record's timestamp field falls inside
// [since, until]. Records with a missing or unparsable timestamp are kept —
// filtering must never silently drop data on format surprises.
func recordWithinWindow(record *ajson.Node, field string, since, until time.Time) bool {
	value, err := jsonquery.New(record).StringOptional(field)
	if err != nil || value == nil {
		// Non-string or absent timestamp: keep the record.
		return true
	}

	timestamp, ok := parseRecordTime(*value)
	if !ok {
		return true
	}

	if !since.IsZero() && timestamp.Before(since) {
		return false
	}

	if !until.IsZero() && timestamp.After(until) {
		return false
	}

	return true
}

func parseRecordTime(value string) (time.Time, bool) {
	for _, layout := range recordTimeLayouts {
		if timestamp, err := time.ParseInLocation(layout, value, time.UTC); err == nil {
			return timestamp, true
		}
	}

	return time.Time{}, false
}
