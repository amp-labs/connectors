package getresponse

import "github.com/amp-labs/connectors/providers/getresponse/metadata"

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals
