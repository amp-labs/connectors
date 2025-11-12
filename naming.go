package connectors

import (
	"context"
	"errors"
)

// ErrNilConnector is returned when a nil connector is passed to a function that requires a valid connector.
var ErrNilConnector = errors.New("nil connector")

// GetEntityName returns the normalized entity name for the given connector and entity type.
// If the connector implements EntityNamingConnector, it will use the provider-specific
// normalization rules. Otherwise, it returns the input value unchanged.
//
// This is a convenience function that allows callers to normalize entity names without
// explicitly checking if the connector supports entity naming normalization.
//
// Parameters:
//   - ctx: Context for the operation
//   - conn: The connector to use for normalization (must not be nil)
//   - entity: The type of entity (EntityObject or EntityField)
//   - value: The entity name to normalize
//
// Returns:
//   - The normalized entity name, or the original value if normalization is not supported
//   - ErrNilConnector if conn is nil
//   - Any error returned by the connector's NormalizeEntityName method
func GetEntityName(ctx context.Context, conn Connector, entity Entity, value string) (string, error) {
	if conn == nil {
		return "", ErrNilConnector
	}

	enc, ok := conn.(EntityNamingConnector)
	if ok {
		return enc.NormalizeEntityName(ctx, entity, value)
	}

	return value, nil
}
