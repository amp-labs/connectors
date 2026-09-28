package klaviyo

import (
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// RawFilterTypeJSONAPI is the raw filter type Klaviyo reads accept: comma-separated
// methods in Klaviyo's JSON:API filtering syntax. https://developers.klaviyo.com/en/docs/filtering_
const RawFilterTypeJSONAPI = "jsonApi"

var _ connectors.RawFilterConnector = (*Connector)(nil)

// ValidateRawFilter accepts raw filters of type RawFilterTypeJSONAPI with a string filter.
func (*Connector) ValidateRawFilter(filter common.RawFilter) error {
	return common.ValidateStringRawFilter(filter, RawFilterTypeJSONAPI)
}
