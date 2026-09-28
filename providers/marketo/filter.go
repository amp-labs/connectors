package marketo

import (
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// RawFilterTypeActivityTypeIDs is the raw filter type Marketo reads accept: comma-separated
// activityTypeIds, required when reading lead activities (e.g. "1,6,12").
const RawFilterTypeActivityTypeIDs = "activityTypeIds"

// ActivityTypeIDsFilter is Marketo's raw filter: the activityTypeIds to read lead activities for.
type ActivityTypeIDsFilter struct {
	// IDs is a comma-separated list of activity type IDs, e.g. "1,6,12".
	IDs string
}

// RawFilterType returns RawFilterTypeActivityTypeIDs.
func (ActivityTypeIDsFilter) RawFilterType() string {
	return RawFilterTypeActivityTypeIDs
}

// FromFilter accepts a non-blank string of comma-separated activity type IDs.
func (ActivityTypeIDsFilter) FromFilter(filter any) (ActivityTypeIDsFilter, error) {
	ids, err := common.StringFilterValue(filter)
	if err != nil {
		return ActivityTypeIDsFilter{}, err
	}

	return ActivityTypeIDsFilter{IDs: ids}, nil
}

var _ connectors.RawFilterConnector = (*Connector)(nil)

// ValidateRawFilter accepts raw filters of type RawFilterTypeActivityTypeIDs.
func (*Connector) ValidateRawFilter(filter common.RawFilter) error {
	_, err := common.RawFilterAs[ActivityTypeIDsFilter](filter)

	return err
}
