package iterable

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/iterable/metadata"
)

const (
	objectNameCatalogs  = "catalogs"
	objectNameJourneys  = "journeys"
	objectNameTemplates = "templates"
)

var paginatedObjects = datautils.NewSet( //nolint:gochecknoglobals
	objectNameCatalogs,
	objectNameJourneys,
)

var incrementalReadObjects = datautils.NewSet( //nolint:gochecknoglobals
	objectNameTemplates,
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals
