package asana

import "github.com/amp-labs/connectors/internal/datautils"

const (
	objectNameProjects   = "projects"
	objectNameTags       = "tags"
	objectNameUsers      = "users"
	objectNameWorkspaces = "workspaces"
)

var supportedObjectsByRead = datautils.NewSet( //nolint:gochecknoglobals
	objectNameProjects,
	objectNameUsers,
	objectNameWorkspaces,
	objectNameTags,
)

var supportedObjectsByWrite = datautils.NewSet( //nolint:gochecknoglobals
	objectNameProjects,
)

var supportLimitAndOffset = datautils.NewSet( //nolint:gochecknoglobals
	objectNameWorkspaces,
)
