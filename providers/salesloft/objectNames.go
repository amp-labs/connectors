package salesloft

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/providers/salesloft/metadata"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = handy.NewSet( //nolint:gochecknoglobals
	metadata.Schemas.GetObjectNames(),
)
