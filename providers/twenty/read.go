package twenty

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	// List endpoints take `limit` up to 200 (default 60) and page with `starting_after`, a cursor taken from
	// pageInfo.endCursor: the OpenAPI document each instance serves at /rest/open-api/core (v2.50.0).
	maxPageSize = 200

	// Twenty compares date-time filters at millisecond precision, both bounds inclusive (checked live on
	// v2.50.0 against a value stored as 12:40:11.79276: [gte] and [lte] .792Z both matched it, .793Z [gte]
	// and .791Z [lte] did not), and rejects six fractional digits ("Expected format:
	// 'YYYY-MM-DDTHH:mm:ssZ'"). Records show updatedAt with milliseconds, so a bound truncated to the
	// millisecond keeps every record whose shown updatedAt is inside [Since, Until].
	filterTimeLayout = "2006-01-02T15:04:05.000Z"

	// cutRecordKeys is the number of keys from which a record read without `fields` may have been cut by
	// Twenty's default set of 200 fields (see mayBeCut).
	cutRecordKeys = 150

	// fieldsCursorPrefix marks the NextPage of a read that sends `fields` (see Read). Twenty's own cursors are
	// base64, which has no colon.
	fieldsCursorPrefix = "fields:"
)

var (
	errFieldNotReturned = errors.New("twenty's response omitted a requested field that the object's schema has")
	// errCaseCollision: two fields of one object whose names differ only by case (Twenty allows renewalNotes
	// next to renewalnotes, seen live on v2.50.0) cannot be told apart through ReadResult.Fields, whose keys
	// are lowercase.
	errCaseCollision = errors.New("field name matches several Twenty fields when case is ignored")
)

// Read reads each page without the `fields` parameter, so Raw is the record as Twenty holds it. On an object
// with more than 200 readable fields Twenty then returns a default set of 200, which a requested field can fall
// outside; its docs say to pass `fields` explicitly for any field beyond that set
// (https://docs.twenty.com/developers/extend/api). So a page missing a field the schema has is read again with
// `fields` naming the requested fields, and so is every later page of the read, the mode being kept in NextPage.
// Each such page costs one metadata request on top of the record request, and its Raw holds the requested
// fields only.
func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	result, err := c.Reader.Read(ctx, params)
	if !errors.Is(err, errFieldNotReturned) {
		return result, err
	}

	if _, explicit := splitNextPage(params.NextPage); explicit {
		return nil, err
	}

	params.NextPage = common.NextPageToken(fieldsCursorPrefix + params.NextPage.String())

	return c.Reader.Read(ctx, params)
}

// splitNextPage returns Twenty's cursor and whether the read sends `fields`.
func splitNextPage(token common.NextPageToken) (string, bool) {
	return strings.CutPrefix(token.String(), fieldsCursorPrefix)
}

// buildReadRequest lists records with depth=0, so a record holds its own scalar and composite fields and the
// join column of each many-to-one relation, and no related records.
//
// Unless Read has switched to `fields`, no `fields` parameter is sent: Twenty then returns every field of the
// record, except on objects with more than 200 readable fields (see Read and checkRequestedFieldsReturned).
// Raw stays the provider's record, and the requested fields are picked from it.
//
// With ReadParams.Deleted, the read lists the trash instead (deletedAt[is]:NOT_NULL): Twenty lists only records
// that are not soft-deleted unless the filter names deletedAt. A soft delete moves updatedAt to the deletion
// time (seen live on v2.50.0), so Since and Until select deletions by when they happened.
//
// Since and Until both go to Twenty as updatedAt[gte] and updatedAt[lte]. Paging follows a keyset cursor
// (pageInfo.endCursor encodes the last record's id), not an offset, so a record edited out of the window
// during the read cannot shift the next page and skip another record.
// Nothing is filtered in code: every record Twenty returns is delivered, and Done follows hasNextPage.
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	readURL, err := c.recordsURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	readURL.WithQueryParam("depth", "0")
	readURL.WithQueryParam("limit",
		strconv.Itoa(min(readhelper.PageSizeWithDefault(params, maxPageSize), maxPageSize)))

	cursor, explicit := splitNextPage(params.NextPage)
	if cursor != "" {
		readURL.WithQueryParam("starting_after", cursor)
	}

	if explicit {
		fields, err := c.fieldsParameter(ctx, params)
		if err != nil {
			return nil, err
		}

		readURL.WithQueryParam("fields", fields)
	}

	if filter := recordsFilter(params); filter != "" {
		readURL.WithQueryParam("filter", filter)
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, readURL.String(), nil)
}

// recordsURL builds {BaseURL}/rest/{objectName}{/recordId}.
func (c *Connector) recordsURL(objectName string, recordID ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, append([]string{"rest", objectName}, recordID...)...)
}

// fieldsParameter names the requested fields as `fields` takes them: Twenty's field names in their exact case
// (renewalnotes for renewalNotes is a 400), a many-to-one relation by the relation's name (vendor for vendorId,
// which Twenty then returns as vendorId), seen live on v2.50.0. id is always selected: each row's id is read
// from it. A requested field the schema does not know is left out, since Twenty refuses the whole request for
// one ("Unknown or unreadable fields"); it is absent from the rows, as it is without `fields`.
func (c *Connector) fieldsParameter(ctx context.Context, params common.ReadParams) (string, error) {
	objects, err := c.fetchObjects(ctx)
	if err != nil {
		return "", err
	}

	index := objects[params.ObjectName].keyIndex()
	names := []string{"id"}

	for _, field := range params.Fields.List() {
		keys := index.byLowercase[lowercase(field)]

		switch len(keys) {
		case 0:
		case 1:
			names = append(names, index.exact[keys[0]].Name)
		default:
			return "", fmt.Errorf("%w: %w: %s matches %v (object %s)",
				common.ErrCaller, errCaseCollision, field, keys, params.ObjectName)
		}
	}

	slices.Sort(names)

	return strings.Join(slices.Compact(names), ","), nil
}

// recordsFilter joins conditions with commas, which Twenty reads as AND.
func recordsFilter(params common.ReadParams) string {
	conditions := make([]string, 0, 3) // nolint:mnd

	if params.Deleted {
		conditions = append(conditions, "deletedAt[is]:NOT_NULL")
	}

	if !params.Since.IsZero() {
		conditions = append(conditions, fmt.Sprintf(`updatedAt[gte]:"%s"`, formatFilterTime(params.Since)))
	}

	if !params.Until.IsZero() {
		conditions = append(conditions, fmt.Sprintf(`updatedAt[lte]:"%s"`, formatFilterTime(params.Until)))
	}

	return strings.Join(conditions, ",")
}

func formatFilterTime(t time.Time) string {
	return t.UTC().Truncate(time.Millisecond).Format(filterTimeLayout)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResultFiltered(params, response,
		// Records are under data.<namePlural>.
		func(node *ajson.Node) ([]*ajson.Node, error) {
			return jsonquery.New(node, "data").ArrayRequired(params.ObjectName)
		},
		func(params common.ReadParams, body *ajson.Node, records []*ajson.Node) ([]*ajson.Node, string, error) {
			if err := checkCaseCollisions(params, records); err != nil {
				return nil, "", err
			}

			if err := c.checkRequestedFieldsReturned(ctx, params, records); err != nil {
				return nil, "", err
			}

			cursor, explicit := splitNextPage(params.NextPage)

			next, err := nextPageCursor(cursor, body)
			if next != "" && explicit {
				next = fieldsCursorPrefix + next
			}

			return records, next, err
		},
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
}

// nextPageCursor is pageInfo.endCursor while pageInfo.hasNextPage is true. A short or empty page is not an
// end: only hasNextPage says so, and a response without it, or a cursor that does not move, fails the read.
func nextPageCursor(current string, body *ajson.Node) (string, error) {
	hasNextPage, err := jsonquery.New(body, "pageInfo").BoolRequired("hasNextPage")
	if err != nil || !hasNextPage {
		return "", err
	}

	cursor, err := jsonquery.New(body, "pageInfo").StrWithDefault("endCursor", "")
	if err != nil {
		return "", err
	}

	if cursor == "" || cursor == current {
		return "", fmt.Errorf("%w: hasNextPage is true without a new endCursor (%q)", common.ErrNextPageInvalid, cursor)
	}

	return cursor, nil
}

// checkCaseCollisions refuses a requested field that names two keys of a record when case is ignored:
// ReadResult.Fields would hold one of them under the lowercase name, and a write of it could reach the other.
func checkCaseCollisions(params common.ReadParams, records []*ajson.Node) error {
	if len(records) == 0 {
		return nil
	}

	keys := make(map[string][]string)
	for _, key := range records[0].Keys() {
		keys[lowercase(key)] = append(keys[lowercase(key)], key)
	}

	for _, field := range params.Fields.List() {
		if matches := keys[lowercase(field)]; len(matches) > 1 {
			return fmt.Errorf("%w: %w: %s matches %v (object %s)",
				common.ErrCaller, errCaseCollision, field, matches, params.ObjectName)
		}
	}

	return nil
}

// checkRequestedFieldsReturned refuses a page without a requested field that the object's schema has. Twenty
// returns every selected field, null ones included, but without a `fields` parameter it selects at most 200
// fields on a wide object, and the selection can be narrower still (a morph relation counts once among the
// 200 and is then left out, seen live: 199 keys). So any missing requested field is checked against the
// schema, which is fetched only then; a field the schema does not know is simply absent. Without `fields`,
// Read then reads the page again with it; with `fields`, a schema field still missing fails the read, as
// Twenty does not leave out a field it was asked for, and no retry would bring it.
//
// A page that Twenty may have cut is also checked against the schema when nothing is missing: the default
// set is taken in name order, so of renewalNotes and renewalnotes it can keep the first and cut the second
// (seen live on v2.50.0), and a request for renewalnotes would otherwise be answered with renewalNotes's
// value. A requested field that names several schema fields when case is ignored is refused, as it is when
// both are in the record (checkCaseCollisions).
func (c *Connector) checkRequestedFieldsReturned(ctx context.Context, params common.ReadParams,
	records []*ajson.Node,
) error {
	_, explicit := splitNextPage(params.NextPage)

	missing := fieldsMissingFromRecords(params, records)
	if len(missing) == 0 && (explicit || !mayBeCut(records)) {
		return nil
	}

	objects, err := c.fetchObjects(ctx)
	if err != nil {
		return err
	}

	object, ok := objects[params.ObjectName]
	if !ok {
		return nil
	}

	index := object.keyIndex()

	for _, field := range params.Fields.List() {
		if keys := index.byLowercase[lowercase(field)]; len(keys) > 1 {
			return fmt.Errorf("%w: %w: %s matches %v (object %s)",
				common.ErrCaller, errCaseCollision, field, keys, params.ObjectName)
		}
	}

	for _, field := range missing {
		if object.hasRecordKey(field) {
			return fmt.Errorf("%w: %w: %s (object %s)",
				common.ErrServerNonRetryable, errFieldNotReturned, field, params.ObjectName)
		}
	}

	return nil
}

// mayBeCut reports a record that Twenty's default set of 200 fields may have cut. Such a record carries close
// to 200 keys (199 seen live, a morph relation taking a place without a key of its own); one with fewer than
// cutRecordKeys was not cut, as its object would need dozens of morph relations among those 200 places.
func mayBeCut(records []*ajson.Node) bool {
	for _, record := range records {
		if len(record.Keys()) >= cutRecordKeys {
			return true
		}
	}

	return false
}

// fieldsMissingFromRecords lists the requested fields absent from the first record that lacks any.
func fieldsMissingFromRecords(params common.ReadParams, records []*ajson.Node) []string {
	for _, record := range records {
		keys := record.Keys()

		present := make(map[string]bool, len(keys))
		for _, key := range keys {
			present[lowercase(key)] = true
		}

		var missing []string

		for _, field := range params.Fields.List() {
			if !present[lowercase(field)] {
				missing = append(missing, field)
			}
		}

		if len(missing) != 0 {
			return missing
		}
	}

	return nil
}
