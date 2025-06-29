package google

import "github.com/amp-labs/connectors/providers/google/metadata"

const objectNameCalendarList = "calendarList"

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals
