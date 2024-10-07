// nolint
package attio

import "github.com/amp-labs/connectors/common/handy"

const (
	objectNameObjects          = "objects"
	objectNameLists            = "lists"
	objectNameWorkspacemembers = "workspace_members"
	objectNameWebhooks         = "webhooks"
	objectNameTasks            = "tasks"
	objectNameNotes            = "notes"
)

var supportedObjectsByRead = handy.NewSet(
	objectNameObjects,
	objectNameLists,
	objectNameWorkspacemembers,
	objectNameWebhooks,
	objectNameTasks,
	objectNameNotes,
)

var supportedObjectsByWrite = handy.NewSet(
	objectNameObjects,
	objectNameLists,
	objectNameTasks,
	objectNameNotes,
	objectNameWebhooks,
)
