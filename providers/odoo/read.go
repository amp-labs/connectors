package odoo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const defaultReadLimit = 500

// odooSearchReadDomainTimeLayout is how Odoo expects datetimes in search_read domains:
// naive string in UTC (no "Z" or offset — Odoo rejects timezone-aware literals).
const odooSearchReadDomainTimeLayout = "2006-01-02 15:04:05"

func formatOdooSearchReadDomainTime(t time.Time) string {
	return t.UTC().Format(odooSearchReadDomainTimeLayout)
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	urlStr, err := c.getURL(params.ObjectName, "search_read")
	if err != nil {
		return nil, err
	}

	limit := params.PageSize
	if limit <= 0 {
		limit = defaultReadLimit
	}

	offset := 0

	if params.NextPage != "" {
		o, convErr := strconv.Atoi(params.NextPage.String())
		if convErr != nil {
			return nil, fmt.Errorf("invalid NextPage (expected numeric offset): %w", convErr)
		}

		offset = o
	}

	body := map[string]any{
		"domain": buildSearchReadDomain(params),
		"fields": params.Fields.List(),
		"limit":  limit,
		"offset": offset,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal search_read body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func buildSearchReadDomain(params common.ReadParams) []any {
	var domain []any

	if !params.Since.IsZero() {
		domain = append(domain, []any{"write_date", ">", formatOdooSearchReadDomainTime(params.Since)})
	}

	if !params.Until.IsZero() {
		domain = append(domain, []any{"write_date", "<=", formatOdooSearchReadDomainTime(params.Until)})
	}

	return domain
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	_ *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	limit := params.PageSize
	if limit <= 0 {
		limit = defaultReadLimit
	}

	offset := 0

	if params.NextPage != "" {
		o, convErr := strconv.Atoi(params.NextPage.String())
		if convErr != nil {
			return nil, fmt.Errorf("invalid NextPage: %w", convErr)
		}

		offset = o
	}

	extractRecords := func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayRequired("")
	}

	nextPage := func(node *ajson.Node) (string, error) {
		return searchReadNextPageOffset(offset, limit, node)
	}

	return common.ParseResult(
		resp,
		extractRecords,
		nextPage,
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
}

// searchReadNextPageOffset returns the next search_read offset as a decimal string when the
// response might have more rows (exactly `limit` rows returned). Otherwise it
// returns an empty string. currentOffset is the offset used for this request (from ReadParams.NextPage).
func searchReadNextPageOffset(currentOffset, limit int, body *ajson.Node) (string, error) {
	records, err := jsonquery.New(body).ArrayRequired("")
	if err != nil {
		return "", err
	}

	n := len(records)
	// check creating a new one in the middle
	if limit <= 0 || n == 0 || n < limit {
		return "", nil
	}

	return strconv.Itoa(currentOffset + n), nil
}
