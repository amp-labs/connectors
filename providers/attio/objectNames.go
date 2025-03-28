package attio

import "github.com/amp-labs/connectors/internal/datautils"

const (
	objectNameLists            = "lists"
	objectNameWorkspacemembers = "workspace_members"
	objectNameTasks            = "tasks"
	objectNameNotes            = "notes"
)

// supportWriteObjects represents the APIs listed under the Attio API section in the docs
// (this does not cover the entire Attio API). Reference: https://developers.attio.com/reference.
var supportWriteObjects = datautils.NewSet( //nolint:gochecknoglobals
	objectNameLists,
	objectNameTasks,
	objectNameNotes,
)

var supportLimitAndOffset = datautils.NewSet( //nolint:gochecknoglobals
	objectNameTasks,
	objectNameNotes,
)

var supportAttioGeneralApi = datautils.NewSet( //nolint:gochecknoglobals
	objectNameLists,
	objectNameWorkspacemembers,
	objectNameTasks,
	objectNameNotes,
)
