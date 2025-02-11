package google

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/google/metadata"
)

const objectNameCalendarList = "calendarList"

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var supportedObjectsByCreate = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	ModuleCalendar: datautils.NewSet(
		objectNameCalendarList,
	),
}

var supportedObjectsByUpdate = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	ModuleCalendar: datautils.NewSet(
		objectNameCalendarList,
	),
}

var supportedObjectsByDelete = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	ModuleCalendar: datautils.NewSet(
		objectNameCalendarList,
	),
}
