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

var supportAttioGeneralApi = datautils.NewSet( //nolint:gochecknoglobals
	objectNameLists,
	objectNameWorkspacemembers,
	objectNameTasks,
	objectNameNotes,
)
