package asana

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

// Some of the objects (allocations, goals, memberships, portfolios, tasks)
// require us to pass either the team ID or the workspace.
// although the API documentation doesn’t explicitly state that these fields are mandatory for fetching data, they are.

const (
	objectNameProjects   = "projects"
	objectNameTags       = "tags"
	objectNameUsers      = "users"
	objectNameWorkspaces = "workspaces"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = datautils.NewSet( //nolint:gochecknoglobals
	objectNameProjects,
	objectNameTags,
	objectNameUsers,
	objectNameWorkspaces,
)

var supportedObjectsByWrite = datautils.NewSet( //nolint:gochecknoglobals
	objectNameProjects,
)

var supportLimitAndOffset = datautils.NewSet( //nolint:gochecknoglobals
	objectNameWorkspaces,
)
