package getresponse

import (
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// RawFilterTypeQueryParams is the raw filter type GetResponse reads accept: '&'-separated
// query[...] and sort[...] parameters, e.g. "query[name]=test&sort[createdOn]=DESC".
const RawFilterTypeQueryParams = "queryParams"

var _ connectors.RawFilterConnector = (*Connector)(nil)

// ValidateRawFilter accepts raw filters of type RawFilterTypeQueryParams with a string filter.
func (*Connector) ValidateRawFilter(filter common.RawFilter) error {
	return common.ValidateStringRawFilter(filter, RawFilterTypeQueryParams)
}
