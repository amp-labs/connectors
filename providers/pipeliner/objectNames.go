package pipeliner

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/providers/pipeliner/metadata"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = handy.NewSetFromList( //nolint:gochecknoglobals
	metadata.Schemas.GetObjectNames(),
)
