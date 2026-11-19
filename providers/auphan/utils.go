package auphan

import (
	"encoding/json"
	"strconv"

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

// extractReportRecords cuts a report table into records. A report response is a table:
//
//	{"data": {"columns": [{"id": "payment_type", "kind": "dimension"}, ...], "rows": [["MC", null], ...]}}
//
// Each record keeps the table's columns and the original rows it is made of, so its Raw is the
// table as Auphan sent it: {"columns": [...], "rows": [[...], ...]}. mergeReportRecord then maps
// those rows to one {column id: value} object for Fields.
func extractReportRecords(body *ajson.Node) ([]*ajson.Node, error) {
	columnNodes, err := jsonquery.New(body, "data").ArrayRequired("columns")
	if err != nil {
		return nil, err
	}

	rowNodes, err := jsonquery.New(body, "data").ArrayOptional("rows")
	if err != nil {
		return nil, err
	}

	columns := make([]reportColumn, len(columnNodes))
	for i, node := range columnNodes {
		if err := json.Unmarshal(node.Source(), &columns[i]); err != nil {
			return nil, err
		}
	}

	rows := make([][]*ajson.Node, len(rowNodes))
	for i, node := range rowNodes {
		if rows[i], err = node.GetArray(); err != nil {
			return nil, err
		}
	}

	groups := groupReportRows(columns, rows)
	records := make([]*ajson.Node, len(groups))

	for i, group := range groups {
		groupRows := make([]*ajson.Node, len(group))
		for j, row := range group {
			groupRows[j] = rowNodes[row].Clone()
		}

		records[i] = ajson.ObjectNode("", map[string]*ajson.Node{
			"columns": ajson.ArrayNode("", cloneNodes(columnNodes)),
			"rows":    ajson.ArrayNode("", groupRows),
		})
	}

	return records, nil
}

// groupReportRows returns, for each record, the indexes of the rows it is made of.
//
// The "summary" layout sends each group as a dimension row (every measure null) followed by
// one or more measure rows (every dimension null). Usually there is one measure row, but some
// columns split a group into detail rows (tax with tax_date sends one measure row per invoice).
// Each measure row becomes one record together with its group's dimension row.
// Any other row is a record on its own, so a layout that sends whole rows still works.
func groupReportRows(columns []reportColumn, rows [][]*ajson.Node) [][]int {
	groups := make([][]int, 0, len(rows))

	for row := 0; row < len(rows); row++ {
		isDimensionRow := isHomogenous(columns, rows[row], false)
		if !isDimensionRow {
			groups = append(groups, []int{row})

			continue
		}

		dimensionRow := row
		for row+1 < len(rows) && isHomogenous(columns, rows[row+1], true) {
			row++
			groups = append(groups, []int{dimensionRow, row})
		}

		// A dimension row with no measure rows under it.
		if row == dimensionRow {
			groups = append(groups, []int{dimensionRow})
		}
	}

	return groups
}

// isHomogenous reports whether the row has values only in measure columns (measures=true)
// or only in dimension columns (measures=false). Rows that mix both kinds are not homogenous.
func isHomogenous(columns []reportColumn, row []*ajson.Node, measures bool) bool {
	for i, column := range columns {
		isMeasure := column.Kind == measureKind
		if isMeasure != measures && i < len(row) && !row[i].IsNull() {
			return false
		}
	}

	return true
}

// mergeReportRecord maps a record from extractReportRecords to {column id: value}.
// A value in a later row (the measures) fills in what the dimension row left null.
func mergeReportRecord(node *ajson.Node) (map[string]any, error) {
	columns, err := jsonquery.New(node).ArrayRequired("columns")
	if err != nil {
		return nil, err
	}

	rows, err := jsonquery.New(node).ArrayRequired("rows")
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(columns))
	record := make(map[string]any, len(columns))

	for i, column := range columns {
		if ids[i], err = jsonquery.New(column).StringRequired("id"); err != nil {
			return nil, err
		}

		record[ids[i]] = nil
	}

	if err := fillRecord(record, ids, rows); err != nil {
		return nil, err
	}

	return record, nil
}

// fillRecord sets record[ids[i]] to every non-null value in column i of the rows.
func fillRecord(record map[string]any, ids []string, rows []*ajson.Node) error {
	for _, row := range rows {
		cells, err := row.GetArray()
		if err != nil {
			return err
		}

		for i, cell := range cells {
			if i >= len(ids) || cell.IsNull() {
				continue
			}

			if record[ids[i]], err = cellValue(cell); err != nil {
				return err
			}
		}
	}

	return nil
}

// cellValue keeps numbers as json.Number, so no value goes through float64.
func cellValue(cell *ajson.Node) (any, error) {
	if cell.IsNumeric() {
		return json.Number(cell.Source()), nil
	}

	return cell.Unpack()
}

func cloneNodes(nodes []*ajson.Node) []*ajson.Node {
	clones := make([]*ajson.Node, len(nodes))
	for i, node := range nodes {
		clones[i] = node.Clone()
	}

	return clones
}

// reportNextPage ends the read after one call: Auphan has no paging inside a report.
func reportNextPage(*ajson.Node) (string, error) {
	return "", nil
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
