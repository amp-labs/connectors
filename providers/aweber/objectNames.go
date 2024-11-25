package aweber

import "github.com/amp-labs/connectors/providers/aweber/metadata"

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals
