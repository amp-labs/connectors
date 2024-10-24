package salesloft

import "github.com/amp-labs/connectors/providers/salesloft/metadata"

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals
