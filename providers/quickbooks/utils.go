package quickbooks

import (
	"fmt"
	"log"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

func inferValueTypeFromData(value any) common.ValueType {
	if value == nil {
		return common.ValueTypeOther
	}

	switch value.(type) {
	case string:
		return common.ValueTypeString
	case float64, int, int64:
		return common.ValueTypeFloat
	case bool:
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}

func buildQuery(params common.ReadParams) string {
	var (
		sinceQuery      string
		paginationQuery string
		untilQuery      string
	)

	query := "SELECT * FROM " + naming.CapitalizeFirstLetter(params.ObjectName)

	if params.NextPage != "" {
		paginationQuery = " STARTPOSITION " + params.NextPage.String() + " MAXRESULTS " + pageSize
	} else {
		paginationQuery = " STARTPOSITION 1 MAXRESULTS " + pageSize
	}

	if !params.Since.IsZero() {
		t := params.Since.Format(time.RFC3339)
		sinceQuery = fmt.Sprintf(" MetaData.LastUpdatedTime >= '%s'", t)
	}

	if !params.Until.IsZero() {
		t := params.Until.Format(time.RFC3339)
		untilQuery = fmt.Sprintf(" MetaData.LastUpdatedTime <= '%s'", t)
	}

	if sinceQuery != "" && untilQuery != "" { //nolint:gocritic
		query += " WHERE " + sinceQuery + " AND " + untilQuery
	} else if sinceQuery != "" {
		query += " WHERE " + sinceQuery
	} else if untilQuery != "" {
		query += " WHERE " + untilQuery
	}

	query += paginationQuery

	log.Printf("Constructed query: %s", query)

	return query
}
