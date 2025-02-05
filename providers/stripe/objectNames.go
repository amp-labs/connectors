package stripe

import (
	"github.com/amp-labs/connectors/providers/stripe/metadata"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals
