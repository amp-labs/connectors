package connectors

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// connectorWrapper is implemented by connectors that wrap another one, such as a metrics wrapper.
type connectorWrapper interface {
	Unwrap() Connector
}

// ValidateRawFilter checks a raw filter with the connector that will run it, before the filter is
// passed to Read or Search. Wrapped connectors are unwrapped until a RawFilterConnector is found.
// A connector that doesn't implement RawFilterConnector rejects every raw filter.
// Errors wrap common.ErrInvalidRawFilter.
func ValidateRawFilter(conn Connector, filter common.RawFilter) error {
	for conn != nil {
		if rawFilterConn, ok := conn.(RawFilterConnector); ok {
			return rawFilterConn.ValidateRawFilter(filter)
		}

		wrapper, ok := conn.(connectorWrapper)
		if !ok {
			break
		}

		conn = wrapper.Unwrap()
	}

	return fmt.Errorf("%w: the connector does not support raw filters", common.ErrInvalidRawFilter)
}
