package pipeliner

import "github.com/amp-labs/connectors/providers/pipeliner/metadata"

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals
