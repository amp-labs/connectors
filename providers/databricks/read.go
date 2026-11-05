package databricks

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

const (
	defaultPageSize = 1000
	maxPageSize     = 10000
)

// pageWindow returns the LIMIT and OFFSET of the requested page.
// NextPage is the decimal offset of the first row of the page.
func pageWindow(params common.ReadParams) (limit, offset int, err error) {
	limit = params.PageSize
	if limit <= 0 {
		limit = defaultPageSize
	}

	if limit > maxPageSize {
		limit = maxPageSize
	}

	if params.NextPage == "" {
		return limit, 0, nil
	}

	offset, err = strconv.Atoi(params.NextPage.String())
	if err != nil || offset < 0 {
		return 0, 0, fmt.Errorf("%w: %q", errInvalidNextPage, params.NextPage)
	}

	return limit, offset, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.Deleted {
		return nil, common.ErrOperationNotSupportedForObject
	}

	statement, err := c.buildReadStatement(params)
	if err != nil {
		return nil, err
	}

	return c.buildStatementHTTPRequest(ctx, statement)
}

func (c *Connector) buildReadStatement(params common.ReadParams) (statementRequest, error) {
	limit, offset, err := pageWindow(params)
	if err != nil {
		return statementRequest{}, err
	}

	hasTimeBound := !params.Since.IsZero() || !params.Until.IsZero()
	if hasTimeBound && c.timestampColumn == "" {
		return statementRequest{}, errTimestampRequired
	}

	var (
		conditions []string
		parameters []statementParameter
	)

	// Since is inclusive so that rows sharing the last seen timestamp are not lost.
	if !params.Since.IsZero() {
		conditions = append(conditions, quoteIdentifier(c.timestampColumn)+" >= :since")
		parameters = append(parameters, timestampParameter("since", params.Since))
	}

	if !params.Until.IsZero() {
		conditions = append(conditions, quoteIdentifier(c.timestampColumn)+" <= :until")
		parameters = append(parameters, timestampParameter("until", params.Until))
	}

	var statement strings.Builder

	statement.WriteString("SELECT ")
	statement.WriteString(c.selectColumns(params))
	statement.WriteString(" FROM ")
	statement.WriteString(c.qualifiedTable(params.ObjectName))

	if len(conditions) != 0 {
		statement.WriteString(" WHERE ")
		statement.WriteString(strings.Join(conditions, " AND "))
	}

	statement.WriteString(" ORDER BY ")
	statement.WriteString(c.orderBy())
	statement.WriteString(fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset))

	return c.newStatementRequest(statement.String(), parameters...), nil
}

// selectColumns lists the requested fields plus the primary key, which identifies each row.
func (c *Connector) selectColumns(params common.ReadParams) string {
	columns := params.Fields.List()

	hasPrimaryKey := false

	for _, column := range columns {
		if strings.EqualFold(column, c.primaryKey) {
			hasPrimaryKey = true

			break
		}
	}

	if !hasPrimaryKey {
		columns = append(columns, c.primaryKey)
	}

	sort.Strings(columns)

	quoted := make([]string, len(columns))
	for index, column := range columns {
		quoted[index] = quoteIdentifier(column)
	}

	return strings.Join(quoted, ", ")
}

func (c *Connector) qualifiedTable(table string) string {
	return quoteIdentifier(c.catalog) + "." + quoteIdentifier(c.schema) + "." + quoteIdentifier(table)
}

func (c *Connector) orderBy() string {
	if c.timestampColumn == "" || strings.EqualFold(c.timestampColumn, c.primaryKey) {
		return quoteIdentifier(c.primaryKey) + " ASC"
	}

	return quoteIdentifier(c.timestampColumn) + " ASC, " + quoteIdentifier(c.primaryKey) + " ASC"
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	_ *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	limit, offset, err := pageWindow(params)
	if err != nil {
		return nil, err
	}

	records, err := parseStatementRecords(resp)
	if err != nil {
		return nil, err
	}

	extractRecords := func(*ajson.Node) ([]map[string]any, error) {
		return records, nil
	}

	nextPage := func(*ajson.Node) (string, error) {
		// A short page is the last one.
		if len(records) < limit {
			return "", nil
		}

		return strconv.Itoa(offset + len(records)), nil
	}

	return common.ParseResult(
		resp,
		extractRecords,
		nextPage,
		c.marshalRecords,
		params.Fields,
	)
}

func (c *Connector) marshalRecords(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
	rows := make([]common.ReadResultRow, len(records))

	for index, record := range records {
		rows[index] = common.ReadResultRow{
			Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
			Raw:    record,
			Id:     c.recordID(record),
		}
	}

	return rows, nil
}

// recordID renders the primary key column of a row as text. Integer keys arrive as numbers.
func (c *Connector) recordID(record map[string]any) string {
	for column, value := range record {
		if !strings.EqualFold(column, c.primaryKey) {
			continue
		}

		switch typed := value.(type) {
		case nil:
			return ""
		case string:
			return typed
		case float64:
			return strconv.FormatFloat(typed, 'f', -1, 64)
		default:
			return fmt.Sprint(typed)
		}
	}

	return ""
}
