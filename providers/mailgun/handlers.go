package mailgun

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/amp-labs/connectors/providers/mailgun/metadata"
	"github.com/spyzhov/ajson"
)

// errMissingDomain is returned when a domain-scoped object is read without the
// connector having a workspace (Mailgun sending domain) configured.
var errMissingDomain = errors.New("mailgun: a domain (workspace) is required to read this object")

// errListMemberPagesExceeded guards the per-list member pagination loop.
var errListMemberPagesExceeded = errors.New("mailgun: list members exceeded page cap")

const (
	// Mailgun list endpoints cap page size differently (100 for most, 1000 for
	// the domain event/suppression endpoints). 100 is accepted everywhere and is
	// the documented default. We use it as both default and cap so the offset
	// "records < limit ⇒ last page" check can never be fooled by a silent
	// server-side cap below what we requested.
	defaultPageSize = 100
	maxPageSize     = 100

	limitParam = "limit"
	skipParam  = "skip"

	// Concurrency + safety caps for the lists/members fan-out.
	maxConcurrentListFetch = 4
	maxListMemberPages     = 200
)

// domainPlaceholders are the path parameters that all resolve to the connector's
// workspace (the Mailgun sending domain). The Mailgun spec spells the domain
// under several names; schemas.json keeps them verbatim and we substitute here.
//
//nolint:gochecknoglobals
var domainPlaceholders = []string{"{domain_name}", "{authority_name}"}

// listAddressPlaceholder is the parent identifier in the lists/members path. It
// is NOT the workspace domain; lists/members is read by fanning out over the
// account's mailing lists (see parseListMembersResponse).
const listAddressPlaceholder = "{list_address}"

type paginationStyle int

const (
	// paginationCursor follows the response's paging.next URL. Mailgun always
	// returns paging.next (even on the last page); the follow-up request returns
	// an empty items page, which common.ParseResult treats as Done.
	paginationCursor paginationStyle = iota
	// paginationOffset advances skip by limit until a short page is returned.
	paginationOffset
	// paginationNone means the endpoint returns the full collection in one call.
	paginationNone
	// paginationLogsToken is the POST /v1/analytics/logs cursor: the next token
	// lives in pagination.next of the JSON body and is echoed back in the request
	// body of the following page.
	paginationLogsToken
	// paginationBodyOffset is the POST /v1/analytics/tags style: skip and limit
	// travel in the request body's pagination object; a short page ends the read.
	// NextPage carries the next skip value.
	paginationBodyOffset
)

type objectScope int

const (
	scopeAccount objectScope = iota
	scopeDomain              // substitute domain placeholders with the workspace
	scopeList                // nested fan-out over mailing lists (lists/members)
)

type objectSpec struct {
	pagination paginationStyle
	scope      objectScope
	// timeField is the record timestamp used for connector-side Since/Until
	// filtering; empty means the object has no usable record timestamp (or, for
	// analytics/logs, that filtering happens provider-side).
	timeField string
}

// objectReadSpecs is the verified per-object read behaviour, sourced from the
// Mailgun OpenAPI spec (providers/mailgun/metadata/openapi/mailgun.yaml). The
// records-array key is not stored here — it comes from the static schema
// (responseKey) via metadata.Schemas.
//
// Incremental read: Mailgun's list endpoints do not accept an "updated since"
// filter, so for objects whose records carry a timestamp we sieve each page
// connector-side against Since/Until (see timeFilter). Objects with no record
// timestamp ignore Since/Until; only analytics/logs supports provider-side
// time scoping (see buildLogsRequest).
//
// API reference: https://documentation.mailgun.com/docs/mailgun/api-reference/intro/
//
//nolint:gochecknoglobals
var objectReadSpecs = map[string]objectSpec{
	// Domain-scoped, cursor paginated.
	"bounces":      {paginationCursor, scopeDomain, "created_at"},
	"complaints":   {paginationCursor, scopeDomain, "created_at"},
	"unsubscribes": {paginationCursor, scopeDomain, "created_at"},
	"whitelists":   {paginationCursor, scopeDomain, "createdAt"},
	"templates":    {paginationCursor, scopeDomain, "createdAt"},

	// Domain-scoped, offset paginated / single-shot.
	"domains/credentials": {paginationOffset, scopeDomain, "created_at"},
	"domains/keys":        {paginationNone, scopeDomain, ""},

	// Account-scoped, cursor paginated.
	"forwards":              {paginationCursor, scopeAccount, "updated_at"},
	"dkim/keys":             {paginationCursor, scopeAccount, ""},
	"account/templates":     {paginationCursor, scopeAccount, "createdAt"},
	"dynamic_pools/domains": {paginationCursor, scopeAccount, ""}, // response carries paging.Next

	// Account-scoped, offset paginated.
	"domains":              {paginationOffset, scopeAccount, "created_at"},
	"routes":               {paginationOffset, scopeAccount, "created_at"},
	"lists":                {paginationOffset, scopeAccount, "created_at"},
	"accounts/subaccounts": {paginationOffset, scopeAccount, "updated_at"},
	"users":                {paginationOffset, scopeAccount, ""},

	// Account-scoped, single-shot (no pagination parameters).
	"ip_pools":                          {paginationNone, scopeAccount, ""},
	"ips":                               {paginationNone, scopeAccount, ""},
	"keys":                              {paginationNone, scopeAccount, "updated_at"},
	"webhooks":                          {paginationNone, scopeAccount, "created_at"},
	"accounts/subaccounts/ip_pools/all": {paginationNone, scopeAccount, ""},
	"ip_whitelist":                      {paginationNone, scopeAccount, ""},
	"thresholds/limits":                 {paginationNone, scopeAccount, "updated_at"},
	"thresholds/alerts/send":            {paginationNone, scopeAccount, "updated_at"},

	// Nested: fan-out over mailing lists. Member records carry no timestamps.
	"lists/members": {paginationOffset, scopeList, ""},

	// POST-sourced logs with body-token pagination; time scoping is
	// provider-side via the request body's start/end window.
	"analytics/logs": {paginationLogsToken, scopeAccount, ""},

	// POST-sourced account tags (the modern Tags API, replacing the deprecated
	// GET /v3/{domain}/tags) with body-offset pagination. last_seen filters
	// connector-side when the API returns it as a timestamp string.
	"analytics/tags": {paginationBodyOffset, scopeAccount, "last_seen"},
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	spec, ok := objectReadSpecs[params.ObjectName]
	if !ok {
		return nil, common.ErrOperationNotSupportedForObject
	}

	// The analytics objects are POSTs with a JSON body carrying the pagination
	// state (a cursor token for logs, skip/limit for tags).
	switch spec.pagination { //nolint:exhaustive // the remaining styles are GET-based
	case paginationLogsToken:
		return c.buildLogsRequest(ctx, params)
	case paginationBodyOffset:
		return c.buildTagsRequest(ctx, params)
	}

	// Subsequent pages are fetched from the exact URL we handed back (offset and
	// cursor styles both encode their state in that URL). lists/members reuses
	// this: its NextPage is the parent mailing-lists list URL.
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	url, err := c.buildInitialURL(params, spec)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

// buildInitialURL resolves the first-page URL for GET-based objects. For
// lists/members it returns the parent mailing-lists list URL; the fan-out to
// members happens in parseReadResponse.
func (c *Connector) buildInitialURL(params common.ReadParams, spec objectSpec) (*urlbuilder.URL, error) {
	objectName := params.ObjectName
	if spec.scope == scopeList {
		objectName = "lists"
	}

	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), objectName)
	if err != nil {
		return nil, err
	}

	if spec.scope == scopeDomain {
		path, err = c.substituteDomain(path)
		if err != nil {
			return nil, err
		}
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	applyPagination(url, spec.pagination, params)

	return url, nil
}

// applyPagination sets the first-page query parameters for the given style.
func applyPagination(url *urlbuilder.URL, style paginationStyle, params common.ReadParams) {
	switch style {
	case paginationCursor:
		url.WithQueryParam(limitParam, pageSize(params))
	case paginationOffset:
		url.WithQueryParam(limitParam, pageSize(params))
		url.WithQueryParam(skipParam, "0")
	case paginationNone, paginationLogsToken, paginationBodyOffset:
		// No query-parameter pagination.
	}
}

// substituteDomain replaces every domain placeholder in path with the workspace.
func (c *Connector) substituteDomain(path string) (string, error) {
	if c.workspace == "" {
		return "", errMissingDomain
	}

	for _, placeholder := range domainPlaceholders {
		path = strings.ReplaceAll(path, placeholder, c.workspace)
	}

	return path, nil
}

func pageSize(params common.ReadParams) string {
	if params.PageSize <= 0 || params.PageSize > maxPageSize {
		return strconv.Itoa(defaultPageSize)
	}

	return strconv.Itoa(params.PageSize)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	spec := objectReadSpecs[params.ObjectName]

	if spec.scope == scopeList {
		return c.parseListMembersResponse(ctx, params, request, resp)
	}

	recordsKey := c.arrayFieldName(params.ObjectName)

	// Connector-side incremental read: when the caller passes Since/Until and
	// the object's records carry a timestamp, sieve each page by that field
	// (the endpoints themselves accept no time parameters). Pagination still
	// walks every page — the next-page token is derived from the unfiltered
	// page, so a page that filters down to zero records does not end the read.
	if spec.timeField != "" && (!params.Since.IsZero() || !params.Until.IsZero()) {
		return common.ParseResultFiltered(
			params,
			resp,
			optionalNodeRecords(recordsKey),
			c.timeFilter(spec, params, request.URL, recordsKey),
			common.MakeMarshaledDataFunc(nil),
			params.Fields,
		)
	}

	return common.ParseResult(
		resp,
		// Mailgun returns "items": null (not []) for an empty/last page, so use
		// the optional extractor — it yields an empty slice instead of erroring,
		// and ParseResult then marks the page Done.
		common.ExtractOptionalRecordsFromPath(recordsKey),
		c.nextPageFunc(spec, params, request.URL, recordsKey),
		common.GetMarshaledData,
		params.Fields,
	)
}

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

// recordTimeLayouts are the timestamp formats observed in Mailgun record
// fields, e.g. "Wed, 15 May 2024 02:50:53 +0000" (RFC 1123 with numeric zone),
// "Thu, 11 Dec 2025 01:49:40 UTC" (RFC 1123 with zone abbreviation).
//
//nolint:gochecknoglobals
var recordTimeLayouts = []string{time.RFC1123Z, time.RFC1123, time.RFC3339}

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
		if timestamp, err := time.Parse(layout, value); err == nil {
			return timestamp, true
		}
	}

	return time.Time{}, false
}

func (c *Connector) arrayFieldName(objectName string) string {
	return metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), objectName)
}

// nextPageFunc returns the pagination extractor for the object's style.
func (c *Connector) nextPageFunc(
	spec objectSpec, params common.ReadParams, reqURL *neturl.URL, recordsKey string,
) func(*ajson.Node) (string, error) {
	switch spec.pagination {
	case paginationCursor:
		return c.cursorNextPage
	case paginationOffset:
		return c.offsetNextPage(reqURL, recordsKey)
	case paginationLogsToken:
		return logsNextPage
	case paginationBodyOffset:
		return bodyOffsetNextPage(params, recordsKey)
	case paginationNone:
		fallthrough
	default:
		return func(*ajson.Node) (string, error) { return "", nil }
	}
}

// bodyOffsetNextPage advances the body-carried skip (analytics/tags) by the
// returned record count; a short page ends the read. NextPage is the next skip
// value, which buildTagsRequest echoes into the following request body.
func bodyOffsetNextPage(params common.ReadParams, recordsKey string) func(*ajson.Node) (string, error) {
	return func(body *ajson.Node) (string, error) {
		records, err := common.ExtractOptionalRecordsFromPath(recordsKey)(body)
		if err != nil {
			return "", err
		}

		limit := mustPageSize(params)
		if len(records) < limit {
			return "", nil
		}

		return strconv.Itoa(nextPageInt(params.NextPage) + len(records)), nil
	}
}

// nextPageInt interprets NextPage as a skip offset; empty or invalid means 0.
func nextPageInt(token common.NextPageToken) int {
	value, err := strconv.Atoi(token.String())
	if err != nil || value < 0 {
		return 0
	}

	return value
}

// cursorNextPage reads paging.next and resolves it against the base URL when the
// provider returns a relative path (some v1 endpoints do). An empty items page
// is what actually stops pagination — handled by common.ParseResult — so we can
// unconditionally hand back paging.next here.
func (c *Connector) cursorNextPage(body *ajson.Node) (string, error) {
	next, err := pagingString(body, "paging", "next")
	if err != nil {
		return "", err
	}

	if next == "" {
		// Some v1 endpoints spell the paging keys capitalized — the OpenAPI spec
		// declares GET /v1/dynamic_pools/domains with paging {Next, Previous,
		// First, Last} — so fall back before concluding there is no next page.
		next, err = pagingString(body, "paging", "Next")
		if err != nil || next == "" {
			return "", err
		}
	}

	return c.absoluteURL(next), nil
}

// logsNextPage reads the raw pagination.next token from the logs response.
func logsNextPage(body *ajson.Node) (string, error) {
	return pagingString(body, "pagination", "next")
}

// pagingString reads container.key as a string, returning "" when either the
// container or the key is absent.
func pagingString(body *ajson.Node, container, key string) (string, error) {
	node, err := jsonquery.New(body).ObjectOptional(container)
	if err != nil || node == nil {
		return "", err
	}

	value, err := jsonquery.New(node).StringOptional(key)
	if err != nil || value == nil {
		return "", err
	}

	return *value, nil
}

// offsetNextPage advances skip by the returned record count until a short page
// signals the end.
func (c *Connector) offsetNextPage(reqURL *neturl.URL, recordsKey string) func(*ajson.Node) (string, error) {
	return func(body *ajson.Node) (string, error) {
		if reqURL == nil {
			return "", nil
		}

		records, err := common.ExtractOptionalRecordsFromPath(recordsKey)(body)
		if err != nil {
			return "", err
		}

		limit := queryParamInt(reqURL, limitParam, defaultPageSize)
		if len(records) < limit {
			return "", nil
		}

		next, err := urlbuilder.New(reqURL.String())
		if err != nil {
			return "", err
		}

		skip := queryParamInt(reqURL, skipParam, 0)
		next.WithQueryParam(skipParam, strconv.Itoa(skip+len(records)))

		return next.String(), nil
	}
}

// absoluteURL prefixes relative next-page paths with the connector base URL.
func (c *Connector) absoluteURL(next string) string {
	if strings.HasPrefix(next, "http://") || strings.HasPrefix(next, "https://") {
		return next
	}

	base := strings.TrimRight(c.ProviderInfo().BaseURL, "/")

	if !strings.HasPrefix(next, "/") {
		next = "/" + next
	}

	return base + next
}

func queryParamInt(u *neturl.URL, key string, defaultValue int) int {
	v, err := strconv.Atoi(u.Query().Get(key))
	if err != nil || v < 0 {
		return defaultValue
	}

	return v
}

// noNextPage is a pagination extractor for single-page fetches.
func noNextPage(*ajson.Node) (string, error) { return "", nil }

// listsArrayKey is the records key of the /v3/lists response.
const listsArrayKey = "items"

// parseListMembersResponse implements the lists/members read. The response in
// hand is a page of the account's mailing lists; for each list we fetch all of
// its members and flatten them. Pagination advances through the mailing-lists
// list itself (offset), so one Read call returns the members of one page of
// lists and Done fires when the lists list is exhausted.
func (c *Connector) parseListMembersResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	body, ok := resp.Body()
	if !ok {
		return &common.ReadResult{Done: true, Data: []common.ReadResultRow{}}, nil
	}

	listRecords, err := common.ExtractOptionalRecordsFromPath(listsArrayKey)(body)
	if err != nil {
		return nil, err
	}

	rows, err := c.fetchMembersForLists(ctx, listAddresses(listRecords), params)
	if err != nil {
		return nil, err
	}

	nextPage, err := c.offsetNextPage(request.URL, listsArrayKey)(body)
	if err != nil {
		return nil, err
	}

	return &common.ReadResult{
		Rows:     int64(len(rows)),
		Data:     rows,
		NextPage: common.NextPageToken(nextPage),
		Done:     nextPage == "",
	}, nil
}

// listAddresses pulls the "address" of each mailing list, which serves as the
// list identifier substituted into the members path.
func listAddresses(records []map[string]any) []string {
	addresses := make([]string, 0, len(records))

	for _, record := range records {
		if address, ok := record["address"].(string); ok && address != "" {
			addresses = append(addresses, address)
		}
	}

	return addresses
}

// fetchMembersForLists fans out one member fetch per mailing list concurrently
// and preserves list order in the flattened output.
func (c *Connector) fetchMembersForLists(
	ctx context.Context, addresses []string, params common.ReadParams,
) ([]common.ReadResultRow, error) {
	if len(addresses) == 0 {
		return []common.ReadResultRow{}, nil
	}

	perList := make([][]common.ReadResultRow, len(addresses))
	jobs := make([]simultaneously.Job, len(addresses))

	for i, listAddress := range addresses {
		idx, address := i, listAddress

		jobs[idx] = func(ctx context.Context) error {
			rows, fetchErr := c.fetchAllListMembers(ctx, address, params)
			if fetchErr != nil {
				return fmt.Errorf("fetching members for list %s: %w", address, fetchErr)
			}

			perList[idx] = rows

			return nil
		}
	}

	if err := simultaneously.DoCtx(ctx, maxConcurrentListFetch, jobs...); err != nil {
		return nil, err
	}

	var data []common.ReadResultRow
	for _, rows := range perList {
		data = append(data, rows...)
	}

	return data, nil
}

// fetchAllListMembers walks every offset page of one list's members.
func (c *Connector) fetchAllListMembers(
	ctx context.Context, listAddress string, params common.ReadParams,
) ([]common.ReadResultRow, error) {
	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), "lists/members")
	if err != nil {
		return nil, err
	}

	path = strings.ReplaceAll(path, listAddressPlaceholder, listAddress)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	limit := mustPageSize(params)
	recordsKey := c.arrayFieldName("lists/members")

	url.WithQueryParam(limitParam, strconv.Itoa(limit))

	var (
		allRows []common.ReadResultRow
		skip    int
	)

	for range maxListMemberPages {
		url.WithQueryParam(skipParam, strconv.Itoa(skip))

		resp, err := c.JSONHTTPClient().Get(ctx, url.String())
		if err != nil {
			return nil, err
		}

		result, err := common.ParseResult(
			resp,
			common.ExtractOptionalRecordsFromPath(recordsKey),
			noNextPage,
			common.GetMarshaledData,
			params.Fields,
		)
		if err != nil {
			return nil, err
		}

		allRows = append(allRows, result.Data...)

		// A short page (fewer than the requested limit) is the last page.
		if len(result.Data) < limit {
			return allRows, nil
		}

		skip += len(result.Data)
	}

	return nil, fmt.Errorf("%w: list %s after %d pages",
		errListMemberPagesExceeded, listAddress, maxListMemberPages)
}

// logsRequestBody is the POST /v1/analytics/logs request payload. Start and End
// are omitted when the caller supplies no Since/Until so the API's own defaults
// apply (start: 1 day before now, end: now — per the OpenAPI spec).
type logsRequestBody struct {
	Start      string         `json:"start,omitempty"`
	End        string         `json:"end,omitempty"`
	Pagination logsPagination `json:"pagination"`
}

type logsPagination struct {
	Limit int    `json:"limit"`
	Sort  string `json:"sort,omitempty"`
	Token string `json:"token,omitempty"`
}

// logsTimeLayout matches Mailgun's documented log timestamp format
// (e.g. "Mon, 08 Jul 2024 00:00:00 -0000").
const logsTimeLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

// buildLogsRequest constructs the POST request for the analytics/logs object.
// The pagination cursor (params.NextPage) is carried in the request body.
//
// API reference: https://documentation.mailgun.com/docs/mailgun/api-reference/send/mailgun/logs
//
// Quirk: Mailgun rejects a start time older than the account's log-retention
// window with HTTP 400 ("start time ... is before log retention time ...").
// The retention period is plan-dependent (e.g. ~5 days on trial/sandbox
// accounts) and is not discoverable ahead of time. When the caller supplies no
// Since/Until we therefore omit start/end entirely and let the API's defaults
// (last 1 day) apply — always inside any plan's retention. Callers passing an
// explicit Since must keep it within their account's retention window.
func (c *Connector) buildLogsRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), "analytics/logs")
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	payload := logsRequestBody{
		Pagination: logsPagination{
			Limit: mustPageSize(params),
			Sort:  "@timestamp:asc",
			Token: params.NextPage.String(),
		},
	}

	if !params.Since.IsZero() {
		payload.Start = params.Since.Format(logsTimeLayout)
	}

	if !params.Until.IsZero() {
		payload.End = params.Until.Format(logsTimeLayout)
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// tagsRequestBody is the POST /v1/analytics/tags request payload. The modern
// Tags API (tag group "Tags New") paginates via skip/limit carried in the body;
// limit default is 10, max 1000 per the OpenAPI spec.
type tagsRequestBody struct {
	Pagination tagsPagination `json:"pagination"`
}

type tagsPagination struct {
	Skip  int `json:"skip"`
	Limit int `json:"limit"`
}

// buildTagsRequest constructs the POST request for the analytics/tags object —
// the modern account-scoped Tags API replacing the deprecated GET
// /v3/{domain}/tags. The skip offset (params.NextPage) rides in the body.
func (c *Connector) buildTagsRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), "analytics/tags")
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	payload := tagsRequestBody{
		Pagination: tagsPagination{
			Skip:  nextPageInt(params.NextPage),
			Limit: mustPageSize(params),
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func mustPageSize(params common.ReadParams) int {
	if params.PageSize <= 0 || params.PageSize > maxPageSize {
		return defaultPageSize
	}

	return params.PageSize
}
