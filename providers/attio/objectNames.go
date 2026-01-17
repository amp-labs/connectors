package attio

import "github.com/amp-labs/connectors/internal/datautils"

const (
	objectNameLists            = "lists"
	objectNameWorkspacemembers = "workspace_members"
	objectNameTasks            = "tasks"
	objectNameNotes            = "notes"
)

var supportWriteObjects = datautils.NewSet( //nolint:gochecknoglobals
	objectNameLists,
	objectNameTasks,
	objectNameNotes,
)

var supportLimitAndOffset = datautils.NewSet( //nolint:gochecknoglobals
	objectNameTasks,
	objectNameNotes,
)

var supportAttioApi = datautils.NewSet( //nolint:gochecknoglobals
	objectNameLists,
	objectNameWorkspacemembers,
	objectNameTasks,
	objectNameNotes,
)

var readObjectNameToSubscriptionName = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	objectNameLists:            "list",
	objectNameWorkspacemembers: "workspace-member",
	objectNameNotes:            "note",
	objectNameTasks:            "task",
},
	func(objectName string) (fieldName string) {
		return objectName
	},
)
