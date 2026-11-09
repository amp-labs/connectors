package auphan

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const measureKind = "measure"

type reportColumn struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Kind  string `json:"kind"`
}

type reportTypeResponse struct {
	Data struct {
		Columns []reportColumn `json:"columns"`
	} `json:"data"`
}

type reportResponse struct {
	Data struct {
		Columns []reportColumn `json:"columns"`
		// Cells are kept raw so money and counts reach the caller exactly as sent.
		Rows [][]json.RawMessage `json:"rows"`
	} `json:"data"`
}

// parseReport reads one report window. Auphan has no paging inside a report, so one call returns
// every row of the window; next is the following window, or empty when this was the last one.
func parseReport(response *common.JSONHTTPResponse, fields []string, next string) (*common.ReadResult, error) {
	report, err := common.UnmarshalJSON[reportResponse](response)
	if err != nil {
		return nil, err
	}

	records := mergeReportRows(report.Data.Columns, report.Data.Rows)

	rows, err := common.GetMarshaledData(records, fields)
	if err != nil {
		return nil, err
	}

	return &common.ReadResult{
		Rows:     int64(len(rows)),
		Data:     rows,
		NextPage: common.NextPageToken(next),
		Done:     next == "",
	}, nil
}

// mergeReportRows turns report rows into records keyed by column id.
//
// The "summary" layout sends each group as two rows: a dimension row (every measure null)
// followed by a measure row (every dimension null). Such a pair becomes one record.
// Each row is checked on its own, so a layout that sends whole rows still works.
func mergeReportRows(columns []reportColumn, rows [][]json.RawMessage) []map[string]any {
	records := make([]map[string]any, 0, len(rows))

	for i := 0; i < len(rows); i++ {
		record := rowToRecord(columns, rows[i])

		if i+1 < len(rows) && onlyKind(columns, rows[i], false) && onlyKind(columns, rows[i+1], true) {
			for id, value := range rowToRecord(columns, rows[i+1]) {
				if value != nil {
					record[id] = value
				}
			}

			i++
		}

		records = append(records, record)
	}

	return records
}

// onlyKind reports whether the row has values only in measure columns (measures=true)
// or only in dimension columns (measures=false). Both kinds must be present among the columns.
func onlyKind(columns []reportColumn, row []json.RawMessage, measures bool) bool {
	var hasOwn, hasOther bool

	for i, column := range columns {
		own := (column.Kind == measureKind) == measures
		if own {
			hasOwn = true

			continue
		}

		hasOther = true

		if i < len(row) && !isNull(row[i]) {
			return false
		}
	}

	return hasOwn && hasOther
}

func rowToRecord(columns []reportColumn, row []json.RawMessage) map[string]any {
	record := make(map[string]any, len(columns))

	for i, column := range columns {
		if i < len(row) {
			record[column.ID] = cellValue(row[i])
		} else {
			record[column.ID] = nil
		}
	}

	return record
}

// cellValue keeps strings as strings and numbers as json.Number, so no value goes through float64.
func cellValue(raw json.RawMessage) any {
	if isNull(raw) {
		return nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}

	var value any

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	if err := decoder.Decode(&value); err != nil {
		return string(raw)
	}

	return value
}

func isNull(raw json.RawMessage) bool {
	return len(raw) == 0 || string(raw) == "null"
}

// nextPage reads data.pagination {page, total_pages} from a list response.
func nextPage(node *ajson.Node) (string, error) {
	pagination := jsonquery.New(node, "data", "pagination")

	page, err := pagination.IntegerWithDefault("page", 0)
	if err != nil {
		return "", err
	}

	totalPages, err := pagination.IntegerWithDefault("total_pages", 0)
	if err != nil {
		return "", err
	}

	if page == 0 || page >= totalPages {
		return "", nil
	}

	return strconv.FormatInt(page+1, 10), nil
}
