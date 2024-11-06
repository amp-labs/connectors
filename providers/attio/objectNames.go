package attio

import "github.com/amp-labs/connectors/internal/datautils"

const (
	objectNameObjects          = "objects"
	objectNameLists            = "lists"
	objectNameWorkspacemembers = "workspace_members"
	objectNameWebhooks         = "webhooks"
	objectNameTasks            = "tasks"
	objectNameNotes            = "notes"
)

var supportedObjectsByRead = datautils.NewSet( //nolint:gochecknoglobals
	objectNameObjects,
	objectNameLists,
	objectNameWorkspacemembers,
	objectNameWebhooks,
	objectNameTasks,
	objectNameNotes,
)

var supportedObjectsByWrite = datautils.NewSet( //nolint:gochecknoglobals
	objectNameObjects,
	objectNameLists,
	objectNameTasks,
	objectNameNotes,
	objectNameWebhooks,
)

var supportLimitAndOffset = datautils.NewSet( //nolint:gochecknoglobals
	objectNameTasks,
	objectNameNotes,
	objectNameWebhooks,
)
