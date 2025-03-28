package attio

import "github.com/amp-labs/connectors/internal/datautils"

const (
	objectNameLists            = "lists"
	objectNameWorkspacemembers = "workspace_members"
	objectNameTasks            = "tasks"
	objectNameNotes            = "notes"
)

var supportedObjectsByWrite = datautils.NewSet( //nolint:gochecknoglobals
	objectNameLists,
	objectNameTasks,
	objectNameNotes,
)

var supportLimitAndOffset = datautils.NewSet( //nolint:gochecknoglobals
	objectNameTasks,
	objectNameNotes,
)

// supportAttioApi represents the APIs listed under the Attio API section in the docs
// (this does not cover the entire Attio API). Reference: https://developers.attio.com/reference.
var supportAttioApi = datautils.NewSet( //nolint:gochecknoglobals
	objectNameLists,
	objectNameWorkspacemembers,
	objectNameTasks,
	objectNameNotes,
)
