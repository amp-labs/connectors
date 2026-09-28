package marketo

import (
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// RawFilterTypeActivityTypeIDs is the raw filter type Marketo reads accept: comma-separated
// activityTypeIds, required when reading lead activities (e.g. "1,6,12").
const RawFilterTypeActivityTypeIDs = "activityTypeIds"

var _ connectors.RawFilterConnector = (*Connector)(nil)

// ValidateRawFilter accepts raw filters of type RawFilterTypeActivityTypeIDs with a string filter.
func (*Connector) ValidateRawFilter(filter common.RawFilter) error {
	return common.ValidateStringRawFilter(filter, RawFilterTypeActivityTypeIDs)
}
