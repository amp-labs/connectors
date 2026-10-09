package jump

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/graphql"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	// Jump rejects any query whose complexity exceeds this limit.
	// https://my.jumpapp.com/enterprise/documentation (Query Complexity)
	maxQueryComplexity = 1000

	maxPageSize = 100
)

//go:embed graphql/*.graphql
var queryFiles embed.FS

type QueryParameters struct {
	First         int
	After         string
	InsertedAfter string
	UpdatedAfter  string
	UpdatedBefore string
	Fields        map[string]bool
}

// buildReadRequest renders the object's query template with the page size, page token and time window.
// https://my.jumpapp.com/enterprise/documentation (Pagination)
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	queryParams := QueryParameters{
		First:  pageSize(params),
		Fields: queryTemplateFields(params.ObjectName, params.Fields),
	}

	if params.NextPage != "" {
		if err := setNextPage(&queryParams, params); err != nil {
			return nil, err
		}
	}

	// Jump compares timestamps with microsecond precision, so the fraction of a second must be kept.
	// Dropping it would exclude records updated within the last second before Until.
	if !params.Since.IsZero() {
		queryParams.UpdatedAfter = params.Since.UTC().Format(time.RFC3339Nano)
	}

	if !params.Until.IsZero() {
		queryParams.UpdatedBefore = params.Until.UTC().Format(time.RFC3339Nano)
	}

	return c.buildGraphQLRequest(ctx, params.ObjectName, queryParams)
}

// queryTemplateFields marks the requested fields for the query template.
// Nested relations are matched case-insensitively, so templates can use the exact schema names.
func queryTemplateFields(objectName string, requested datautils.StringSet) map[string]bool {
	fields := make(map[string]bool, len(requested))
	for field := range requested {
		fields[field] = true
	}

	for canonical := range optionalFieldComplexity[objectName] {
		if isFieldRequested(requested, canonical) {
			fields[canonical] = true
		}
	}

	return fields
}

func isFieldRequested(requested datautils.StringSet, field string) bool {
	for name := range requested {
		if strings.EqualFold(name, field) {
			return true
		}
	}

	return false
}

func setNextPage(queryParams *QueryParameters, params common.ReadParams) error {
	if params.ObjectName != contactsObjectName {
		queryParams.After = params.NextPage.String()

		return nil
	}

	token, err := newContactsPageToken(params.NextPage)
	if err != nil {
		return err
	}

	queryParams.InsertedAfter = token.InsertedAfter
	// The inclusive filter returns the previous page's last record again, so one more record is requested
	// when the complexity limit allows it. Otherwise that record takes one slot of the page.
	queryParams.First = min(queryParams.First+1, maxPageSizeWithinLimit(params))

	return nil
}

func (c *Connector) buildGraphQLRequest(
	ctx context.Context,
	objectName string,
	queryParams QueryParameters,
) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, err
	}

	query, err := graphql.Operation(queryFiles, "query", objectName, queryParams)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", common.ErrOperationNotSupportedForObject, objectName)
	}

	if err != nil {
		return nil, err
	}

	requestBody := map[string]string{
		"query": query,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	if err := interpretGraphQLError(resp, params.ObjectName); err != nil {
		return nil, err
	}

	// Records are located at "data.<objectName>.items", see the response example below.
	if params.ObjectName == contactsObjectName {
		return common.ParseResultFiltered(
			params,
			resp,
			common.MakeRecordsFunc("items", "data", params.ObjectName),
			filterContacts,
			readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
			params.Fields,
		)
	}

	return common.ParseResult(
		resp,
		common.ExtractOptionalRecordsFromPath("items", "data", params.ObjectName),
		makeNextRecordsURL(params.ObjectName),
		readhelper.MakeGetMarshaledDataWithId(readhelper.NewIdField("id")),
		params.Fields,
	)
}

// makeNextRecordsURL returns the endCursor while hasNextPage is true.
// Cursors are opaque and passed back unchanged as the "after" argument.
//
//	{
//	  "data": {
//	    "meetings": {
//	      "items": [...],
//	      "pageInfo": {
//	        "hasNextPage": true,
//	        "endCursor": "<opaque cursor>"
//	      }
//	    }
//	  }
//	}
func makeNextRecordsURL(objectName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pageInfo := jsonquery.New(node, "data", objectName, "pageInfo")

		hasNextPage, err := pageInfo.BoolWithDefault("hasNextPage", false)
		if err != nil || !hasNextPage {
			return "", err
		}

		return pageInfo.StrWithDefault("endCursor", "")
	}
}

// pageSize returns the largest page size that keeps the list query within the complexity limit.
// A smaller page size requested by the caller is honoured.
func pageSize(params common.ReadParams) int {
	size := maxPageSizeWithinLimit(params)
	if params.PageSize > 0 {
		size = min(size, params.PageSize)
	}

	return max(size, 1)
}

func maxPageSizeWithinLimit(params common.ReadParams) int {
	return min(maxPageSize, maxQueryComplexity/recordComplexity(params))
}

// recordComplexity is the complexity of a single record for the requested fields.
func recordComplexity(params common.ReadParams) int {
	cost := max(objectComplexity[params.ObjectName], 1)

	for field, fieldCost := range optionalFieldComplexity[params.ObjectName] {
		if isFieldRequested(params.Fields, field) {
			cost += fieldCost
		}
	}

	return cost
}

type contactsPageToken struct {
	InsertedAfter string `json:"insertedAfter"`
	LastID        string `json:"lastId"`
}

func newContactsPageToken(token common.NextPageToken) (*contactsPageToken, error) {
	var pageToken contactsPageToken
	if err := json.Unmarshal([]byte(token), &pageToken); err != nil {
		return nil, fmt.Errorf("%w: %w", common.ErrNextPageInvalid, err)
	}

	if _, err := time.Parse(time.RFC3339Nano, pageToken.InsertedAfter); err != nil {
		return nil, fmt.Errorf("%w: %w", common.ErrNextPageInvalid, err)
	}

	return &pageToken, nil
}

func (t contactsPageToken) String() (string, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// seen reports whether a record was returned on an earlier page.
// The previous page's last record is always returned again because "insertedAfter" is inclusive.
func (t contactsPageToken) seen(insertedAt time.Time, id string) bool {
	boundary, err := time.Parse(time.RFC3339Nano, t.InsertedAfter)
	if err != nil {
		return false
	}

	return insertedAt.Before(boundary) || (insertedAt.Equal(boundary) && id <= t.LastID)
}

// filterContacts drops records already returned on earlier pages and produces the next page token
// from the last record not returned before.
func filterContacts(
	params common.ReadParams, body *ajson.Node, records []*ajson.Node,
) ([]*ajson.Node, string, error) {
	var previous *contactsPageToken

	if params.NextPage != "" {
		token, err := newContactsPageToken(params.NextPage)
		if err != nil {
			return nil, "", err
		}

		previous = token
	}

	filtered, err := dropSeenContacts(records, previous)
	if err != nil {
		return nil, "", err
	}

	hasNextPage, err := jsonquery.New(body, "data", contactsObjectName, "pageInfo").BoolWithDefault("hasNextPage", false)
	if err != nil {
		return nil, "", err
	}

	if !hasNextPage {
		return filtered, "", nil
	}

	// More contacts than fit on a page share the same insertedAt, so the window cannot advance.
	if len(filtered) == 0 {
		return nil, "", fmt.Errorf("%w: contacts pagination cannot advance", common.ErrNextPageInvalid)
	}

	next, err := nextContactsPageToken(filtered[len(filtered)-1])
	if err != nil {
		return nil, "", err
	}

	return filtered, next, nil
}

func dropSeenContacts(records []*ajson.Node, previous *contactsPageToken) ([]*ajson.Node, error) {
	if previous == nil {
		return records, nil
	}

	filtered := make([]*ajson.Node, 0, len(records))

	for _, record := range records {
		insertedAt, id, err := contactPosition(record)
		if err != nil {
			return nil, err
		}

		if !previous.seen(insertedAt, id) {
			filtered = append(filtered, record)
		}
	}

	return filtered, nil
}

func nextContactsPageToken(lastRecord *ajson.Node) (string, error) {
	insertedAt, err := jsonquery.New(lastRecord).StringRequired("insertedAt")
	if err != nil {
		return "", err
	}

	id, err := jsonquery.New(lastRecord).StringRequired("id")
	if err != nil {
		return "", err
	}

	return contactsPageToken{
		InsertedAfter: insertedAt,
		LastID:        id,
	}.String()
}

func contactPosition(record *ajson.Node) (time.Time, string, error) {
	insertedAtText, err := jsonquery.New(record).StringRequired("insertedAt")
	if err != nil {
		return time.Time{}, "", err
	}

	insertedAt, err := time.Parse(time.RFC3339Nano, insertedAtText)
	if err != nil {
		return time.Time{}, "", err
	}

	id, err := jsonquery.New(record).StringRequired("id")
	if err != nil {
		return time.Time{}, "", err
	}

	return insertedAt, id, nil
}
