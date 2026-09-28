package connectors

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// RawFilterValidator checks a raw filter before it is passed to Read or Search.
type RawFilterValidator func(filter common.RawFilter) error

// connectorWrapper is implemented by connectors that wrap another one, such as a metrics wrapper.
type connectorWrapper interface {
	Unwrap() Connector
}

// NewRawFilterValidator returns a RawFilterValidator backed by the connector that will run the filter.
// Wrapped connectors are unwrapped until a RawFilterConnector is found. A connector that doesn't
// implement RawFilterConnector rejects every raw filter. Errors wrap common.ErrInvalidRawFilter.
func NewRawFilterValidator(conn Connector) RawFilterValidator {
	for conn != nil {
		if rawFilterConn, ok := conn.(RawFilterConnector); ok {
			return rawFilterConn.ValidateRawFilter
		}

		wrapper, ok := conn.(connectorWrapper)
		if !ok {
			break
		}

		conn = wrapper.Unwrap()
	}

	return func(common.RawFilter) error {
		return fmt.Errorf("%w: the connector does not support raw filters", common.ErrInvalidRawFilter)
	}
}
