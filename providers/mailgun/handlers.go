package mailgun

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/mailgun/metadata"
)

// errMissingDomain is returned when a domain-scoped object is accessed without
// the connector having a workspace (Mailgun sending domain) configured.
var errMissingDomain = errors.New("mailgun: a domain (workspace) is required to access this object")

const (
	// Mailgun caps page size per endpoint: 1000 for the suppression and domains
	// endpoints, 100 for mailing-list members, templates and dynamic pools. 100
	// is accepted everywhere and is the documented default. We use it as both
	// default and cap so the offset "records < limit ⇒ last page" check can
	// never be fooled by a silent server-side cap below what we requested.
	defaultPageSize = 100
	maxPageSize     = 100

	limitParam = "limit"
	skipParam  = "skip"
)

// domainPlaceholders are the path parameters that all resolve to the connector's
// workspace (the Mailgun sending domain). The Mailgun spec spells the domain
// under several names; schemas.json keeps them verbatim and we substitute here.
//
//nolint:gochecknoglobals
var domainPlaceholders = []string{"{domain_name}", "{authority_name}"}

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
	// The spec spells the page-size parameter "Limit", but the live API parses
	// lowercase "limit": its paging cursor encodes the size, and only "limit=1"
	// yields {"s":1} (verified 2026-09-15), so the common limit param is used.
	"dynamic_pools/history": {paginationCursor, scopeAccount, "timestamp"},

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
	"thresholds/hits":                   {paginationNone, scopeAccount, "created_at"},
	// Alert settings records sit under "events" beside the webhook/slack channel
	// config in the same envelope; responseKey selects the array.
	"alerts/settings": {paginationNone, scopeAccount, ""},

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

	path, err = c.substituteDomain(path)
	if err != nil {
		return nil, err
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

// substituteDomain fills every domain placeholder in path with the workspace. It
// is a no-op for paths without a domain placeholder, so callers apply it
// unconditionally; the workspace is required only when a placeholder is present.
func (c *Connector) substituteDomain(path string) (string, error) {
	if !pathHasDomainPlaceholder(path) {
		return path, nil
	}

	if c.workspace == "" {
		return "", errMissingDomain
	}

	for _, placeholder := range domainPlaceholders {
		path = strings.ReplaceAll(path, placeholder, c.workspace)
	}

	return path, nil
}

func pathHasDomainPlaceholder(path string) bool {
	for _, placeholder := range domainPlaceholders {
		if strings.Contains(path, placeholder) {
			return true
		}
	}

	return false
}

func pageSize(params common.ReadParams) string {
	return strconv.Itoa(mustPageSize(params))
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
			readhelper.MakeMarshaledDataFuncWithId(nil, readIDField(params.ObjectName)),
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
		readhelper.MakeGetMarshaledDataWithId(readIDField(params.ObjectName)),
		params.Fields,
	)
}

// readIDFields names, per object, the record field that identifies a record,
// populated into ReadResultRow.Id. Mailgun has no uniform "id": suppressions,
// mailing lists and members are keyed by address, allowlist entries by value,
// templates by name, webhooks by webhook_id, and so on. The keys match the
// RecordId that Write returns and Delete expects for the same object. Objects
// absent here use "id"; those without a stable per-record identifier
// (dkim/keys, domains/keys) therefore leave Id empty.
//
//nolint:gochecknoglobals
var readIDFields = map[string]string{
	"bounces":                           "address",
	"complaints":                        "address",
	"unsubscribes":                      "address",
	"lists":                             "address",
	"lists/members":                     "address",
	"whitelists":                        "value",
	"templates":                         "name",
	"account/templates":                 "name",
	"webhooks":                          "webhook_id",
	"ip_pools":                          "pool_id",
	"accounts/subaccounts/ip_pools/all": "pool_id",
	"ips":                               "ip",
	"ip_whitelist":                      "ip_address",
	"domains/credentials":               "login",
	"analytics/tags":                    "tag",
}

func readIDField(objectName string) readhelper.IdFieldQuery {
	if field, ok := readIDFields[objectName]; ok {
		return readhelper.NewIdField(field)
	}

	return readhelper.NewIdField("id")
}

func (c *Connector) arrayFieldName(objectName string) string {
	return metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), objectName)
}

func mustPageSize(params common.ReadParams) int {
	return min(readhelper.PageSizeWithDefault(params, defaultPageSize), maxPageSize)
}
