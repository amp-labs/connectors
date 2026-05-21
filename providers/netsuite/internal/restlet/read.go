package restlet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	defaultPageSize = 1000
)

func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	pageIndex := 0

	if len(params.NextPage) != 0 {
		idx, err := strconv.Atoi(params.NextPage.String())
		if err != nil {
			return nil, fmt.Errorf("invalid nextPage token: %w", err)
		}

		pageIndex = idx
	}

	pageSize := defaultPageSize
	if params.PageSize > 0 {
		pageSize = params.PageSize
	}

	payload := searchRequest{
		Action:    "search",
		Type:      params.ObjectName,
		Columns:   params.Fields.List(),
		Filters:   buildDateFilters(params),
		PageSize:  pageSize,
		PageIndex: pageIndex,
		Limit:     pageSize,
		Sort:      []sortSpec{{Column: "internalid", Direction: "ASC"}},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.restletURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// buildDateFilters constructs NetSuite Since/Until filter expressions.
// NetSuite requires explicit "AND" between multiple filter expressions.
func buildDateFilters(params common.ReadParams) []any {
	var filters []any

	if !params.Since.IsZero() {
		filters = append(filters, []string{
			"lastmodifieddate", "onorafter", params.Since.UTC().Format(time.RFC3339),
		})
	}

	if !params.Until.IsZero() {
		if len(filters) > 0 {
			filters = append(filters, "AND")
		}

		filters = append(filters, []string{
			"lastmodifieddate", "onorbefore", params.Until.UTC().Format(time.RFC3339),
		})
	}

	return filters
}

func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return parseSearchResults(resp)
}

// parseSearchResults parses a RESTlet search response envelope into a ReadResult.
// Shared by both Read and Search since they use the same RESTlet search action and response format.
func parseSearchResults(resp *common.JSONHTTPResponse) (*common.ReadResult, error) {
	fullResp, err := common.UnmarshalJSON[restletResponse](resp)
	if err != nil {
		return nil, err
	}

	if fullResp.Header.Status != statusSuccess {
		return nil, parseRestletError(fullResp)
	}

	var records []map[string]json.RawMessage
	if err := json.Unmarshal(fullResp.Body, &records); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	rows := make([]common.ReadResultRow, 0, len(records))
	for _, rec := range records {
		rows = append(rows, parseSearchRecord(rec))
	}

	result := &common.ReadResult{
		Rows: fullResp.Header.TotalResults,
		Data: rows,
		Done: !fullResp.Header.HasMore,
	}

	if fullResp.Header.NextPage != nil {
		result.NextPage = common.NextPageToken(strconv.Itoa(*fullResp.Header.NextPage))
	}

	return result, nil
}

// parseSearchRecord converts a single raw RESTlet record into a ReadResultRow.
func parseSearchRecord(rec map[string]json.RawMessage) common.ReadResultRow {
	row := common.ReadResultRow{
		Fields: make(map[string]any),
		Raw:    make(map[string]any),
	}

	if idRaw, ok := rec["_id"]; ok {
		if node, err := ajson.Unmarshal(idRaw); err == nil {
			if id, err := jsonquery.New(node).TextWithDefault("", ""); err == nil {
				row.Id = id
			}
		}
	}

	for colName, colRaw := range rec {
		if colName == "_id" || colName == "_type" {
			continue
		}

		var fieldVal searchFieldValue
		if err := json.Unmarshal(colRaw, &fieldVal); err != nil {
			// Not a {value,text} pair — store raw.
			var raw any
			if err := json.Unmarshal(colRaw, &raw); err == nil {
				row.Fields[strings.ToLower(colName)] = raw
				row.Raw[colName] = raw
			}

			continue
		}

		row.Fields[strings.ToLower(colName)] = fieldVal.Value
		row.Raw[colName] = map[string]any{"value": fieldVal.Value, "text": fieldVal.Text}
	}

	return row
}
