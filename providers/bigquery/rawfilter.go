package bigquery

import (
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// RawFilterTypeSQL is the raw filter type BigQuery reads accept: a row restriction
// (a SQL boolean expression), joined with AND to the read's time window.
const RawFilterTypeSQL = "sql"

var _ connectors.RawFilterConnector = (*Connector)(nil)

// ValidateRawFilter accepts raw filters of type RawFilterTypeSQL with a string filter.
func (*Connector) ValidateRawFilter(filter common.RawFilter) error {
	return common.ValidateStringRawFilter(filter, RawFilterTypeSQL)
}
