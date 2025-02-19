package zendesksupport

import "github.com/amp-labs/connectors/providers/zendesksupport/metadata"

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals
